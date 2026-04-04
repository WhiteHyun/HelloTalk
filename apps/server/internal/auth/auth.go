package auth

import (
	"context"
	"fmt"
	"net/mail"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/whitehyun/HelloTalk/apps/server/internal/config"
	"github.com/whitehyun/HelloTalk/apps/server/internal/token"
)

// --- Models ---

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	AvatarURL    *string   `json:"avatar_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"-"`
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// --- Request/Response ---

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url"`
	CreatedAt string  `json:"created_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type AuthResponse struct {
	User  UserResponse  `json:"user"`
	Token TokenResponse `json:"token"`
}

// --- Errors ---

var (
	ErrInvalidEmail    = fmt.Errorf("유효하지 않은 이메일 형식입니다")
	ErrShortPassword   = fmt.Errorf("비밀번호는 8자 이상이어야 합니다")
	ErrEmptyName       = fmt.Errorf("이름은 필수입니다")
	ErrEmailExists     = fmt.Errorf("이미 존재하는 이메일입니다")
	ErrInvalidCredentials = fmt.Errorf("이메일 또는 비밀번호가 올바르지 않습니다")
	ErrInvalidRefreshToken = fmt.Errorf("유효하지 않은 리프레시 토큰입니다")
	ErrExpiredRefreshToken = fmt.Errorf("만료된 리프레시 토큰입니다")
)

// --- Store Interface ---

type StoreInterface interface {
	CreateUser(ctx context.Context, email, passwordHash, name string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
	DeleteAllRefreshTokens(ctx context.Context, userID string) error
}

// --- Service ---

type Service struct {
	store StoreInterface
	cfg   *config.Config
}

func NewService(store StoreInterface, cfg *config.Config) *Service {
	return &Service{store: store, cfg: cfg}
}

func (s *Service) Signup(ctx context.Context, req SignupRequest) (*AuthResponse, error) {
	// 유효성 검증
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, ErrInvalidEmail
	}
	if len(req.Password) < 8 {
		return nil, ErrShortPassword
	}
	if req.Name == "" {
		return nil, ErrEmptyName
	}

	// 이메일 중복 확인
	existing, _ := s.store.GetUserByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrEmailExists
	}

	// 비밀번호 해싱
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 유저 생성
	user, err := s.store.CreateUser(ctx, req.Email, string(hash), req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 토큰 발급
	return s.issueTokens(ctx, user)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*TokenResponse, error) {
	// 해시로 DB 조회
	hash := token.HashToken(req.RefreshToken)
	stored, err := s.store.GetRefreshToken(ctx, hash)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// 만료 확인
	if time.Now().After(stored.ExpiresAt) {
		s.store.DeleteRefreshToken(ctx, hash)
		return nil, ErrExpiredRefreshToken
	}

	// 기존 토큰 삭제 (rotation)
	s.store.DeleteRefreshToken(ctx, hash)

	// 새 토큰 발급
	accessToken, err := token.GenerateAccess(stored.UserID, s.cfg.JWTSecret, s.cfg.JWTAccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	plainRefresh, refreshHash, err := token.GenerateRefresh()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresAt := time.Now().Add(s.cfg.JWTRefreshExpiry)
	if err := s.store.SaveRefreshToken(ctx, stored.UserID, refreshHash, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: plainRefresh,
		ExpiresIn:    int(s.cfg.JWTAccessExpiry.Seconds()),
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	hash := token.HashToken(refreshToken)
	return s.store.DeleteRefreshToken(ctx, hash)
}

// issueTokens generates access + refresh tokens for a user.
func (s *Service) issueTokens(ctx context.Context, user *User) (*AuthResponse, error) {
	accessToken, err := token.GenerateAccess(user.ID, s.cfg.JWTSecret, s.cfg.JWTAccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	plainRefresh, refreshHash, err := token.GenerateRefresh()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresAt := time.Now().Add(s.cfg.JWTRefreshExpiry)
	if err := s.store.SaveRefreshToken(ctx, user.ID, refreshHash, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &AuthResponse{
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
		Token: TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: plainRefresh,
			ExpiresIn:    int(s.cfg.JWTAccessExpiry.Seconds()),
		},
	}, nil
}
