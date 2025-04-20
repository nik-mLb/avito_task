package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	auth "github.com/nik-mLb/avito_task/internal/transport/auth"
	"github.com/nik-mLb/avito_task/internal/transport/dto"
	"github.com/nik-mLb/avito_task/internal/usecase/mocks"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
)

func TestAuthHandler_DummyLogin(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockReturn     string
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "successful dummy login",
			requestBody:    `{"role": "admin"}`,
			mockReturn:     "dummy_token",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"token":"dummy_token"}`,
		},
		{
			name:           "empty role",
			requestBody:    `{"role": ""}`,
			mockReturn:     "",
			mockError:      errors.New("role required"),
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message":"Failed to generate token"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mocks.NewMockAuthUsecase(ctrl)
			h := auth.New(mockUsecase)

			if tt.requestBody != "" {
				var req dto.DummyLoginRequest
				if err := json.Unmarshal([]byte(tt.requestBody), &req); err == nil {
					mockUsecase.EXPECT().
						DummyLogin(req.Role).
						Return(tt.mockReturn, tt.mockError).
						Times(1)
				}
			}

			req := httptest.NewRequest("POST", "/dummy-login", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			h.DummyLogin(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != tt.expectedBody {
				t.Errorf("expected body %s, got %s", tt.expectedBody, body)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockReturn     string
		mockError      error
		expectedStatus int
		expectedBody   string
		expectCookie   bool
	}{
		{
			name:           "successful login",
			requestBody:    `{"login": "test@example.com", "password": "password"}`,
			mockReturn:     "auth_token",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"token":"auth_token"}`,
			expectCookie:   true,
		},
		{
			name:           "invalid credentials",
			requestBody:    `{"login": "test@example.com", "password": "wrong"}`,
			mockReturn:     "",
			mockError:      errors.New("invalid credentials"),
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"message":"Incorrect data"}`,
			expectCookie:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUsecase := mocks.NewMockAuthUsecase(ctrl)
			h := auth.New(mockUsecase)

			// Only set up mock expectation if we expect the usecase to be called
			if tt.expectedStatus != http.StatusBadRequest {
				var req dto.LoginRequest
				if err := json.Unmarshal([]byte(tt.requestBody), &req); err == nil {
					mockUsecase.EXPECT().
						Authenticate(gomock.Any(), req.Email, req.Password).
						Return(tt.mockReturn, tt.mockError).
						Times(1)
				}
			}

			req := httptest.NewRequest("POST", "/login", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			h.Login(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != tt.expectedBody {
				t.Errorf("expected body %s, got %s", tt.expectedBody, body)
			}

			cookies := resp.Cookies()
			if tt.expectCookie && len(cookies) == 0 {
				t.Error("expected cookie to be set")
			}
			if !tt.expectCookie && len(cookies) > 0 {
				t.Error("expected no cookie to be set")
			}
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockReturn     string
		mockError      error
		expectedStatus int
		expectedBody   string
		expectCookie   bool
	}{
		{
			name:           "successful registration",
			requestBody:    `{"login": "new@example.com", "password": "password", "role": "user"}`,
			mockReturn:     "reg_token",
			mockError:      nil,
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"token":"reg_token"}`,
			expectCookie:   true,
		},
		{
			name:           "registration error",
			requestBody:    `{"login": "new@example.com", "password": "password", "role": "user"}`,
			mockReturn:     "",
			mockError:      errors.New("registration failed"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"message":"Failed registration"}`,
			expectCookie:   false,
		},
		{
			name:           "invalid role",
			requestBody:    `{"login": "new@example.com", "password": "password", "role": "invalid"}`,
			mockReturn:     "",
			mockError:      errs.ErrRoleNotAllowed,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"message":"Invalid role"}`,
			expectCookie:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
	
			mockUsecase := mocks.NewMockAuthUsecase(ctrl)
			h := auth.New(mockUsecase)
	
			// Декодируем тело запроса для проверки параметров
			var reqBody dto.RegisterRequest
			json.Unmarshal([]byte(tt.requestBody), &reqBody)
	
			// Устанавливаем ожидания для мока, если usecase должен вызываться
			if tt.mockError != nil || tt.mockReturn != "" {
				mockUsecase.EXPECT().
					Register(gomock.Any(), reqBody.Email, reqBody.Password, reqBody.Role).
					Return(tt.mockReturn, tt.mockError).
					Times(1)
			}
	
			// Создаем HTTP запрос
			req := httptest.NewRequest("POST", "/register", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()
	
			h.Register(w, req)
	
			// Проверки результатов
			resp := w.Result()
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
	
			body := strings.TrimSpace(w.Body.String())
			if body != tt.expectedBody {
				t.Errorf("expected body %s, got %s", tt.expectedBody, body)
			}
	
			cookies := resp.Cookies()
			if tt.expectCookie && len(cookies) == 0 {
				t.Error("expected cookie to be set")
			}
			if !tt.expectCookie && len(cookies) > 0 {
				t.Error("expected no cookie to be set")
			}
		})
	}
}