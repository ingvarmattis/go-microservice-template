package user

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"

	userserver "gitlab.com/ingvarmattis/example/gen/servergrpc/user"
	"gitlab.com/ingvarmattis/example/src/services"
)

type Handlers struct {
	Service services.SvcLayer
}

func (s *Handlers) Auth(ctx context.Context, req *userserver.AuthRequest) (*userserver.AuthResponse, error) {
	jwtToken, err := s.Service.AuthService.Auth(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, fmt.Errorf("cannot authorize | %w", err)
	}

	return &userserver.AuthResponse{JWT: jwtToken}, nil
}

func (s *Handlers) Register(ctx context.Context, req *userserver.RegisterRequest) (*userserver.RegisterResponse, error) {
	jwtToken, err := s.Service.AuthService.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, fmt.Errorf("cannot register | %w", err)
	}

	return &userserver.RegisterResponse{JWT: jwtToken}, nil
}

func (s *Handlers) ChangeEmail(ctx context.Context, req *userserver.ChangeEmailRequest) (*emptypb.Empty, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get user id | %w", err)
	}

	if err = s.Service.AuthService.EditEmail(ctx, userID, req.GetCurrentEmail(), req.GetNewEmail()); err != nil {
		return nil, fmt.Errorf("cannot edit email | %w", err)
	}

	return nil, nil
}

func (s *Handlers) ChangePassword(ctx context.Context, req *userserver.ChangePasswordRequest) (*emptypb.Empty, error) {
	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot get user id | %w", err)
	}

	if err = s.Service.AuthService.EditPassword(ctx, userID, req.GetCurrentPassword(), req.GetNewPassword()); err != nil {
		return nil, fmt.Errorf("cannot edit password | %w", err)
	}

	return nil, nil
}

func userIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value("user_id").(int)
	if !ok {
		return 0, fmt.Errorf("cannot get user id")
	}

	return userID, nil
}
