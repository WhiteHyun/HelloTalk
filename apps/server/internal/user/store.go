package user

import (
	"context"
	"database/sql"
	"fmt"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, name, avatar_url, created_at, updated_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return u, nil
}

func (s *Store) Update(ctx context.Context, id string, req UpdateRequest) (*User, error) {
	u := &User{}
	err := s.db.QueryRowContext(ctx,
		`UPDATE users SET name = COALESCE($1, name), updated_at = NOW()
		 WHERE id = $2
		 RETURNING id, email, name, avatar_url, created_at, updated_at`,
		req.Name, id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return u, nil
}

func (s *Store) UpdateAvatar(ctx context.Context, id string, avatarURL string) (*User, error) {
	u := &User{}
	err := s.db.QueryRowContext(ctx,
		`UPDATE users SET avatar_url = $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING id, email, name, avatar_url, created_at, updated_at`,
		avatarURL, id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update avatar: %w", err)
	}
	return u, nil
}
