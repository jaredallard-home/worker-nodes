package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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
			log.WithFields(log.Fields{"host": host}).Info("registering device with registrar")

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

			return w.StartClient(host, resp.IpAddress, k, serverPub)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Fatalf("failed to start")
	}
}
