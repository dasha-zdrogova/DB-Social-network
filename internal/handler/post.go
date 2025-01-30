package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"social-network/internal/repository/postgres"
	"social-network/internal/service"
)

type createPostRequest struct {
	Content string `json:"content"`
}

type updatePostRequest struct {
	Content string `json:"content"`
}

type postResponse struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Author    string    `json:"author"`
	Hashtags  []string  `json:"hashtags,omitempty"`
}

func (h *Handler) createPost(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Content) < 1 || len(req.Content) > 280 {
		http.Error(w, "content must be between 1 and 280 characters", http.StatusBadRequest)
		return
	}

	post, err := h.services.Posts.Create(userID, req.Content)
	if err != nil {
		http.Error(w, "failed to create post", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(postResponse{
		ID:        post.ID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
		Author:    post.Author.Username,
		Hashtags:  post.Hashtags,
	})
}

func (h *Handler) getPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	post, err := h.services.Posts.GetByID(postID)
	if err != nil {
		if err == postgres.ErrPostNotFound {
			http.Error(w, "post not found", http.StatusNotFound)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(postResponse{
		ID:        post.ID,
		Content:   post.Content,
		CreatedAt: post.CreatedAt,
		Author:    post.Author.Username,
		Hashtags:  post.Hashtags,
	})
}

func (h *Handler) updatePost(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	var req updatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Content) < 1 || len(req.Content) > 280 {
		http.Error(w, "content must be between 1 and 280 characters", http.StatusBadRequest)
		return
	}

	if err := h.services.Posts.Update(postID, userID, req.Content); err != nil {
		if err == service.ErrNotYourPost {
			http.Error(w, "do not have access rights", http.StatusForbidden)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) deletePost(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	if err := h.services.Posts.Delete(postID, userID); err != nil {
		if err == service.ErrNotYourPost {
			http.Error(w, "do not have access rights", http.StatusForbidden)
			return
		}
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) getMyPosts(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	page, perPage, desc := getPaginationParams(r)

	posts, err := h.services.Posts.GetMyPosts(userID, page, perPage, desc)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *Handler) getFeed(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	page, perPage, desc := getPaginationParams(r)
	hashtag := getHashtagParam(r)

	posts, err := h.services.Posts.GetFeed(userID, page, perPage, desc, hashtag)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *Handler) getPostsByHashtag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	hashtag := vars["hashtag"]
	if hashtag == "" {
		http.Error(w, "hashtag is required", http.StatusBadRequest)
		return
	}

	page, perPage, desc := getPaginationParams(r)

	posts, err := h.services.Posts.GetPostsByHashtag(hashtag, page, perPage, desc)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}
