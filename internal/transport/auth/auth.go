package transport

import (
	"context"
	"encoding/json"
	"net/http"

	errs "github.com/nik-mLb/avito_task/internal/models/errs"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	response "github.com/nik-mLb/avito_task/internal/transport/utils"
)

//go:generate mockgen -source=auth.go -destination=../../usecase/mocks/auth_usecase_mock.go -package=mocks AuthUsecase
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

// DummyLogin godoc
//	@Summary		Получить тестовый токен
//	@Description	Генерирует токен для указанной роли (для тестирования)
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.DummyLoginRequest	true	"Запрос"
//	@Success		200		{object}	dto.TokenResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Router			/dummyLogin [post]
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

// Login godoc
//	@Summary		Аутентификация пользователя
//	@Description	Вход в систему с email и паролем
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest	true	"Учетные данные"
//	@Success		200		{object}	dto.TokenResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		401		{object}	dto.ErrorResponse
//	@Router			/login [post]
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

// Register godoc
//	@Summary		Регистрация нового пользователя
//	@Description	Создает нового пользователя с указанными данными
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RegisterRequest	true	"Данные регистрации"
//	@Success		201		{object}	dto.TokenResponse
//	@Failure		400		{object}	dto.ErrorResponse
//	@Failure		500		{object}	dto.ErrorResponse
//	@Router			/register [post]
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