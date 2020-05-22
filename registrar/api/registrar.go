package api

import "context"

//go:generate protoc -I. --go_out=plugins=grpc,paths=source_relative:. ./registrar.proto

// Service is the registrar server interface
//
// This interface is implemented by the server and the rpc client
type Service interface {
	Register(ctx context.Context, r *RegisterRequest) (*RegisterResponse, error)
}
