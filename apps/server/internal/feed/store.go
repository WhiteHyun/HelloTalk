package feed

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreatePost(ctx context.Context, userID, content string) (*Post, error) {
	p := &Post{}
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO posts (user_id, content)
		 VALUES ($1, $2)
		 RETURNING id, user_id, content, created_at, updated_at`,
		userID, content,
	).Scan(&p.ID, &p.UserID, &p.Content, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	return p, nil
}

func (s *Store) GetPost(ctx context.Context, postID, currentUserID string) (*PostResponse, error) {
	p := &PostResponse{}
	err := s.db.QueryRowContext(ctx,
		`SELECT p.id, p.user_id, u.name, p.content, p.created_at,
		        COUNT(l.user_id) AS like_count,
		        EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = $2) AS liked
		 FROM posts p
		 JOIN users u ON p.user_id = u.id
		 LEFT JOIN likes l ON p.id = l.post_id
		 WHERE p.id = $1
		 GROUP BY p.id, u.name`,
		postID, currentUserID,
	).Scan(&p.ID, &p.UserID, &p.UserName, &p.Content, &p.CreatedAt, &p.LikeCount, &p.Liked)
	if err != nil {
		return nil, fmt.Errorf("post not found: %w", err)
	}
	return p, nil
}

func (s *Store) GetFeed(ctx context.Context, currentUserID string, cursor time.Time, limit int) ([]PostResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.user_id, u.name, p.content, p.created_at,
		        COUNT(l.user_id) AS like_count,
		        EXISTS(SELECT 1 FROM likes WHERE post_id = p.id AND user_id = $1) AS liked
		 FROM posts p
		 JOIN users u ON p.user_id = u.id
		 LEFT JOIN likes l ON p.id = l.post_id
		 WHERE p.created_at < $2
		 GROUP BY p.id, u.name
		 ORDER BY p.created_at DESC
		 LIMIT $3`,
		currentUserID, cursor, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query feed: %w", err)
	}
	defer rows.Close()

	var posts []PostResponse
	for rows.Next() {
		var p PostResponse
		if err := rows.Scan(&p.ID, &p.UserID, &p.UserName, &p.Content, &p.CreatedAt, &p.LikeCount, &p.Liked); err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

func (s *Store) UpdatePost(ctx context.Context, postID, userID, content string) (*Post, error) {
	p := &Post{}
	err := s.db.QueryRowContext(ctx,
		`UPDATE posts SET content = $1, updated_at = NOW()
		 WHERE id = $2 AND user_id = $3
		 RETURNING id, user_id, content, created_at, updated_at`,
		content, postID, userID,
	).Scan(&p.ID, &p.UserID, &p.Content, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}
	return p, nil
}

func (s *Store) DeletePost(ctx context.Context, postID, userID string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM posts WHERE id = $1 AND user_id = $2`,
		postID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("post not found or not owned by user")
	}
	return nil
}

func (s *Store) LikePost(ctx context.Context, postID, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO likes (user_id, post_id) VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		userID, postID,
	)
	return err
}

func (s *Store) UnlikePost(ctx context.Context, postID, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM likes WHERE user_id = $1 AND post_id = $2`,
		userID, postID,
	)
	return err
}

// EncodeCursor encodes a time.Time to a base64 cursor string.
func EncodeCursor(t time.Time) string {
	return base64.URLEncoding.EncodeToString([]byte(t.Format(time.RFC3339Nano)))
}

// DecodeCursor decodes a base64 cursor string to time.Time.
func DecodeCursor(cursor string) (time.Time, error) {
	bytes, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cursor: %w", err)
	}
	return time.Parse(time.RFC3339Nano, string(bytes))
}
