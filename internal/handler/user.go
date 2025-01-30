package handler

import (
	"encoding/json"
	"net/http"
	"social-network/internal/repository/postgres"
	"social-network/internal/service"
	"strconv"

	"github.com/gorilla/mux"
)

type followRequest struct {
	Username string `json:"username"`
}

type updateRoleRequest struct {
	UserID  int    `json:"user_id"`
	NewRole string `json:"new_role"`
}

func (h *Handler) getUserPosts(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]

	page, perPage, desc := getPaginationParams(r)
	hashtag := getHashtagParam(r)

	user, err := h.services.Users.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	posts, err := h.services.Posts.GetUserPosts(user.ID, page, perPage, desc, hashtag)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *Handler) followUser(w http.ResponseWriter, r *http.Request) {
	followerID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var req followRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userToFollow, err := h.services.Users.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if err := h.services.Users.Follow(followerID, userToFollow.ID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) unfollowUser(w http.ResponseWriter, r *http.Request) {
	followerID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var req followRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userToUnfollow, err := h.services.Users.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	if err := h.services.Users.Unfollow(followerID, userToUnfollow.ID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) getFollowers(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	page, perPage, _ := getPaginationParams(r)

	followers, err := h.services.Users.GetFollowers(userID, page, perPage)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followers)
}

func (h *Handler) getFollowing(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	page, perPage, _ := getPaginationParams(r)

	following, err := h.services.Users.GetFollowing(userID, page, perPage)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(following)
}

func (h *Handler) getMutualFollows(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	page, perPage, _ := getPaginationParams(r)

	mutuals, err := h.services.Users.GetMutualFollows(userID, page, perPage)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mutuals)
}

func (h *Handler) deleteAccount(w http.ResponseWriter, r *http.Request) {
	withPosts, err := strconv.ParseBool(r.URL.Query().Get("with_posts"))
	if err != nil {
		withPosts = false
	}
	username := mux.Vars(r)["username"]
	user, err := h.services.Users.GetUserByUsername(username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	userID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	token := r.Header.Get("Authorization")

	if err := h.services.Users.DeleteAccount(userID, user.ID, withPosts); err != nil {
		if err == service.ErrNotAdmin {
			http.Error(w, "do not have access", http.StatusForbidden)
			return
		}
		http.Error(w, "failed to delete account", http.StatusInternalServerError)
		return
	}

	h.tokenManager.RemoveToken(token)

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) updateUserRole(w http.ResponseWriter, r *http.Request) {
	adminID, err := h.getUserIDFromContext(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if adminID == req.UserID {
		http.Error(w, "cannot change own role", http.StatusBadRequest)
		return
	}

	if err := h.services.Users.UpdateRole(adminID, req.UserID, req.NewRole); err != nil {
		switch err {
		case postgres.ErrUserNotFound:
			http.Error(w, "user not found", http.StatusNotFound)
		case service.ErrNotAdmin:
			http.Error(w, "not admin", http.StatusForbidden)
		case service.ErrUnknownRole:
			http.Error(w, "unknown role", http.StatusBadRequest)
		default:
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
