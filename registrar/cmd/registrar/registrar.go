package main

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jaredallard-home/worker-nodes/registrar/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tritonmedia/pkg/app"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func leaderMode(ctx context.Context, c *cli.Context) error { //nolint:funlen
	cmd := exec.Command("bash", "/tmp/k3s-install.sh")
	cmd.Env = append(
		os.Environ(),
		"INSTALL_K3S_SKIP_ENABLE=true",
		"INSTALL_K3S_SKIP_START=true",
		"INSTALL_K3S_BIN_DIR=/host/usr/local/bin",
		"INSTALL_K3S_SYSTEMD_DIR=/host/etc/systemd/system",
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start k3s install")
	}

	return errors.Wrap(cmd.Wait(), "failed to install k3s")
}

func agentMode(ctx context.Context, resp *api.RegisterResponse) error {
	cmd := exec.Command("bash", "/tmp/k3s-install.sh")
	cmd.Env = append(
		os.Environ(),
		"INSTALL_K3S_SKIP_ENABLE=true",
		"INSTALL_K3S_SKIP_START=true",
		"INSTALL_K3S_BIN_DIR=/host/usr/local/bin",
		"INSTALL_K3S_SYSTEMD_DIR=/host/etc/systemd/system",
		"K3S_URL=https://myserver:6443",
		"K3S_TOKEN=mynodetoken",
	)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return errors.Wrap(cmd.Wait(), "failed to install k3s")
}

func main() { //nolint:funlen,gocyclo
	ctx := context.Background()

	app := cli.App{
		Name:    "registrar",
		Usage:   "Configure a device using a remote registrar server",
		Version: app.Version,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "registrard-host",
				Usage:  "Specify the registrard hostname",
				EnvVar: "REGISTRARD_HOST",
				Value:  "127.0.0.1:8000",
			},
			cli.BoolFlag{
				Name:   "leader-mode",
				Usage:  "Run a node in leader mode.",
				EnvVar: "LEADER_MODE",
			},
			cli.BoolFlag{
				Name:   "registrard-enable-tls",
				Usage:  "Enable TLS when talking to registrard",
				EnvVar: "REGISTRARD_ENABLE_TLS",
			},
			cli.StringFlag{
				Name:   "registrard-token",
				Usage:  "registrard auth token",
				EnvVar: "REGISTRARD_TOKEN",
			},
		},
		Action: func(c *cli.Context) error {
			if _, err := os.Stat("/tmp/k3s-install.sh"); os.IsNotExist(err) {
				_ = os.Mkdir("/run/systemd", 0777)
				// TODO(jaredallard): move off of this one day
				log.Info("fetching k3s install script")
				resp, err := http.Get("https://raw.githubusercontent.com/rancher/k3s/master/install.sh")
				if err != nil {
					return errors.Wrap(err, "failed to download k3s install script")
				}
				defer resp.Body.Close()

				b, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return errors.Wrap(err, "failed to read install script from remote")
				}

				if err := ioutil.WriteFile("/tmp/k3s-install.sh", b, 0755); err != nil {
					return errors.Wrap(err, "failed to write install script")
				}
			}

			if c.Bool("leader-mode") {
				return leaderMode(ctx, c)
			}

			host := c.String("registrard-host")

			log.WithFields(log.Fields{"host": host}).
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
