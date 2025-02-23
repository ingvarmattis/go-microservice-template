package example

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const packageName = "example"

var ErrNotFound = errors.New("not found")

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (p *Postgres) ServiceName(ctx context.Context) (string, error) {
	ctx, span := otel.Tracer(packageName).Start(ctx, "ServiceName")
	defer span.End()

	query := `
select service_name
from example.services
limit 1;`

	span.SetAttributes(attribute.String("query", query))

	row := p.pool.QueryRow(ctx, query)

	var serviceName string
	if err := row.Scan(&serviceName); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return "", ErrNotFound
	}

	return serviceName, nil
}

func (p *Postgres) RegisterService(ctx context.Context, serviceName string) error {
	ctx, span := otel.Tracer(packageName).Start(ctx, "RegisterService")
	defer span.End()

	query := `
insert into example.services
values ($1)
on conflict (service_name) do nothing;`

	span.SetAttributes(attribute.String("query", query))

	if _, err := p.pool.Exec(ctx, query, serviceName); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to insert service | %w", err)
	}

	return nil
}

func (p *Postgres) Exists(ctx context.Context, serviceName string) (bool, error) {
	ctx, span := otel.Tracer(packageName).Start(ctx, "Exists")
	defer span.End()

	query := `
select exists (
    select service_name
    from example.services
    where service_name = $1
);`

	span.SetAttributes(attribute.String("query", query))

	row := p.pool.QueryRow(ctx, query, serviceName)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		span.SetStatus(codes.Error, err.Error())
		return false, fmt.Errorf("cannot get service name | %w", err)
	}

	return exists, nil
}
