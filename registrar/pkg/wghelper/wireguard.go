package wghelper

import (
	"context"
	"fmt"
	"net"
	"time"

	registrar "github.com/jaredallard-home/worker-nodes/registrar/apis/clientset/v1alpha1"
	"github.com/jaredallard-home/worker-nodes/registrar/apis/types/v1alpha1"
	"github.com/pkg/errors"
	wgnetlink "github.com/schu/wireguard-cni/pkg/netlink"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Wireguard struct {
	device *wgtypes.Device
	w      *wgctrl.Client
	k      *registrar.RegistrarClientset
	l      netlink.Link
}

// NewWireguard creates a new wireguard configuration instance, that stores
// IP information in Kubernetes
func NewWireguard(k *registrar.RegistrarClientset) (*Wireguard, error) {
	w, err := wgctrl.New()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create wireguard controller")
	}

	resp := &Wireguard{
		w: w,
		k: k,
	}

	devices, err := w.Devices()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list wireguard devices")
	}

	if len(devices) > 1 {
		return nil, fmt.Errorf("found more than one wireguard device, only one is supported")
	}

	// attempt to create a wireguard interface
	if len(devices) == 0 {
		log.Infof("creating a wireguard interface")

		attrs := netlink.NewLinkAttrs()
		attrs.Name = "wg0"

		l := &wgnetlink.Wireguard{
			LinkAttrs: attrs,
		}

		if err := netlink.LinkAdd(l); err != nil {
			return nil, errors.Wrap(err, "failed to create link")
		}

		resp.device, err = w.Device(attrs.Name)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get created wireguard link")
		}
	} else {
		resp.device = devices[0]
	}

	resp.l, err = netlink.LinkByName(resp.device.Name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get link by device name")
	}

	return resp, nil
}

func (w *Wireguard) StartClient(endpoint string, ourIP string, k wgtypes.Key, pubk wgtypes.Key) error {
	ip, _, err := net.ParseCIDR(ourIP + "/32")
	if err != nil {
		return errors.Wrap(err, "failed to parse wireguard ip address")
	}

	udpAddr, err := net.ResolveUDPAddr("udp", "172.92.139.101:51820")
	if err != nil {
		return errors.Wrap(err, "failed to resolve endpoint address")
	}

	pki := time.Duration(5 * time.Second)
	_, globalCidr, _ := net.ParseCIDR("0.0.0.0/0")
	peer := &wgtypes.PeerConfig{
		PublicKey:         pubk,
		UpdateOnly:        false,
		ReplaceAllowedIPs: true,
		AllowedIPs:        []net.IPNet{*globalCidr},
		Endpoint:          udpAddr,
		// Allows this peer to survive when running behind NAT
		PersistentKeepaliveInterval: &pki,
	}

	// add the peer to our device
	err = w.w.ConfigureDevice(w.device.Name, wgtypes.Config{
		PrivateKey: &k,

		Peers:        []wgtypes.PeerConfig{*peer},
		ReplacePeers: true,
	})
	if err != nil {
		return errors.Wrap(err, "failed to configure wireguard device")
	}

	err = netlink.AddrReplace(w.l, &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: net.IPv4bcast.DefaultMask(),
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to set ip address on wireguard interface")
	}

	err = netlink.LinkSetUp(w.l)
	return errors.Wrap(err, "failed to set link to up")
}

func (w *Wireguard) StartServer(ipool *v1alpha1.WireguardIPPool) error {
	if w.device.PrivateKey.String() == "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" {
		log.Info("failed to find initialized device, creating new server")
		if err := w.initServer(ipool); err != nil {
			return errors.Wrap(err, "failed to init server ")
		}
	}

	ip, cidr, err := net.ParseCIDR(ipool.Spec.CIDR)
	if err != nil {
		return errors.Wrap(err, "failed to parse CIDR")
	}

	// detect a null ip address, if set
	if lo, err := netlink.LinkByName("lo"); err == nil {
		addrs, err := netlink.AddrList(lo, 0)
		if err == nil {
			for _, addr := range addrs {
				if addr.IP.String() == ip.String() {
					err := netlink.AddrDel(lo, &addr)
					if err != nil {
						log.Errorf("failed to remove null address on loopback: %v", err)
					} else {
						log.Infof("successfully removed a loopback placeholder IP address from device lo")
					}
				}
			}
		}
	} else if err != nil {
		log.Warnf("failed to find lo link to check for null ip: %v", err)
	}

	err = netlink.AddrReplace(w.l, &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ip,
			Mask: net.IPv4bcast.DefaultMask(),
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to assign IP to wg0")
	}

	if err := netlink.LinkSetUp(w.l); err != nil {
		return errors.Wrap(err, "failed to set link to up")
	}

	log.WithFields(log.Fields{"ip": ip, "cidr": cidr}).Info("wireguard server started")

	return nil
}

