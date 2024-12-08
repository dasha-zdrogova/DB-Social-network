package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

type createPostRequest struct {
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
	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	if len(req.Content) < 1 || len(req.Content) > 2048 {
		http.Error(w, "content must be between 1 and 2048 characters", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
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
