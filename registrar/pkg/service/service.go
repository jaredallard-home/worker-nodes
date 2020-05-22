package service

import (
	"context"
	"fmt"
)

type Service interface {
	Run(ctx context.Context) error
	Close() error
}

type Runner struct {
	services []Service
}

// NewServiceRunner creates a new service runner that launches a service
// in a goroutine and handles termination of other services when one
// fails to launch.
func NewServiceRunner(ctx context.Context, services []Service) *Runner {
	return &Runner{
		services: services,
	}
}

func (r *Runner) shutdown() {
	for _, s := range r.services {
		s.Close()
	}
}

// Run starts the service runner
func (r *Runner) Run(ctx context.Context) error {
	errChan := make(chan error)
	for _, s := range r.services {
		go func(s Service) {
			errChan <- s.Run(ctx)
		}(s)
	}

	completed := 0
	for {
		select {
		case err := <-errChan:
			if err == nil {
				completed++
				if completed == len(r.services) {
					return nil
				}

				continue
			}

			// an error occurred, shutdown all services
			r.shutdown()

			// return the last error that occurred
			return err
		case <-ctx.Done():
			r.shutdown()
			return fmt.Errorf("context cancelled")
		}
	}
}
