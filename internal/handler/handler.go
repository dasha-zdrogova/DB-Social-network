package handler

import (
	"social-network/internal/handler/auth"
	"social-network/internal/service"

	"github.com/gorilla/mux"
)

type Handler struct {
	services     *service.Services
	tokenManager *auth.TokenManager
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{
		services:     services,
		tokenManager: auth.NewTokenManager(),
	}
}

func (h *Handler) Register(router *mux.Router) {
	// Auth routes
	router.HandleFunc("/auth/register", h.register).Methods("POST")
	router.HandleFunc("/auth/login", h.login).Methods("POST")

	// API routes (with auth middleware)
	api := router.PathPrefix("/api").Subrouter()
	api.Use(h.authMiddleware)

	// Posts routes
	api.HandleFunc("/posts", h.createPost).Methods("POST")
	api.HandleFunc("/posts/{id}", h.getPost).Methods("GET")
	api.HandleFunc("/posts/{id}", h.updatePost).Methods("PUT")
	api.HandleFunc("/posts/{id}", h.deletePost).Methods("DELETE")
	api.HandleFunc("/posts/my", h.getMyPosts).Methods("GET")
	api.HandleFunc("/posts/feed", h.getFeed).Methods("GET")
	api.HandleFunc("/posts/hashtag/{hashtag}", h.getPostsByHashtag).Methods("GET")

	// Users routes
	api.HandleFunc("/users/{username}/posts", h.getUserPosts).Methods("GET")
	api.HandleFunc("/users/follow", h.followUser).Methods("POST")
	api.HandleFunc("/users/unfollow", h.unfollowUser).Methods("POST")
	api.HandleFunc("/users/followers", h.getFollowers).Methods("GET")
	api.HandleFunc("/users/following", h.getFollowing).Methods("GET")
	api.HandleFunc("/users/mutual", h.getMutualFollows).Methods("GET")
}
