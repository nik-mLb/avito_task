package transport

import (
	"context"
	"encoding/json"
	"net/http"
)

type AuthUsecase interface {
	Authenticate(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email, password, role string) (string, error)
	DummyLogin(role string) (string, error)
}

type AuthHandler struct {
	uc AuthUsecase
}

func New(uc AuthUsecase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

type (
	dummyLoginRequest struct {
		Role string `json:"role"`
	}

	loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	registerRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	authResponse struct {
		Token string `json:"token"`
	}
)

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := h.uc.DummyLogin(req.Role)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(authResponse{Token: token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := h.uc.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})

	json.NewEncoder(w).Encode(authResponse{Token: token})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, err := h.uc.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})

	json.NewEncoder(w).Encode(authResponse{Token: token})
}