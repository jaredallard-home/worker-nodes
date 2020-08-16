package registrard

import (
	"context"
	"net"
	"os"
	"strconv"

	"github.com/jaredallard-home/worker-nodes/registrar/api"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type GRPCService struct {
	lis *net.Listener
	srv *grpc.Server
}

func (s *GRPCService) Run(ctx context.Context, log logrus.FieldLogger) error { //nolint:funlen
	listAddr := ":" + strconv.Itoa(8000)
	l, err := net.Listen("tcp", listAddr)
	if err != nil {
		return err
	}
	s.lis = &l

	server, err := NewServer(ctx)
	if err != nil {
		return err
	}

	serverOpts := make([]grpc.ServerOption, 0)
	if os.Getenv("REGISTRARD_ENABLE_TLS") != "" {
		pem := os.Getenv("REGISTRARD_PEM_FILEPATH")
		key := os.Getenv("REGISTRARD_KEY_FILEPATH")
		creds, err := credentials.NewServerTLSFromFile(pem, key)
		if err != nil {
			log.WithError(err).Fatalf("failed to setup tls")
		}

		serverOpts = append(serverOpts, grpc.Creds(creds))
	}

	s.srv = grpc.NewServer(serverOpts...)
	api.RegisterRegistrarServer(s.srv, server)

	// Note: .Serve() blocks
	log.Info("Serving GRPC Service on " + listAddr)
	if err := s.srv.Serve(l); err != nil {
		log.Errorf("unexpected grpc Serve error: %v", err)
		return err
	}

	return nil
}

func (s *GRPCService) Close() error {
	if s.srv != nil {
		s.srv.GracefulStop()
	}
	if s.lis != nil {
		return (*s.lis).Close()
	}
	log.Infof("grpc service shutdown")
	return nil
}

type rpcservice struct { //nolint:unused
	api.Service
}

// Register registers a new device into the wireguard VPN and returns the information
// needed to join a Kubernetes Cluster
func (s *rpcservice) Register(ctx context.Context, r *api.RegisterRequest) (*api.RegisterResponse, error) {
	return s.Service.Register(ctx, r)
}
