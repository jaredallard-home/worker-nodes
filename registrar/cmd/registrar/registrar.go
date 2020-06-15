package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	dockerclient "github.com/docker/docker/client"
	"github.com/jaredallard-home/worker-nodes/registrar/api"
	"github.com/jaredallard-home/worker-nodes/registrar/pkg/wghelper"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	app := cli.App{
		Name:  "registrar",
		Usage: "Configure a device using a remote registrar server",
		Authors: []cli.Author{
			{
				Name:  "Jared Allard",
				Email: "jaredallard@outlook.com",
			},
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "skip-wireguard-check",
				Usage: "Skip wireguard kernel module check",
			},
			cli.StringFlag{
				Name:   "registrard-host",
				Usage:  "Specify the registard hostname",
				EnvVar: "REGISTRARD_HOST",
				Value:  "127.0.0.1:8000",
			},
			cli.StringFlag{
				Name:   "rancher-endpoint",
				Usage:  "Rancher Endpoint",
				EnvVar: "RANCHER_HOST",
				Value:  "https://rancher.tritonjs.com",
			},
			cli.StringFlag{
				Name:   "wireguard-endpoint",
				Usage:  "Wireguard server endpoint",
				EnvVar: "WIREGUARD_HOST",
			},
			cli.BoolFlag{
				Name:  "no-agent",
				Usage: "Disable launching the rancher-agent, just spin up wireguard",
			},
		},
		Action: func(c *cli.Context) error {
			if !c.Bool("skip-wireguard-check") {
				// TODO(jaredallard): This breaks windows support
				b, err := ioutil.ReadFile("/proc/modules")
				if err != nil {
					return errors.Wrap(err, "failed to check for wireguard module, pass --skip-wireguard-check to disable")
				}

				if !strings.Contains(string(b), "wireguard") {
					return fmt.Errorf("failed to find wireguard kernel module, ensure it's loaded")
				}
			}

			host := c.String("registrard-host")
			wireguardHost := c.String("wireguard-endpoint")
			rancherHost := c.String("rancher-endpoint")

			if wireguardHost == "" {
				wireguardHost = host
			}

			log.WithFields(log.Fields{"host": host, "wireguard": wireguardHost, "rancher": rancherHost}).
				Info("registering device with registrar")

			confDir := "/etc/registrar"
			ipConfDir := filepath.Join(confDir, "id")
			if _, err := os.Stat(confDir); err != nil {
				err := os.MkdirAll(confDir, 0755)
				if err != nil {
					return errors.Wrap(err, "failed to create configuration directory")
				}
			}

			var id string
			if b, err := ioutil.ReadFile(ipConfDir); err == nil {
				id = string(b)
			}

			grpcOption := make([]grpc.DialOption, 0)
			if os.Getenv("REGISTRARD_ENABLE_TLS") != "" {
				tlsConf := &tls.Config{}
				if os.Getenv("REGISTRARD_INSECURE") != "" {
					log.Warn("skipping TLS certificate host verification")
					tlsConf.InsecureSkipVerify = true
				}

				grpcOption = append(grpcOption, grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)))
			} else {
				grpcOption = append(grpcOption, grpc.WithInsecure())
			}

			ctx := context.Background()
			conn, err := grpc.DialContext(ctx, host, grpcOption...)
			if err != nil {
				return errors.Wrap(err, "failed to connect to registrard")
			}

			r := api.NewRegistrarClient(conn)
			resp, err := r.Register(ctx, &api.RegisterRequest{
				Id:        id,
				AuthToken: os.Getenv("REGISTRARD_TOKEN"),
			})
			if err != nil {
				return errors.Wrap(err, "failed to register devices")
			}

			k, err := wgtypes.ParseKey(resp.Key)
			if err != nil {
				return errors.Wrap(err, "failed to parse returned wireguard key")
			}

			serverPub, err := wgtypes.ParseKey(resp.PublicKey)
			if err != nil {
				return errors.Wrap(err, "failed to parse returned server wireguard public key")
			}

			log.WithFields(log.Fields{"ip": resp.IpAddress, "key": k.PublicKey, "id": id}).
				Infof("got registration information")

			// we didn't find one to start with, so we write it to disk
			if id == "" {
				if err := ioutil.WriteFile(ipConfDir, []byte(resp.Id), 0755); err != nil {
					log.Errorf("failed to save registration id")
				}
			}

			w, err := wghelper.NewWireguard(nil)
			if err != nil {
				return errors.Wrap(err, "failed to create wireguard device")
			}

			err = w.StartClient(wireguardHost, resp.IpAddress, k, serverPub)
			if err != nil {
				return errors.Wrap(err, "failed to start wireguard client")
			}

			if c.Bool("no-agent") {
				return nil
			}

			// TODO(jaredallard): get this from the server
			dockerImage := "docker.io/rancher/rancher-agent:v2.4.4"
			cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
			if err != nil {
				return errors.Wrap(err, "failed to create docker client")
			}

			if _, err := cli.ContainerInspect(ctx, "rancher-agent"); err != nil {
				reader, err := cli.ImagePull(ctx, dockerImage, types.ImagePullOptions{})
				if err != nil {
					return errors.Wrap(err, "failed to pull docker image")
				}

				if _, err := io.Copy(os.Stdout, reader); err != nil {
					return errors.Wrap(err, "failed to pull docker image")
				}

				cont, err := cli.ContainerCreate(
					ctx,
					&container.Config{
						Image: dockerImage,
						Cmd: []string{
							"--server",
							rancherHost,
							"--token",
							resp.RancherToken,
							"--worker",
						},
					},
					&container.HostConfig{
						Privileged: true,

						Mounts: []mount.Mount{
							{
								Type:   mount.TypeBind,
								Source: "/etc/kubernetes",
								Target: "/etc/kubernetes",
							},
							{
								Type:   mount.TypeBind,
								Source: "/var/run/balena.sock",
								Target: "/var/run/docker.sock",
							},
						},
						NetworkMode: "host",
						RestartPolicy: container.RestartPolicy{
							Name: "unless-stopped",
						},
					}, nil, "rancher-agent")
				if err != nil {
					return errors.Wrap(err, "failed to create rancher agent container")
				}

				// start the container
				return cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Fatalf("failed to start")
	}

	// Stay up for an hour for debugging, if needed.
	time.Sleep(time.Minute * 60)
}
