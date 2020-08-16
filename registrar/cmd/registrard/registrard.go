package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jaredallard-home/worker-nodes/registrar/internal/registrard"
	"github.com/sirupsen/logrus"
	"github.com/tritonmedia/pkg/app"
	"github.com/tritonmedia/pkg/service"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log := logrus.New().WithContext(ctx)

	app := cli.App{
		Name:    "registrar",
		Version: app.Version,
	}
	app.Action = func(c *cli.Context) error {
		r := service.NewServiceRunner(ctx, []service.Service{
			&registrard.ShutdownService{},
			&registrard.GRPCService{},
		})
		sigC := make(chan os.Signal)

		// listen for signals that we want to cancel on, and cancel
		// the context if one is passed
		signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigC
			cancel()
		}()

		return r.Run(ctx, log)
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("failed to run: %v", err)
	}
}
