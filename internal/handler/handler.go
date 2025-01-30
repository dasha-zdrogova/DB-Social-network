package handler

import (
	"net/http"
	"social-network/internal/handler/auth"
	"social-network/internal/service"
	"strconv"

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
	// регистрация
	router.HandleFunc("/auth/register", h.register).Methods("POST")
	// вход
	router.HandleFunc("/auth/login", h.login).Methods("POST")

	// ручки
	api := router.PathPrefix("/api").Subrouter()
	api.Use(h.authMiddleware)

	// посты
	api.HandleFunc("/posts", h.createPost).Methods("POST")
	api.HandleFunc("/posts/my", h.getMyPosts).Methods("GET")
	api.HandleFunc("/posts/feed", h.getFeed).Methods("GET")
	api.HandleFunc("/posts/{id}", h.getPost).Methods("GET")
	api.HandleFunc("/posts/{id}", h.updatePost).Methods("PUT")
	api.HandleFunc("/posts/{id}", h.deletePost).Methods("DELETE")
	api.HandleFunc("/posts/hashtag/{hashtag}", h.getPostsByHashtag).Methods("GET")

	// пользователи
	api.HandleFunc("/users/{username}/posts", h.getUserPosts).Methods("GET")
	api.HandleFunc("/users/follow", h.followUser).Methods("POST")
	api.HandleFunc("/users/unfollow", h.unfollowUser).Methods("POST")
	api.HandleFunc("/users/followers", h.getFollowers).Methods("GET")
	api.HandleFunc("/users/following", h.getFollowing).Methods("GET")
	api.HandleFunc("/users/mutual", h.getMutualFollows).Methods("GET")
	api.HandleFunc("/users/{username}", h.deleteAccount).Methods("DELETE")
	api.HandleFunc("/users/role", h.updateUserRole).Methods("PUT")
}

func getPaginationParams(r *http.Request) (int, int, bool) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(r.URL.Query().Get("per_page"))
	if err != nil || perPage < 1 {
		perPage = 10
	}

	desc, err := strconv.ParseBool(r.URL.Query().Get("desc"))
	if err != nil {
		desc = true
	}

	return page, perPage, desc
}

func getHashtagParam(r *http.Request) *string {
	hashtag := r.URL.Query().Get("hashtag")
	if hashtag == "" {
		return nil
	}
	return &hashtag
}
