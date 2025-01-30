package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

type tokenData struct {
	UserID    int
	ExpiresAt time.Time
}

type TokenManager struct {
	tokens map[string]tokenData
	mu     sync.RWMutex
}

func NewTokenManager() *TokenManager {
	tm := &TokenManager{
		tokens: make(map[string]tokenData),
	}
	go tm.startCleanup()
	return tm
}

func (tm *TokenManager) AddToken(token string, userID int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.tokens[token] = tokenData{
		UserID:    userID,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
}

func (tm *TokenManager) ValidateToken(token string) (int, error) {
	tm.mu.RLock()
	data, exists := tm.tokens[token]
	tm.mu.RUnlock()

	if !exists || time.Now().After(data.ExpiresAt) {
		return 0, errors.New("invalid token")

	}

	data.ExpiresAt = time.Now().Add(30 * time.Minute)
	tm.mu.Lock()
	tm.tokens[token] = data
	tm.mu.Unlock()

	return data.UserID, nil
}

func (tm *TokenManager) startCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		tm.cleanup()
	}
}

func (tm *TokenManager) cleanup() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	now := time.Now()
	for token, data := range tm.tokens {
		if now.After(data.ExpiresAt) {
			delete(tm.tokens, token)
		}
	}
}

func (*TokenManager) GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (tm *TokenManager) RemoveToken(token string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.tokens, token)
}