// Flush removes all peer from a wireguard instance
func (w *Wireguard) Flush(ctx context.Context) error {
	return w.w.ConfigureDevice(w.device.Name, wgtypes.Config{
		Peers:        []wgtypes.PeerConfig{},
		ReplacePeers: true,
	})
}

// initServer initializes a new wireguard server
func (w *Wireguard) initServer(ipool *v1alpha1.WireguardIPPool) error {
	if ipool.Status.SecretRef == "" {
		log.Info("failed to find a secret key for this ippool, creating new one")
		privk, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return errors.Wrap(err, "failed to generate private key")
		}

		// TODO(jaredallard): default hardcode
		secretName := fmt.Sprintf("wgipp-%s", ipool.ObjectMeta.Name)
		_, err = w.k.CoreV1().Secrets("default").Create(context.TODO(), &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: secretName,
			},
			StringData: map[string]string{
				"privk": privk.String(),
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return errors.Wrap(err, "failed to store wireguard secret in kubernetes")
		}

		ipool.Status.SecretRef = secretName
		ipool.Status.Created = true
		ipool.Status.PublicKey = privk.PublicKey().String()

		_, err = w.k.RegistrarV1Alpha1Client().WireguardIPPools("default").Update(context.TODO(), ipool)
		if err != nil {
			return errors.Wrap(err, "failed to update ipool in k8s")
		}
	}

	sec, err := w.k.CoreV1().Secrets("default").Get(context.TODO(), ipool.Status.SecretRef, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get wireguard server privk")
	}

	privk, err := wgtypes.ParseKey(string(sec.Data["privk"]))
	if err != nil {
		return errors.Wrap(err, "failed to parse wireguard server privk")
	}

	// set our server's private key
	port := 51820
	err = w.w.ConfigureDevice(w.device.Name, wgtypes.Config{
		ListenPort: &port,
		PrivateKey: &privk,
	})
	if err != nil {
		return errors.Wrap(err, "failed to configure wireguard device")
	}

	return nil
}

// Register adds a new peer to a device, and returns the information needed to connect
// as said peer
func (w *Wireguard) Register(ctx context.Context, ip *v1alpha1.WireguardIP) (*wgtypes.PeerConfig, error) {
	var privk wgtypes.Key

	// TODO(jaredallard): this entire section needs help.
	d, err := w.k.RegistrarV1Alpha1Client().Devices("default").
		Get(ctx, ip.Spec.DeviceRef, metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		// device was just created, so we generate a private key for it
		var err error
		privk, err = wgtypes.GeneratePrivateKey()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate private key")
		}
	} else if err == nil {
		// we found the device, so we fetch the existing secret
		s, err := w.k.CoreV1().Secrets("default").Get(ctx, d.Spec.SecretRef, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "failed to get device ip secret")
		}

		privk, err = wgtypes.ParseKey(string(s.Data["wireguard-key"]))
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse private key for device")
		}
	} else if err != nil {
		// other error occurred, we fail
		return nil, err
	}

	pki := 5 * time.Second

	// HACK: better way...
	_, cidr, err := net.ParseCIDR(ip.Spec.IPAdress + "/32")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cidr from ip address")
	}
	cidr.IP = net.ParseIP(ip.Spec.IPAdress)

	log.WithContext(ctx).WithFields(log.Fields{"ip": ip.Spec.IPAdress}).Info("adding wireguard peer")
	peer := &wgtypes.PeerConfig{
		PublicKey:         privk.PublicKey(),
		UpdateOnly:        false,
		ReplaceAllowedIPs: true,
		// Allows this peer to survive when running behind NAT
		PersistentKeepaliveInterval: &pki,
		AllowedIPs:                  []net.IPNet{*cidr},
	}

	// add the peer to our device
	err = w.w.ConfigureDevice(w.device.Name, wgtypes.Config{
		Peers: []wgtypes.PeerConfig{*peer},
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to configure wireguard device")
	}

	return peer, err
}
