package feed

import "time"

type Post struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type PostResponse struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	UserName  string  `json:"user_name"`
	Content   string  `json:"content"`
	LikeCount int     `json:"like_count"`
	Liked     bool    `json:"liked"`
	CreatedAt string  `json:"created_at"`
}

type FeedResponse struct {
	Posts      []PostResponse `json:"posts"`
	NextCursor *string        `json:"next_cursor"`
}

type CreatePostRequest struct {
	Content string `json:"content"`
}

type UpdatePostRequest struct {
	Content string `json:"content"`
}
