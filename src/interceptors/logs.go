package interceptors

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/ingvarmattis/example/src/log"
)

func UnaryServerLogInterceptor(logger *log.Zap, debugMode bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		startTime := time.Now()

		traceID := trace.SpanFromContext(ctx).SpanContext().TraceID().String()

		resp, err := handler(ctx, req)

		executionDuration := time.Since(startTime)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.String("traceID", traceID),
			zap.Duration("duration", executionDuration),
			zap.String("status", status.Code(err).String()),
		}

		if debugMode {
			fields = append(fields, zap.Any("request", req), zap.Any("response", resp))
		}

		logger.Info("incoming grpc request", fields...)

		return resp, err
	}
}
