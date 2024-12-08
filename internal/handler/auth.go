package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"social-network/internal/repository/postgres"
	"social-network/internal/service"
	"social-network/internal/models"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (h *Handler) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "missing auth token", http.StatusUnauthorized)
			return
		}

		userID, err := h.tokenManager.ValidateToken(token)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if len(req.Username) < 3 || len(req.Password) < 6 {
		http.Error(w, "username or password too short", http.StatusBadRequest)
		return
	}

	if err := h.services.Users.Register(req.Username, req.Password); err != nil {
		if err == postgres.ErrUserExists {
			http.Error(w, "username already taken", http.StatusConflict)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate credentials
	user, err := h.services.Users.ValidateUser(req.Username, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Generate and store token
	token := h.tokenManager.GenerateToken()
	h.tokenManager.AddToken(token, user.ID)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Return response
	json.NewEncoder(w).Encode(authResponse{
		Token: token,
	})
}
