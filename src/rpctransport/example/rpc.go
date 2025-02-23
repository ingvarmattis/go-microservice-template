package example

import (
	"context"
	"fmt"

	servergrpc "github.com/ingvarmattis/example/gen/servergrpc/example"
	"github.com/ingvarmattis/example/src/services"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handlers struct {
	Service services.SvcLayer
}

func (s *Handlers) ServiceName(ctx context.Context, _ *emptypb.Empty) (*servergrpc.ServiceNameResponse, error) {
	serviceName, err := s.Service.ExampleService.ServiceName(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get service name | %w", err)
	}

	return &servergrpc.ServiceNameResponse{Name: serviceName}, nil
}

func (s *Handlers) Status(ctx context.Context, req *servergrpc.StatusRequest) (*servergrpc.StatusResponse, error) {
	exists, err := s.Service.ExampleService.Exists(ctx, req.GetServiceName())
	if err != nil {
		return nil, fmt.Errorf("cannot get service | %w", err)
	}

	return mapStatus(exists, err)
}

func mapStatus(exists bool, err error) (*servergrpc.StatusResponse, error) {
	switch {
	case err != nil:
		return nil, fmt.Errorf("cannot get status | %w", err)
	case exists:
		return &servergrpc.StatusResponse{Status: servergrpc.Status_REGISTERED}, nil
	default:
		return &servergrpc.StatusResponse{Status: servergrpc.Status_NOT_REGISTERED}, nil
	}
}
