package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	infraauth "github.com/GarikMirzoyan/gophermart/internal/infrastructure/auth"
	"github.com/GarikMirzoyan/gophermart/internal/usecase/auth"
)

type AuthHandler struct {
	AuthService *auth.Service
	JWTManager  *infraauth.JWTManager
}

func NewAuthHandler(authService *auth.Service, jwtManager *infraauth.JWTManager) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
		JWTManager:  jwtManager,
	}
}

type credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.AuthService.Register(r.Context(), creds.Login, creds.Password)
	if err != nil {
		if errors.Is(err, auth.ErrLoginTaken) {
			http.Error(w, "login already used", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("server error: %v", err), http.StatusInternalServerError)
		return
	}

	token, _ := h.JWTManager.Generate(int(user.ID))
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := h.AuthService.Authenticate(r.Context(), creds.Login, creds.Password)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	token, _ := h.JWTManager.Generate(int(user.ID))
	w.Header().Set("Authorization", "Bearer "+token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]string{
		"token": token,
	})
}
