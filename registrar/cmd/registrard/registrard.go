package main

import (
	"context"
	"os"

	registard "github.com/jaredallard-home/worker-nodes/registrar/internal/registrard"
	"github.com/jaredallard-home/worker-nodes/registrar/pkg/service"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	app := cli.App{
		Name:  "registrard",
		Usage: "Launch a registrar server instance",
		Authors: []cli.Author{
			{
				Name:  "Jared Allard",
				Email: "jaredallard@outlook.com",
			},
		},
		Action: func(c *cli.Context) error {
			log.Info("starting registrard")

			ctx := context.Background()

			r := service.NewServiceRunner(ctx, []service.Service{
				&registard.ShutdownService{},
				&registard.GRPCService{},
			})

			return r.Run(ctx)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("failed to start: %v", err)
	}
}
