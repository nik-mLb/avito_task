package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	models "github.com/nik-mLb/avito_task/internal/models/user"
	"github.com/nik-mLb/avito_task/internal/transport/jwt"
	"golang.org/x/crypto/bcrypt"
	errs "github.com/nik-mLb/avito_task/internal/models/errs"
)

var (
	allowedRoles = map[string]bool{
		"worker":          true,
		"admin": 		   true,
	}
)

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
	user, err := uc.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	return uc.tokenator.CreateJWT(user.ID.String(), user.Role)
}

func (uc *AuthUsecase) Register(ctx context.Context, email, password, role string) (string, error) {
	if !allowedRoles[role] {
		return "", errs.ErrCityNotAllowed
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user, err := uc.repo.CreateUser(ctx, email, hashedPassword, role)
	if err != nil {
		return "", err
	}

	return uc.tokenator.CreateJWT(user.ID.String(), user.Role)
}

func (uc *AuthUsecase) DummyLogin(role string) (string, error) {
	return uc.tokenator.CreateJWT(uuid.New().String(), role)
}