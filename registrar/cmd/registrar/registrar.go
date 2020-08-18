package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jaredallard-home/worker-nodes/registrar/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tritonmedia/pkg/app"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// copyFile is a suitable file copier for small files
func copyFile(src, dest string) error {
	f, err := os.Stat(src)
	if err != nil {
		return errors.Wrap(err, "failed to stat src")
	}

	b, err := ioutil.ReadFile(src)
	if err != nil {
		return errors.Wrap(err, "failed to read src")
	}

	return errors.Wrap(ioutil.WriteFile(dest, b, f.Mode()), "failed to copy src to dest")
}

func installK3S(ctx context.Context) error {
	k3sBin := "/host/usr/local/bin/k3s"
	if _, err := os.Stat(k3sBin); !os.IsNotExist(err) {
		// k3s already exists, skip...
		// TODO(jaredallard): checksum validation and all of that would be nice
		return nil
	}

	downloadSuffix := ""
	switch runtime.GOARCH {
	case "arm":
		downloadSuffix = "-armhf"
	case "amd64":
		// we don't set a suffix for amd64
	default:
		downloadSuffix = "-" + runtime.GOARCH
	}

	url := "https://github.com/rancher/k3s/releases/download/v1.18.8%2Bk3s1/k3s" + downloadSuffix
	log.WithFields(log.Fields{"url": url, "arch": runtime.GOARCH}).Info("downloading k3s")
	r, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, "failed to download k3s")
	}
	defer r.Body.Close()

	f, err := os.Create(k3sBin)
	if err != nil {
		return errors.Wrap(err, "failed to open k3s")
	}
	defer f.Close()

	if _, err := io.Copy(f, r.Body); err != nil {
		return errors.Wrap(err, "failed to download k3s")
	}

	return errors.Wrap(os.Chmod(k3sBin, 0777), "failed to +x k3s")
}

func leaderMode(ctx context.Context, c *cli.Context) error { //nolint:funlen
	if err := installK3S(ctx); err != nil {
		return err
	}

	return errors.Wrap(
		copyFile("/opt/registrar/systemd/k3s-server.service", "/host/etc/systemd/system/k3s.service"),
		"failed to copy systemd unit file",
	)
}

func agentMode(ctx context.Context, resp *api.RegisterResponse) error {
	if err := installK3S(ctx); err != nil {
		return err
	}

	log.Info("generating k3s env config")

	conf := fmt.Sprintf("K3S_URL=%s\nK3S_TOKEN=%s\n", resp.ClusterHost, resp.ClusterToken)

	if err := ioutil.WriteFile("/host/etc/registrar/k3s", []byte(conf), 0600); err != nil {
		return errors.Wrap(err, "failed to write k3s config to host")
	}

	return errors.Wrap(
		copyFile("/opt/registrar/systemd/k3s-agent.service", "/host/etc/systemd/system/k3s-agent.service"),
		"failed to copy systemd unit file",
	)
}

func main() { //nolint:funlen,gocyclo
	ctx := context.Background()

	app := cli.App{
		Name:    "registrar",
		Usage:   "Configure a device using a remote registrar server",
		Version: app.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "registrard-host",
				Usage:   "Specify the registrard hostname",
				EnvVars: []string{"REGISTRARD_HOST"},
				Value:   "127.0.0.1:8000",
			},
			&cli.BoolFlag{
				Name:    "leader-mode",
				Usage:   "Run a node in leader mode.",
				EnvVars: []string{"LEADER_MODE"},
			},
			&cli.BoolFlag{
				Name:    "registrard-enable-tls",
				Usage:   "Enable TLS when talking to registrard",
				EnvVars: []string{"REGISTRARD_ENABLE_TLS"},
			},
			&cli.StringFlag{
				Name:    "registrard-token",
				Usage:   "registrard auth token",
				EnvVars: []string{"REGISTRARD_TOKEN"},
			},
		},
		Action: func(c *cli.Context) error {
			if c.Bool("leader-mode") {
				return leaderMode(ctx, c)
			}

			host := c.String("registrard-host")

			log.WithFields(log.Fields{"host": host}).
				Info("registering device with registrar")

			confDir := "/host/etc/registrar"
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
			if c.Bool("registrard-enable-tls") {
				grpcOption = append(grpcOption, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
			} else {
				grpcOption = append(grpcOption, grpc.WithInsecure())
			}

			conn, err := grpc.DialContext(ctx, host, grpcOption...)
			if err != nil {
				return errors.Wrap(err, "failed to connect to registrard")
			}

			r := api.NewRegistrarClient(conn)
			regResp, err := r.Register(ctx, &api.RegisterRequest{
				Id:        id,
				AuthToken: c.String("registrard-token"),
			})
			if err != nil {
				return errors.Wrap(err, "failed to register devices")
			}

			return errors.Wrap(agentMode(ctx, regResp), "failed to create agent")
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Fatalf("failed to start")
		// Stay up for an hour for debugging, if needed.
		time.Sleep(time.Minute * 60)
	}
}
