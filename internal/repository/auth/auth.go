package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	models "github.com/nik-mLb/avito_task/internal/models/user"
	"github.com/nik-mLb/avito_task/internal/transport/middleware/logctx"
)

const (
	createUserQuery = `
		INSERT INTO "user" (id, email, password_hash, role) 
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, role`

	getUserByEmailQuery = `
		SELECT id, email, password_hash, role 
		FROM "user" 
		WHERE email = $1`
)

type AuthRepository struct {
	db *sql.DB
}

func New(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(ctx context.Context, email string, passwordHash []byte, role string) (*models.User, error) {
	const op = "AuthRepository.CreateUser"
	logger := logctx.GetLogger(ctx).WithField("op", op).
		WithField("email", email).
		WithField("role", role)
		
	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	}

	err := r.db.QueryRowContext(ctx, createUserQuery, 
		user.ID, user.Email, user.PasswordHash, user.Role).
		Scan(&user.ID, &user.Email, &user.Role)

	if err != nil {
		logger.WithError(err).Error("failed to create user")
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	
	return user, nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "AuthRepository.GetUserByEmail"
	logger := logctx.GetLogger(ctx).WithField("op", op).WithField("email", email)

	var user models.User
	err := r.db.QueryRowContext(ctx, getUserByEmailQuery, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("user not found")
			return nil, nil
		}
		logger.WithError(err).Error("failed to get user by email")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}