package services

import (
	"context"
)

type SvcLayer struct {
	AuthService AuthService
}

type AuthService interface {
	Auth(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email, password string) (string, error)

	EditEmail(ctx context.Context, userID int, currentEmail, newEmail string) error
	EditPassword(ctx context.Context, userID int, currentPassword, newPassword string) error
}
