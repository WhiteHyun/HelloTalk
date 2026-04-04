package feed

import (
	"encoding/json"
	"net/http"
	"strconv"
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
	mux.Handle("POST /api/v1/posts", authMiddleware(http.HandlerFunc(h.handleCreatePost)))
	mux.Handle("GET /api/v1/posts", authMiddleware(http.HandlerFunc(h.handleGetFeed)))
	mux.Handle("GET /api/v1/posts/{id}", authMiddleware(http.HandlerFunc(h.handleGetPost)))
	mux.Handle("PUT /api/v1/posts/{id}", authMiddleware(http.HandlerFunc(h.handleUpdatePost)))
	mux.Handle("DELETE /api/v1/posts/{id}", authMiddleware(http.HandlerFunc(h.handleDeletePost)))
	mux.Handle("POST /api/v1/posts/{id}/like", authMiddleware(http.HandlerFunc(h.handleLikePost)))
	mux.Handle("DELETE /api/v1/posts/{id}/like", authMiddleware(http.HandlerFunc(h.handleUnlikePost)))
}

func (h *Handler) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "잘못된 요청 형식입니다")
		return
	}

	if req.Content == "" {
		response.Error(w, http.StatusBadRequest, "EMPTY_CONTENT", "내용을 입력해주세요")
		return
	}

	post, err := h.store.CreatePost(r.Context(), userID, req.Content)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "게시글 작성에 실패했습니다")
		return
	}

	// 새로 생성된 게시글을 풀 응답으로 반환
	postResp, err := h.store.GetPost(r.Context(), post.ID, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "게시글 조회에 실패했습니다")
		return
	}

	response.JSON(w, http.StatusCreated, postResp)
}

func (h *Handler) handleGetFeed(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	// 커서 파싱
	cursor := time.Now().Add(time.Second) // 기본값: 미래 시간 (모든 게시글 포함)
	if cursorParam := r.URL.Query().Get("cursor"); cursorParam != "" {
		parsed, err := DecodeCursor(cursorParam)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "INVALID_CURSOR", "잘못된 커서입니다")
			return
		}
		cursor = parsed
	}

	// limit 파싱
	limit := 20
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		parsed, err := strconv.Atoi(limitParam)
		if err != nil || parsed < 1 || parsed > 100 {
			response.Error(w, http.StatusBadRequest, "INVALID_LIMIT", "limit은 1~100 사이여야 합니다")
			return
		}
		limit = parsed
	}

	// limit+1개를 요청해서 다음 페이지 존재 여부를 판단
	posts, err := h.store.GetFeed(r.Context(), userID, cursor, limit+1)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "피드 조회에 실패했습니다")
		return
	}

	var nextCursor *string
	if len(posts) > limit {
		// 다음 페이지가 있다
		lastPost := posts[limit-1]
		parsed, _ := time.Parse(time.RFC3339, lastPost.CreatedAt)
		encoded := EncodeCursor(parsed)
		nextCursor = &encoded
		posts = posts[:limit]
	}

	if posts == nil {
		posts = []PostResponse{}
	}

	response.JSON(w, http.StatusOK, FeedResponse{
		Posts:      posts,
		NextCursor: nextCursor,
	})
}

func (h *Handler) handleGetPost(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	postID := r.PathValue("id")

	post, err := h.store.GetPost(r.Context(), postID, userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, "POST_NOT_FOUND", "게시글을 찾을 수 없습니다")
		return
	}

	response.JSON(w, http.StatusOK, post)
}

func (h *Handler) handleUpdatePost(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	postID := r.PathValue("id")

	var req UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "잘못된 요청 형식입니다")
		return
	}

	if req.Content == "" {
		response.Error(w, http.StatusBadRequest, "EMPTY_CONTENT", "내용을 입력해주세요")
		return
	}

	_, err := h.store.UpdatePost(r.Context(), postID, userID, req.Content)
	if err != nil {
		response.Error(w, http.StatusNotFound, "POST_NOT_FOUND", "게시글을 찾을 수 없거나 권한이 없습니다")
		return
	}

	// 업데이트된 게시글을 풀 응답으로 반환
	post, err := h.store.GetPost(r.Context(), postID, userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "게시글 조회에 실패했습니다")
		return
	}

	response.JSON(w, http.StatusOK, post)
}

func (h *Handler) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	postID := r.PathValue("id")

	if err := h.store.DeletePost(r.Context(), postID, userID); err != nil {
		response.Error(w, http.StatusNotFound, "POST_NOT_FOUND", "게시글을 찾을 수 없거나 권한이 없습니다")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleLikePost(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	postID := r.PathValue("id")

	if err := h.store.LikePost(r.Context(), postID, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "좋아요 처리에 실패했습니다")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleUnlikePost(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	postID := r.PathValue("id")

	if err := h.store.UnlikePost(r.Context(), postID, userID); err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "좋아요 취소에 실패했습니다")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
