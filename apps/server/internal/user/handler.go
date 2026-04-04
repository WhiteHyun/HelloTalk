package user

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/whitehyun/HelloTalk/apps/server/internal/middleware"
	"github.com/whitehyun/HelloTalk/apps/server/internal/response"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/users/me", authMiddleware(http.HandlerFunc(h.handleGetMe)))
	mux.Handle("PUT /api/v1/users/me", authMiddleware(http.HandlerFunc(h.handleUpdateMe)))
	mux.Handle("POST /api/v1/users/me/avatar", authMiddleware(http.HandlerFunc(h.handleUploadAvatar)))
	mux.Handle("GET /api/v1/users/{id}", authMiddleware(http.HandlerFunc(h.handleGetUser)))
}

func (h *Handler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	u, err := h.store.GetByID(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "유저를 찾을 수 없습니다")
		return
	}

	response.JSON(w, http.StatusOK, toResponse(u))
}

func (h *Handler) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "잘못된 요청 형식입니다")
		return
	}

	if req.Name != nil && *req.Name == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_NAME", "이름은 비어있을 수 없습니다")
		return
	}

	u, err := h.store.Update(r.Context(), userID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "프로필 수정에 실패했습니다")
		return
	}

	response.JSON(w, http.StatusOK, toResponse(u))
}

func (h *Handler) handleUploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	// 최대 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "FILE_TOO_LARGE", "파일 크기는 10MB 이하여야 합니다")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "MISSING_FILE", "avatar 파일이 필요합니다")
		return
	}
	defer file.Close()

	// 확장자 확인
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		response.Error(w, http.StatusBadRequest, "INVALID_FILE_TYPE", "jpg, jpeg, png 파일만 허용됩니다")
		return
	}

	// 저장 디렉토리 생성
	uploadDir := "./uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "파일 저장에 실패했습니다")
		return
	}

	// 파일명: {userID}_{timestamp}{ext}
	filename := fmt.Sprintf("%s_%d%s", userID, time.Now().Unix(), ext)
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "파일 저장에 실패했습니다")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "파일 저장에 실패했습니다")
		return
	}

	// DB 업데이트
	avatarURL := fmt.Sprintf("/uploads/avatars/%s", filename)
	u, err := h.store.UpdateAvatar(r.Context(), userID, avatarURL)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "아바타 업데이트에 실패했습니다")
		return
	}

	response.JSON(w, http.StatusOK, toResponse(u))
}

func (h *Handler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	u, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "유저를 찾을 수 없습니다")
		return
	}

	response.JSON(w, http.StatusOK, toResponse(u))
}
