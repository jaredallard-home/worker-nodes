package registrard

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

type ShutdownService struct {
	c chan os.Signal
}

func (s *ShutdownService) Run(ctx context.Context, log logrus.FieldLogger) error {
	// listen for interrupts and gracefully shutdown server
	s.c = make(chan os.Signal, 10)
	signal.Notify(s.c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	if out, ok := <-s.c; ok {
		return fmt.Errorf("shutting down due to interrupt: %v", out)
	}

	return nil
}

func (s *ShutdownService) Close() error {
	close(s.c)
	return nil
}
