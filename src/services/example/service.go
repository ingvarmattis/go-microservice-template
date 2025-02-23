package user

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"gitlab.com/ingvarmattis/example/src/repositories/user"
)

var (
	ErrAuthFailed        = errors.New("auth failed")
	ErrAlreadyRegistered = errors.New("already registered")
)

//go:generate bash -c "mkdir -p mocks"
//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
type authStorage interface {
	Auth(ctx context.Context, login, password string) (int64, error)
	Register(ctx context.Context, email, password string) (userID int64, err error)
	EditEmail(ctx context.Context, userID int, currentEmail, newEmail string) error
	EditPassword(ctx context.Context, userID int, currentPassword, newPassword string) error
}

type Service struct {
	secretKey []byte

	authStorage authStorage
}

func NewService(secretKey []byte, authStorage authStorage) *Service {
	return &Service{
		secretKey:   secretKey,
		authStorage: authStorage,
	}
}

func (s *Service) Auth(ctx context.Context, email, password string) (string, error) {
	userID, err := s.authStorage.Auth(ctx, email, password)
	if errors.Is(err, user.ErrNotFound) {
		return "", ErrAuthFailed
	}
	if err != nil {
		return "", fmt.Errorf("cannot auth | %w", err)
	}

	signedToken, err := s.generateToken(int(userID))
	if err != nil {
		return "", fmt.Errorf("cannot sign token | %w", err)
	}

	return signedToken, nil
}

func (s *Service) Register(ctx context.Context, email, password string) (string, error) {
	userID, err := s.authStorage.Register(ctx, email, password)
	if errors.Is(err, user.ErrAlreadyRegistered) {
		return "", ErrAlreadyRegistered
	}
	if err != nil {
		return "", fmt.Errorf("cannot register | %w", err)
	}

	signedToken, err := s.generateToken(int(userID))
	if err != nil {
		return "", fmt.Errorf("cannot sign token | %w", err)
	}

	return signedToken, nil
}

func (s *Service) EditEmail(ctx context.Context, userID int, currentEmail, newEmail string) error {
	if currentEmail == "" || newEmail == "" {
		return fmt.Errorf("email is required")
	}

	if err := s.authStorage.EditEmail(ctx, userID, currentEmail, newEmail); err != nil {
		return fmt.Errorf("cannot edit email | %w", err)
	}

	return nil
}

func (s *Service) EditPassword(ctx context.Context, userID int, currentPassword, newPassword string) error {
	if currentPassword == "" || newPassword == "" {
		return fmt.Errorf("password is required")
	}

	if currentPassword == newPassword {
		return fmt.Errorf("password must be different")
	}

	if err := s.authStorage.EditPassword(ctx, userID, currentPassword, newPassword); err != nil {
		return fmt.Errorf("cannot edit password | %w", err)
	}

	return nil
}

func (s *Service) generateToken(userID int) (string, error) {
	claims := &jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", fmt.Errorf("cannot sign token | %w", err)
	}

	return signedToken, nil
}
