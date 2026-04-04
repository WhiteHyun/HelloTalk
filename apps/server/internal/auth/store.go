package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(ctx context.Context, email, passwordHash, name string) (*User, error) {
	user := &User{}
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, name)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, password_hash, name, avatar_url, created_at, updated_at`,
		email, passwordHash, name,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, avatar_url, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Name, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

func (s *Store) SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

func (s *Store) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	rt := &RefreshToken{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at
		 FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}
	return rt, nil
}

func (s *Store) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	)
	return err
}

func (s *Store) DeleteAllRefreshTokens(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM refresh_tokens WHERE user_id = $1`,
		userID,
	)
	return err
}
