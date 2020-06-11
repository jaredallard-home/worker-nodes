package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/mount"
	"github.com/jaredallard-home/worker-nodes/registrar/api"
	"github.com/jaredallard-home/worker-nodes/registrar/pkg/wghelper"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"google.golang.org/grpc"
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

			ctx := context.Background()
			conn, err := grpc.DialContext(ctx, host, grpc.WithInsecure())
			if err != nil {
				return errors.Wrap(err, "failed to connect to registrard")
			}

			r := api.NewRegistrarClient(conn)
			resp, err := r.Register(ctx, &api.RegisterRequest{
				Id: id,
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

			// TODO(jaredallard): get this from the server
			dockerImage := "rancher/rancher-agent:v2.4.4"
			cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv)
			if err != nil {
				return errors.Wrap(err, "failed to create docker client")
			}
			// cli.ImagePull(ctx, dockerImage, types.ImagePullOptions{})

			if _, err := cli.ContainerInspect(ctx, "rancher-agent"); err != nil {
				log.WithError(err).Info("creating rancher-agent")
				cmd := exec.Command("docker", "pull", dockerImage)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					return errors.Wrap(err, "failed to pull image")
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
								Source: "/etc/kubernetes",
								Target: "/etc/kubernetes",
							},
							{
								Source: "/var/run",
								Target: "/var/run",
							},
						},
						NetworkMode: "host",
						RestartPolicy: container.RestartPolicy{
							Name:              "unless-stopped",
							MaximumRetryCount: -1,
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
}
