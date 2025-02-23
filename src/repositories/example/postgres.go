package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"

	"gitlab.com/ingvarmattis/example/src/encryption"
	"gitlab.com/ingvarmattis/example/src/hash"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyRegistered = errors.New("already registered")
)

type User struct {
	ID          int64
	Email       string
	LastLoginAt *time.Time
	CreatedAt   *time.Time
}

type Postgres struct {
	pool      *pgxpool.Pool
	encryptor *encryption.Crypto
}

func NewPostgres(pool *pgxpool.Pool, encryptor *encryption.Crypto) *Postgres {
	return &Postgres{
		pool:      pool,
		encryptor: encryptor,
	}
}

func (p *Postgres) Auth(ctx context.Context, email, password string) (int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "Auth")
	defer span.End()

	row := p.pool.QueryRow(ctx, `
update auth.users
set last_login_at = now()
where
    email_id = $1
    and password_hash = $2
returning id, email;`, hash.CityHash64(email), hash.CityHash64(password))

	var (
		id             int64
		encryptedEmail string
	)
	if err := row.Scan(&id, &encryptedEmail); err != nil {
		return 0, ErrNotFound
	}

	decryptedEmail, err := p.encryptor.Decrypt(encryptedEmail)
	if err != nil {
		return 0, fmt.Errorf("decrypt email | %w", err)
	}

	if decryptedEmail != email {
		return 0, ErrNotFound
	}

	return id, nil
}

func (p *Postgres) Register(ctx context.Context, email, password string) (userID int64, err error) {
	ctx, span := otel.Tracer("").Start(ctx, "Register")
	defer span.End()

	encryptedEmail, err := p.encryptor.Encrypt(email)
	if err != nil {
		return 0, fmt.Errorf("encryption failed | %w", err)
	}

	userID, err = p.userID(ctx, email)
	if errors.Is(err, ErrNotFound) {
		row := p.pool.QueryRow(ctx, `
insert into auth.users (email_id, email, password_hash)
values ($1, $2, $3)
returning id;`, hash.CityHash64(email), encryptedEmail, hash.CityHash64(password))

		if err = row.Scan(&userID); err != nil {
			return 0, fmt.Errorf("error creating user | %w", err)
		}

		return userID, nil
	}

	return 0, ErrAlreadyRegistered
}

func (p *Postgres) EditEmail(ctx context.Context, userID int, currentEmail, newEmail string) error {
	ctx, span := otel.Tracer("").Start(ctx, "EditEmail")
	defer span.End()

	newEncryptedEmail, err := p.encryptor.Encrypt(newEmail)
	if err != nil {
		return fmt.Errorf("encryption failed | %w", err)
	}

	row := p.pool.QueryRow(ctx, `
update auth.users
set
    email_id   = $3,
    email      = $4,
    updated_at = now()
where
    id           = $1
    and email_id = $2
returning id;`, userID, hash.CityHash64(currentEmail), hash.CityHash64(newEmail), newEncryptedEmail)

	var userIDAfterUpdate int64
	if err = row.Scan(&userIDAfterUpdate); err != nil {
		return fmt.Errorf("failed to update user's email | %w", err)
	}

	if userIDAfterUpdate == 0 {
		return ErrNotFound
	}

	return nil
}

func (p *Postgres) EditPassword(ctx context.Context, userID int, currentPassword, newPassword string) error {
	ctx, span := otel.Tracer("").Start(ctx, "EditPassword")
	defer span.End()

	row := p.pool.QueryRow(ctx, `
update auth.users
set
    password_hash = $3,
    updated_at = now()
where
    id = $1
    and password_hash = $2
returning id;`, userID, hash.CityHash64(currentPassword), hash.CityHash64(newPassword))

	var userIDAfterUpdate int64
	if err := row.Scan(&userIDAfterUpdate); err != nil {
		return fmt.Errorf("failed to update user's password | %w", err)
	}

	if userIDAfterUpdate == 0 {
		return ErrNotFound
	}

	return nil
}

func (p *Postgres) UserInfo(ctx context.Context, userID int) (*User, error) {
	ctx, span := otel.Tracer("").Start(ctx, "UserInfo")
	defer span.End()

	row := p.pool.QueryRow(ctx, `
select
    id,
    email,
    last_login_at,
    created_at
from auth.users
where id = $1;`, userID)

	var u User
	if err := row.Scan(&u.ID, &u.Email, &u.LastLoginAt, &u.CreatedAt); err != nil {
		return nil, ErrNotFound
	}

	return &u, nil
}

func (p *Postgres) userID(ctx context.Context, email string) (int64, error) {
	ctx, span := otel.Tracer("").Start(ctx, "userID")
	defer span.End()

	row := p.pool.QueryRow(ctx, `
select id
from auth.users
where email = $1;`, email)

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, ErrNotFound
	}

	return id, nil
}
