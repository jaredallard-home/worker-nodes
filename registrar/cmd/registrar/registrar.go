package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jaredallard-home/worker-nodes/registrar/api"
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
			cli.StringFlag{
				Name:   "registrard-host",
				EnvVar: "REGISTRARD_HOST",
				Value:  "127.0.0.1:8000",
			},
		},
		Action: func(c *cli.Context) error {
			host := c.String("registrard-host")
			log.WithFields(log.Fields{"host": host}).Info("registering device with registrar")

			confDir := "/etc/registrar"
			ipConfDir := filepath.Join(confDir, "id")
			if _, err := os.Stat(confDir); err != nil {
				err := os.MkdirAll(confDir, 0755)
				if err != nil {
					log.WithError(err).Warn("failed to create configuration directory, will fail to re-associate on reboot")
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

			log.WithFields(log.Fields{"ip": resp.IpAddress, "key": k.PublicKey}).
				Infof("got registration information")

			// we didn't find one to start with, so we write it to disk
			if id == "" {
				if err := ioutil.WriteFile(ipConfDir, []byte(resp.Id), 0755); err != nil {
					log.Errorf("failed to save registration id")
				}
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("failed to start: %v", err)
	}
}
