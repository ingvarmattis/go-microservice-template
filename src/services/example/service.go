package example

import (
	"context"
	"errors"
	"fmt"

	exampleRepo "github.com/ingvarmattis/example/src/repositories/example"
)

const serviceName = "example-service"

var ErrNotFound = errors.New("not found")

//go:generate bash -c "mkdir -p mocks"
//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
type exampleStorage interface {
	ServiceName(ctx context.Context) (string, error)
	RegisterService(ctx context.Context, serviceName string) error
	Exists(ctx context.Context, serviceName string) (bool, error)
}

type Service struct {
	exampleStorage exampleStorage
}

func NewService(ctx context.Context, exampleStorage exampleStorage) (*Service, error) {
	service := &Service{exampleStorage: exampleStorage}

	if err := service.exampleStorage.RegisterService(ctx, serviceName); err != nil {
		return nil, fmt.Errorf("failed register service | %w", err)
	}

	return service, nil
}

func (s *Service) ServiceName(ctx context.Context) (string, error) {
	svcName, err := s.exampleStorage.ServiceName(ctx)
	if err != nil {
		if errors.Is(err, exampleRepo.ErrNotFound) {
			return "", ErrNotFound
		}

		return "", fmt.Errorf("cannot auth | %w", err)
	}

	return svcName, nil
}

func (s *Service) Exists(ctx context.Context, serviceName string) (bool, error) {
	exists, err := s.exampleStorage.Exists(ctx, serviceName)
	if err != nil {
		return false, fmt.Errorf("cannot auth | %w", err)
	}

	return exists, nil
}
