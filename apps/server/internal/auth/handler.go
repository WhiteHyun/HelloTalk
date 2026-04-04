package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/whitehyun/HelloTalk/apps/server/internal/middleware"
	"github.com/whitehyun/HelloTalk/apps/server/internal/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	mux.HandleFunc("POST /api/v1/auth/signup", h.handleSignup)
	mux.HandleFunc("POST /api/v1/auth/login", h.handleLogin)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.handleRefresh)

	// logout은 인증이 필요하다
	mux.Handle("POST /api/v1/auth/logout", authMiddleware(http.HandlerFunc(h.handleLogout)))
}

func (h *Handler) handleSignup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "잘못된 요청 형식입니다")
		return
	}

	result, err := h.service.Signup(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidEmail):
			response.Error(w, http.StatusBadRequest, "INVALID_EMAIL", err.Error())
		case errors.Is(err, ErrShortPassword):
			response.Error(w, http.StatusBadRequest, "INVALID_PASSWORD", err.Error())
		case errors.Is(err, ErrEmptyName):
			response.Error(w, http.StatusBadRequest, "INVALID_NAME", err.Error())
		case errors.Is(err, ErrEmailExists):
			response.Error(w, http.StatusConflict, "EMAIL_EXISTS", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "서버 내부 오류가 발생했습니다")
		}
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "잘못된 요청 형식입니다")
		return
	}

	result, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
		} else {
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "서버 내부 오류가 발생했습니다")
		}
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "잘못된 요청 형식입니다")
		return
	}

	result, err := h.service.Refresh(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRefreshToken), errors.Is(err, ErrExpiredRefreshToken):
			response.Error(w, http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "서버 내부 오류가 발생했습니다")
		}
		return
	}

	response.JSON(w, http.StatusOK, result)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "잘못된 요청 형식입니다")
		return
	}

	userID := middleware.GetUserID(r)
	_ = userID // logout에서는 refresh token 기반으로 삭제하므로 userID는 추가 검증용

	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "로그아웃 처리 중 오류가 발생했습니다")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
