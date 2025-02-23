package services

import (
	"context"
)

type SvcLayer struct {
	ExampleService ExampleService
}

type ExampleService interface {
	ServiceName(ctx context.Context) (string, error)
	Exists(ctx context.Context, serviceName string) (bool, error)
}
