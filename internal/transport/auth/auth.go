package transport

import (
	"context"
	"encoding/json"
	"net/http"

	errs "github.com/nik-mLb/avito_task/internal/models/errs"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	response "github.com/nik-mLb/avito_task/internal/transport/utils"
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

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var req dto.DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	token, err := h.uc.DummyLogin(req.Role)
	if err != nil {
		response.SendError(r.Context(), w, http.StatusBadRequest, "Failed to generate token")
		return
	}

	responseTok := dto.TokenResponse{Token: token}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, responseTok)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	token, err := h.uc.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		response.SendError(r.Context(), w, http.StatusUnauthorized, "Incorrect data")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})

	respTok := dto.TokenResponse{Token: token}

	response.SendJSONResponse(r.Context(), w, http.StatusOK, respTok)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid request")
		return
	}

	token, err := h.uc.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		switch err {
		case errs.ErrRoleNotAllowed:
			response.SendError(r.Context(), w, http.StatusBadRequest, "Invalid role")
		default:
			response.SendError(r.Context(), w, http.StatusInternalServerError, "Failed registration")
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})

	respTok := dto.TokenResponse{Token: token}

	response.SendJSONResponse(r.Context(), w, http.StatusCreated, respTok)
}