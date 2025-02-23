package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/go-playground/validator/v10"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthGRPC "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	userservergrpc "gitlab.com/ingvarmattis/auth/gen/servergrpc/user"
	"gitlab.com/ingvarmattis/auth/src/log"
)

const domain = "mattis.dev"

var ErrPortNotSpecified = errors.New("port not specified")

type GRPCUserHandlers interface {
	Auth(ctx context.Context, in *userservergrpc.AuthRequest) (*userservergrpc.AuthResponse, error)
	Register(ctx context.Context, in *userservergrpc.RegisterRequest) (*userservergrpc.RegisterResponse, error)
	ChangeEmail(ctx context.Context, req *userservergrpc.ChangeEmailRequest) (*emptypb.Empty, error)
	ChangePassword(ctx context.Context, req *userservergrpc.ChangePasswordRequest) (*emptypb.Empty, error)
}

type GRPCErrors interface {
	Error() string
}

type Server struct {
	userservergrpc.UnimplementedUserServiceServer

	UserHandlers GRPCUserHandlers

	Validator *validator.Validate
	Logger    *log.Zap

	grpcServer *grpc.Server
	httpServer *runtime.ServeMux
}

func (s *Server) Serve(serviceName string, port *int) error {
	if port == nil {
		return ErrPortNotSpecified
	}

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		return err
	}

	s.serveHealthCheck(serviceName)

	s.Logger.Info("starting grpc server", zap.Int("port", *port))

	if err = s.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("error while serve grpc | %w", err)
	}

	return nil
}

func (s *Server) ServeHTTP(port *int) error {
	if port == nil {
		return ErrPortNotSpecified
	}

	s.Logger.Info("starting http server", zap.Int("port", *port))

	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", *port),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && r.Header.Get("Content-Type") == "application/grpc" {
				s.grpcServer.ServeHTTP(w, r)
				return
			}

			if len(r.URL.Path) > 1 && r.URL.Path[len(r.URL.Path)-1] == '/' {
				http.Redirect(w, r, r.URL.Path[:len(r.URL.Path)-1], http.StatusPermanentRedirect)
				return
			}

			s.httpServer.ServeHTTP(w, r)
		})); err != nil {
		return fmt.Errorf("error while serve http | %w", err)
	}

	return nil
}

func (s *Server) serveHealthCheck(serviceName string) {
	healthCheckServer := health.NewServer()
	healthGRPC.RegisterHealthServer(s.grpcServer, healthCheckServer)
	healthCheckServer.SetServingStatus(serviceName, healthGRPC.HealthCheckResponse_SERVING)
}

func (s *Server) ServeWithCustomListener(l net.Listener) error {
	s.Logger.Info("starting grpc server with custom listener", zap.Int("port", l.Addr().(*net.TCPAddr).Port))

	if err := s.grpcServer.Serve(l); err != nil {
		return fmt.Errorf("error while Serve grpc | %w", err)
	}

	return nil
}

// Close stops the gRPC server gracefully. It stops the server from
// accepting new connections and RPCs and blocks until all the pending RPCs are
// finished.
func (s *Server) Close() {
	s.grpcServer.GracefulStop()
}

type NewServerOptions struct {
	ServiceName string

	GRPCAuthHandlers GRPCUserHandlers

	Logger    *log.Zap
	Validator *validator.Validate

	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor

	ServerOptions []grpc.ServerOption
}

