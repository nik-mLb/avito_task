package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
	models "github.com/nik-mLb/avito_task/internal/models/user"
	"github.com/nik-mLb/avito_task/internal/transport/jwt"
	"github.com/nik-mLb/avito_task/internal/transport/middleware/logctx"
	"golang.org/x/crypto/bcrypt"
)

var (
	allowedRoles = map[string]bool{
		"worker":          true,
		"admin": 		   true,
	}
)

//go:generate mockgen -source=auth.go -destination=../../repository/mocks/auth_repository_mock.go -package=mocks AuthRepository
type AuthRepository interface {
	CreateUser(ctx context.Context, email string, passwordHash []byte, role string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type AuthUsecase struct {
	repo      AuthRepository
	tokenator *jwt.Tokenator
}

func New(repo AuthRepository, tokenator *jwt.Tokenator) *AuthUsecase {
	return &AuthUsecase{
		repo:      repo,
		tokenator: tokenator,
	}
}

func (uc *AuthUsecase) Authenticate(ctx context.Context, email, password string) (string, error) {
	const op = "AuthUsecase.Authenticate"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("email", email)

	user, err := uc.repo.GetUserByEmail(ctx, email)
	if err != nil {
		logger.WithError(err).Warn("failed to get user by email")
		return "", err
	}
	if user == nil {
		logger.Warn("user not found")
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		logger.Warn("invalid password")
		return "", errors.New("invalid password")
	}

	token, err := uc.tokenator.CreateJWT(user.ID.String(), user.Role)
	if err != nil {
		logger.WithError(err).Error("failed to create JWT")
		return "", err
	}

	return token, nil
}

func (uc *AuthUsecase) Register(ctx context.Context, email, password, role string) (string, error) {
	const op = "AuthUsecase.Register"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithFields(map[string]interface{}{
		"email": email,
		"role":  role,
	})

	if !allowedRoles[role] {
		logger.Warn("role not allowed")
		return "", errs.ErrCityNotAllowed
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.WithError(err).Error("failed to hash password")
		return "", err
	}

	user, err := uc.repo.CreateUser(ctx, email, hashedPassword, role)
	if err != nil {
		logger.WithError(err).Error("failed to create user")
		return "", err
	}

	token, err := uc.tokenator.CreateJWT(user.ID.String(), user.Role)
	if err != nil {
		logger.WithError(err).Error("failed to create JWT after registration")
		return "", err
	}

	return token, nil
}

func (uc *AuthUsecase) DummyLogin(role string) (string, error) {
	const op = "AuthUsecase.DummyLogin"
	logger := logctx.GetLogger(context.Background()).WithField("op", op).WithField("role", role)

	token, err := uc.tokenator.CreateJWT(uuid.New().String(), role)
	if err != nil {
		logger.WithError(err).Error("failed to create dummy token")
		return "", err
	}
	
	return token, nil
}