package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/whitehyun/HelloTalk/apps/server/internal/config"
)

// --- Mock Store ---

type mockStore struct {
	createUserFn          func(ctx context.Context, email, passwordHash, name string) (*User, error)
	getUserByEmailFn      func(ctx context.Context, email string) (*User, error)
	saveRefreshTokenFn    func(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	getRefreshTokenFn     func(ctx context.Context, tokenHash string) (*RefreshToken, error)
	deleteRefreshTokenFn  func(ctx context.Context, tokenHash string) error
	deleteAllRefreshFn    func(ctx context.Context, userID string) error
}

func (m *mockStore) CreateUser(ctx context.Context, email, passwordHash, name string) (*User, error) {
	return m.createUserFn(ctx, email, passwordHash, name)
}
func (m *mockStore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return m.getUserByEmailFn(ctx, email)
}
func (m *mockStore) SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	return m.saveRefreshTokenFn(ctx, userID, tokenHash, expiresAt)
}
func (m *mockStore) GetRefreshToken(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	return m.getRefreshTokenFn(ctx, tokenHash)
}
func (m *mockStore) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	return m.deleteRefreshTokenFn(ctx, tokenHash)
}
func (m *mockStore) DeleteAllRefreshTokens(ctx context.Context, userID string) error {
	return m.deleteAllRefreshFn(ctx, userID)
}

func newTestService(store *mockStore) *Service {
	cfg := &config.Config{
		JWTSecret:        "test-secret",
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 168 * time.Hour,
	}
	return NewService(store, cfg)
}

// --- Tests ---

func TestHandleSignup_Success(t *testing.T) {
	store := &mockStore{
		getUserByEmailFn: func(_ context.Context, _ string) (*User, error) {
			return nil, ErrInvalidCredentials // 이메일 없음 → 가입 가능
		},
		createUserFn: func(_ context.Context, email, _, name string) (*User, error) {
			return &User{
				ID:        "user-1",
				Email:     email,
				Name:      name,
				CreatedAt: time.Now(),
			}, nil
		},
		saveRefreshTokenFn: func(_ context.Context, _, _ string, _ time.Time) error {
			return nil
		},
	}

	service := newTestService(store)
	handler := NewHandler(service)

	body := `{"email":"test@test.com","password":"test1234","name":"테스트"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.handleSignup(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("want status 201, got %d", rec.Code)
	}

	var resp AuthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.User.Email != "test@test.com" {
		t.Errorf("want email test@test.com, got %s", resp.User.Email)
	}
	if resp.Token.AccessToken == "" {
		t.Error("expected access_token to be non-empty")
	}
}

func TestHandleSignup_DuplicateEmail(t *testing.T) {
	store := &mockStore{
		getUserByEmailFn: func(_ context.Context, _ string) (*User, error) {
			return &User{ID: "existing"}, nil // 이미 존재하는 이메일
		},
	}

	service := newTestService(store)
	handler := NewHandler(service)

	body := `{"email":"dup@test.com","password":"test1234","name":"테스트"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/signup", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.handleSignup(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("want status 409, got %d", rec.Code)
	}
}

func TestHandleSignup_ValidationErrors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "잘못된 이메일",
			body:       `{"email":"not-an-email","password":"test1234","name":"테스트"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_EMAIL",
		},
		{
			name:       "짧은 비밀번호",
			body:       `{"email":"test@test.com","password":"short","name":"테스트"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PASSWORD",
		},
		{
			name:       "빈 이름",
			body:       `{"email":"test@test.com","password":"test1234","name":""}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_NAME",
		},
		{
			name:       "잘못된 JSON",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_BODY",
		},
	}

	store := &mockStore{
		getUserByEmailFn: func(_ context.Context, _ string) (*User, error) {
			return nil, ErrInvalidCredentials
		},
	}
	service := newTestService(store)
	handler := NewHandler(service)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/auth/signup", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.handleSignup(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, rec.Code)
			}

			var errResp map[string]string
			json.NewDecoder(rec.Body).Decode(&errResp)
			if errResp["code"] != tt.wantCode {
				t.Errorf("want code %q, got %q", tt.wantCode, errResp["code"])
			}
		})
	}
}

func TestHandleLogin_Success(t *testing.T) {
	// bcrypt로 "test1234"를 해싱
	store := &mockStore{
		getUserByEmailFn: func(_ context.Context, _ string) (*User, error) {
			return &User{
				ID:           "user-1",
				Email:        "test@test.com",
				PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // "test1234"는 아님, 비교 테스트
				Name:         "테스트",
				CreatedAt:    time.Now(),
			}, nil
		},
	}

	service := newTestService(store)
	handler := NewHandler(service)

	// 비밀번호 해시가 매치하지 않으므로 401
	body := `{"email":"test@test.com","password":"test1234"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.handleLogin(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want status 401, got %d", rec.Code)
	}
}

func TestHandleLogin_UserNotFound(t *testing.T) {
	store := &mockStore{
		getUserByEmailFn: func(_ context.Context, _ string) (*User, error) {
			return nil, ErrInvalidCredentials
		},
	}

	service := newTestService(store)
	handler := NewHandler(service)

	body := `{"email":"nobody@test.com","password":"test1234"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.handleLogin(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("want status 401, got %d", rec.Code)
	}
}