func NewGRPCServer(ctx context.Context, grpcPort int, opts *NewServerOptions) *Server {
	srvOpts := make([]grpc.ServerOption, 0)

	srvOpts = append(
		srvOpts,
		grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(opts.UnaryInterceptors...)),
		grpc.StreamInterceptor(grpcMiddleware.ChainStreamServer(opts.StreamInterceptors...)),
	)

	grpcServer := grpc.NewServer(srvOpts...)

	httpServer := runtime.NewServeMux()
	httpOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if opts.Validator == nil {
		opts.Validator = validator.New()
	}

	s := Server{
		grpcServer:                     grpcServer,
		httpServer:                     httpServer,
		Validator:                      opts.Validator,
		UnimplementedUserServiceServer: userservergrpc.UnimplementedUserServiceServer{},

		UserHandlers: opts.GRPCAuthHandlers,

		Logger: opts.Logger,
	}
	userservergrpc.RegisterUserServiceServer(grpcServer, &s)

	if err := userservergrpc.RegisterUserServiceHandlerFromEndpoint(
		ctx, httpServer, fmt.Sprintf("0.0.0.0:%v", grpcPort), httpOpts,
	); err != nil {
		panic(err)
	}

	reflection.Register(grpcServer)

	return &s
}

func (s *Server) Auth(ctx context.Context, req *userservergrpc.AuthRequest) (*userservergrpc.AuthResponse, error) {
	if err := validate(s.Validator, req, errors.New("status error")); err != nil {
		return nil, err
	}

	resp, err := s.UserHandlers.Auth(ctx, req)
	if err != nil {
		return nil, GRPCUnauthorizedError(err, nil)
	}

	return resp, nil
}

func (s *Server) Register(ctx context.Context, req *userservergrpc.RegisterRequest) (*userservergrpc.RegisterResponse, error) {
	if err := validate(s.Validator, req, errors.New("status error")); err != nil {
		return nil, err
	}

	resp, err := s.UserHandlers.Register(ctx, req)
	if err != nil {
		return nil, GRPCUnknownError(err, nil)
	}

	return resp, nil
}

func (s *Server) ChangeEmail(ctx context.Context, req *userservergrpc.ChangeEmailRequest) (*emptypb.Empty, error) {
	if err := validate(s.Validator, req, errors.New("status error")); err != nil {
		return nil, err
	}

	resp, err := s.UserHandlers.ChangeEmail(ctx, req)
	if err != nil {
		return nil, GRPCUnknownError(err, nil)
	}

	return resp, nil
}

func (s *Server) ChangePassword(ctx context.Context, req *userservergrpc.ChangePasswordRequest) (*emptypb.Empty, error) {
	if err := validate(s.Validator, req, errors.New("status error")); err != nil {
		return nil, err
	}

	resp, err := s.UserHandlers.ChangePassword(ctx, req)
	if err != nil {
		return nil, GRPCUnknownError(err, nil)
	}

	return resp, nil
}

func GRPCUnauthorizedError[T GRPCErrors](reason T, err error) error {
	return gRPCError(codes.Unauthenticated, reason, err)
}

func GRPCValidationError[T GRPCErrors](reason T, err error) error {
	return gRPCError(codes.InvalidArgument, reason, err)
}

func GRPCBusinessError[T GRPCErrors](reason T, err error) error {
	return gRPCError(codes.FailedPrecondition, reason, err)
}

func GRPCUnknownError[T GRPCErrors](reason T, err error) error {
	return gRPCError(codes.Unknown, reason, err)
}

func GRPCCustomError[T GRPCErrors](code codes.Code, reason T, err error) error {
	return gRPCError(code, reason, err)
}

func gRPCError[T GRPCErrors](code codes.Code, reason T, serviceErr error) error {
	if serviceErr == nil {
		serviceErr = errors.New("error not set")
	}

	st, err := status.Newf(code, "error: %v", serviceErr.Error()).WithDetails(
		&errdetails.ErrorInfo{
			Reason:   reason.Error(),
			Domain:   domain,
			Metadata: nil,
		},
	)
	if err != nil {
		panic(fmt.Sprintf("unexpected error attaching metadata: %v", err))
	}

	return st.Err()
}

func validate[T GRPCErrors](v *validator.Validate, req any, reason T) error {
	if err := v.Struct(req); err != nil {
		return GRPCValidationError(reason, err)
	}

	return nil
}
