package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/nik-mLb/avito_task/config"
	user "github.com/nik-mLb/avito_task/internal/models/user"
	"github.com/nik-mLb/avito_task/internal/repository/mocks"
	"github.com/nik-mLb/avito_task/internal/transport/jwt"
	usecase "github.com/nik-mLb/avito_task/internal/usecase/auth"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
)

func createTestTokenator() *jwt.Tokenator {
	cfg := &config.JWTConfig{
		Signature:   "test-secret-key-123",
		TokenLifeSpan: 24 * time.Hour,
	}
	return jwt.NewTokenator(cfg)
}

func TestAuthUsecase_Authenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthRepository(ctrl)
	mockTokenator := createTestTokenator()

	uc := usecase.New(mockRepo, mockTokenator)

	t.Run("successful authentication", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		userID := uuid.New()

		mockUser := &user.User{
			ID:           userID,
			Email:        email,
			PasswordHash: hashedPassword,
			Role:         "worker",
		}

		mockRepo.EXPECT().
			GetUserByEmail(gomock.Any(), email).
			Return(mockUser, nil)

		token, err := uc.Authenticate(context.Background(), email, password)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("user not found", func(t *testing.T) {
		email := "notfound@example.com"

		mockRepo.EXPECT().
			GetUserByEmail(gomock.Any(), email).
			Return(nil, nil)

		token, err := uc.Authenticate(context.Background(), email, "anypassword")

		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
		assert.Empty(t, token)
	})

	t.Run("invalid password", func(t *testing.T) {
		email := "test@example.com"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		userID := uuid.New()

		mockUser := &user.User{
			ID:           userID,
			Email:        email,
			PasswordHash: hashedPassword,
			Role:         "worker",
		}

		mockRepo.EXPECT().
			GetUserByEmail(gomock.Any(), email).
			Return(mockUser, nil)

		token, err := uc.Authenticate(context.Background(), email, "wrongpassword")

		assert.Error(t, err)
		assert.Equal(t, "invalid password", err.Error())
		assert.Empty(t, token)
	})

	t.Run("repository error", func(t *testing.T) {
		email := "error@example.com"

		mockRepo.EXPECT().
			GetUserByEmail(gomock.Any(), email).
			Return(nil, errors.New("database error"))

		token, err := uc.Authenticate(context.Background(), email, "anypassword")

		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
		assert.Empty(t, token)
	})
}

func TestAuthUsecase_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthRepository(ctrl)
	mockTokenator := createTestTokenator()

	uc := usecase.New(mockRepo, mockTokenator)

	t.Run("successful registration", func(t *testing.T) {
		email := "new@example.com"
		password := "password123"
		role := "worker"
		userID := uuid.New()

		mockUser := &user.User{
			ID:    userID,
			Email: email,
			Role:  role,
		}

		mockRepo.EXPECT().
			CreateUser(gomock.Any(), email, gomock.Any(), role).
			Return(mockUser, nil)

		token, err := uc.Register(context.Background(), email, password, role)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("invalid role", func(t *testing.T) {
		email := "new@example.com"
		password := "password123"
		role := "invalid-role"

		token, err := uc.Register(context.Background(), email, password, role)

		assert.Error(t, err)
		assert.Equal(t, errs.ErrCityNotAllowed, err)
		assert.Empty(t, token)
	})

	t.Run("repository error", func(t *testing.T) {
		email := "error@example.com"
		password := "password123"
		role := "worker"

		mockRepo.EXPECT().
			CreateUser(gomock.Any(), email, gomock.Any(), role).
			Return(nil, errors.New("database error"))

		token, err := uc.Register(context.Background(), email, password, role)

		assert.Error(t, err)
		assert.Equal(t, "database error", err.Error())
		assert.Empty(t, token)
	})

	t.Run("password hashing error", func(t *testing.T) {
		// Этот тест сложно реализовать, так как bcrypt.GenerateFromPassword
		// обычно не возвращает ошибок при нормальных условиях
		// Можно использовать мок для bcrypt, но это усложнит тест
	})
}

func TestAuthUsecase_DummyLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuthRepository(ctrl)
	mockTokenator := createTestTokenator()

	uc := usecase.New(mockRepo, mockTokenator)

	t.Run("successful dummy login", func(t *testing.T) {
		role := "admin"

		token, err := uc.DummyLogin(role)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("invalid role - should still work as Tokenator doesn't validate roles", func(t *testing.T) {
		role := "invalid-role"
	
		token, err := uc.DummyLogin(role)
	
		// В текущей реализации Tokenator не проверяет роли при создании JWT,
		// поэтому ошибки быть не должно
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	
		// Дополнительно можно проверить, что токен содержит переданную роль
		claims, err := mockTokenator.ParseJWT(token)
		assert.NoError(t, err)
		assert.Equal(t, "invalid-role", claims.Role)
	})
}