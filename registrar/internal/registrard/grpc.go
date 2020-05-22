package registrard

import (
	"context"
	"net"
	"strconv"

	"github.com/jaredallard-home/worker-nodes/registrar/api"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type GRPCService struct {
	lis *net.Listener
	srv *grpc.Server
}

func (s *GRPCService) Run(ctx context.Context) error {
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

	s.srv = grpc.NewServer()
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

type rpcservice struct {
	api.Service
}

// Register registers a new device into the wireguard VPN and returns the information
// needed to join a Kubernetes Cluster
func (s *rpcservice) Register(ctx context.Context, r *api.RegisterRequest) (*api.RegisterResponse, error) {
	return s.Service.Register(ctx, r)
}
