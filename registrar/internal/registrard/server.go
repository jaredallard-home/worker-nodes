// Code generated by Bootstrap.
//
// Please edit this to more accurately match the server implementation.

package registrard

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/google/uuid"
	"github.com/jaredallard-home/worker-nodes/registrar/api"
	"github.com/jaredallard-home/worker-nodes/registrar/apis/clientset/v1alpha1"
	registrar "github.com/jaredallard-home/worker-nodes/registrar/apis/types/v1alpha1"
	"github.com/jaredallard-home/worker-nodes/registrar/internal/kube"
	"github.com/jaredallard-home/worker-nodes/registrar/pkg/rancher"
	"github.com/jaredallard-home/worker-nodes/registrar/pkg/wghelper"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Ensure that we implemented the interface compile time
var (
	_ api.Service = &Server{}
)

// Server is the actual server implementation of the API.
type Server struct {
	k            *v1alpha1.RegistrarClientset
	w            *wghelper.Wireguard
	r            *rancher.Client
	authToken    []byte
	authTokenlen int32
}

// NewServer creates a new grpc server interface
func NewServer(ctx context.Context) (*Server, error) {
	s := &Server{}
	c, err := kube.New()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kube config")
	}

	s.r = rancher.NewClient(os.Getenv("RANCHER_TOKEN"))

	s.k, err = v1alpha1.NewForConfig(c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kubernetes and registrar clientset")
	}

	s.w, err = wghelper.NewWireguard(s.k)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create wireguard controller")
	}

	_, pool, err := s.getCIDR(ctx, "registrard")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get CIDR block")
	}

	err = s.w.StartServer(pool)

	if err := s.w.Flush(ctx); err != nil {
		log.Errorf("Failed to flush wireguard, issues may occur: %v", err)
	}

	// TODO(jaredallard): default namespace hardcode
	// TODO(jaredallard): move this out and support pagination
	// download all of the wireguard peers and insert them into our devices
	ips, err := s.k.RegistrarV1Alpha1Client().WireguardIPs("registrard").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Warnf("failed to fetch peers from kubernetes: %v", err)
	}
	for _, ip := range ips.Items {
		_, err := s.w.Register(ctx, &ip)
		if err != nil {
			log.Errorf("failed to add peer: %v", err)
		} else {
			log.WithFields(log.Fields{"ip": ip.Spec.IPAdress, "device": ip.Spec.DeviceRef}).
				Infof("added peer")
		}
	}

	s.authToken = []byte(os.Getenv("REGISTRARD_TOKEN"))
	s.authTokenlen = int32(len(s.authToken))
	return s, err
}

func (s *Server) getCIDR(ctx context.Context, namespace string) (*net.IPNet, *registrar.WireguardIPPool, error) {
	pools, err := s.k.RegistrarV1Alpha1Client().WireguardIPPools(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get wireguardippools")
	}

	if len(pools.Items) == 0 {
		return nil, nil, fmt.Errorf("no active pools, retry later")
	}

	if len(pools.Items) > 1 {
		log.Warnf("found %d pools, only one is supported, using first pool", len(pools.Items))
	}

	pool := pools.Items[0]
	_, ipnet, err := net.ParseCIDR(pool.Spec.CIDR)
	if err != nil {
		return nil, nil, errors.Wrap(err, "invalid cidr")
	}

	return ipnet, &pools.Items[0], nil
}

func (s *Server) allocateIP(ctx context.Context, namespace string) (net.IP, *registrar.WireguardIPPool, error) {
	ipnet, pool, err := s.getCIDR(ctx, namespace)

	// TODO(jaredallard): we should not do a blind +1/2, we should recycle unused IP addresses
	// we do +2 here to account for the host most likely being set to .1
	// of course, this will break if it's not set to that. Oh well. DHCP addressing is hard.
	ip, err := cidr.Host(ipnet, pool.Status.UsedAddresses+2)

	// we increment the amount, since we're now using another
	pool.Status.UsedAddresses++

	_, err = s.k.RegistrarV1Alpha1Client().WireguardIPPools(pool.Namespace).Update(ctx, pool)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to update usedAddresses on ip pool")
	}

	return ip, pool, err
}

