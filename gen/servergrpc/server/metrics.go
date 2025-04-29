package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/ingvarmattis/example/src/box"
	"github.com/ingvarmattis/example/src/log"
)

type MetricsServer struct {
	*http.Server
	name string
	port int

	logger *log.Zap
}

func (m *MetricsServer) Addr() string {
	return m.Server.Addr
}

func (m *MetricsServer) Name() string {
	return m.name
}

func NewMetricsServer(enabled bool, logger *log.Zap, port int) *MetricsServer {
	if !enabled {
		return &MetricsServer{
			name:   box.NotOperational,
			Server: nil,
			port:   port,
			logger: logger,
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &MetricsServer{
		name: "prometheus",
		Server: &http.Server{
			ReadHeaderTimeout: time.Minute,
			Handler:           mux,
			Addr:              ":" + strconv.Itoa(port),
		},
		port:   port,
		logger: logger,
	}
}

func (m *MetricsServer) ListenAndServe() error {
	m.logger.Info("starting http metrics server", zap.Int("port", m.port))

	if err := m.Server.ListenAndServe(); err != nil {
		return fmt.Errorf("cannot start http metrics server | %w", err)
	}

	return nil
}