func (s *Server) createDevice(ctx context.Context, namespace string, r *api.RegisterRequest) error {
	ip, pool, err := s.allocateIP(ctx, namespace)
	if err != nil {
		return errors.Wrap(err, "failed to allocate IP address")
	}

	wgip := &registrar.WireguardIP{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ReplaceAll(ip.String(), ".", "-"),
		},
		Spec: registrar.WireguardIPSpec{
			DeviceRef: r.Id,
			PoolRef:   pool.ObjectMeta.Name,
			IPAdress:  ip.String(),
		},
		Status: registrar.WireguardIPStatus{
			Active: true,
		},
	}

	// register in wireguard
	conf, err := s.w.Register(ctx, wgip)
	if err != nil {
		return errors.Wrap(err, "failed to add device to wireguard")
	}

	// create the device secret
	_, err = s.k.CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Id,
		},
		StringData: map[string]string{
			"wireguard-key": conf.PresharedKey.String(),
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to create device secret")
	}

	// device doesn't exist, create it
	_, err = s.k.RegistrarV1Alpha1Client().Devices(namespace).Create(ctx, &registrar.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Id,
		},
		Spec: registrar.DeviceSpec{
			SecretRef:      r.Id,
			WireguardIPRef: wgip.ObjectMeta.Name,
		},
		Status: registrar.DeviceStatus{
			Registered: true,
			PublicKey:  conf.PresharedKey.PublicKey().String(),
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to create device")
	}

	_, err = s.k.RegistrarV1Alpha1Client().WireguardIPs(namespace).Create(ctx, wgip, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to create wireguard ip")
	}

	return nil
}

// Register registers a new device into the wireguard network.
// TODO(jaredallard): GC when peer is not added fully
func (s *Server) Register(ctx context.Context, r *api.RegisterRequest) (*api.RegisterResponse, error) {
	namespace := "registrard"
	userTokenByte := []byte(r.AuthToken)

	// we need to check if the auth token is the correct length
	if subtle.ConstantTimeEq(s.authTokenlen, int32(len(userTokenByte))) == 0 {
		return nil, fmt.Errorf("invalid auth token")
	}

	// we need to check if the token is actually valid
	if subtle.ConstantTimeCompare(s.authToken, userTokenByte) == 0 {
		return nil, fmt.Errorf("invalid auth token")
	}

	if r.Id == "" {
		// generate a new UUID for this device
		r.Id = uuid.New().String()
	}

	log.Infof("attempting to register device '%s'", r.Id)
	resp := &api.RegisterResponse{
		Id: r.Id,
	}

	d, err := s.k.RegistrarV1Alpha1Client().Devices(namespace).Get(ctx, r.Id, metav1.GetOptions{})
	if err == nil {
		log.Infof("device '%s' already exists, returning registration information ...", r.Id)
	} else if kerrors.IsNotFound(err) {
		log.Infof("device '%s' is new, registering ...", r.Id)
		if err := s.createDevice(ctx, namespace, r); err != nil {
			return nil, errors.Wrap(err, "failed to register device")
		}
	} else if err != nil {
		// we checked all errors we handle, just return it
		return nil, err
	}

	d, err = s.k.RegistrarV1Alpha1Client().Devices(namespace).Get(ctx, r.Id, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get device")
	}

	sec, err := s.k.CoreV1().Secrets(namespace).Get(ctx, d.Spec.SecretRef, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get device secret")
	}

	wgip, err := s.k.RegistrarV1Alpha1Client().WireguardIPs(namespace).
		Get(ctx, d.Spec.WireguardIPRef, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get device ip address")
	}

	wgp, err := s.k.RegistrarV1Alpha1Client().WireguardIPPools(namespace).
		Get(ctx, wgip.Spec.PoolRef, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get device ip pool")
	}

	// TODO(jaredallard): fix this hardcoded ID
	tr, err := s.r.GetClusterRegistrationToken(ctx, "c-ptjcq")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get rancher token")
	}

	// TODO(jaredallard): we should create one then
	if len(tr) == 0 {
		return nil, fmt.Errorf("no cluster registration tokens available for specified rancher cluster")
	}

	resp.Key = string(sec.Data["wireguard-key"])
	resp.IpAddress = wgip.Spec.IPAdress
	resp.PublicKey = wgp.Status.PublicKey
	resp.RancherToken = tr[0].Token

	return resp, nil
}
