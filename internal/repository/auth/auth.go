package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	models "github.com/nik-mLb/avito_task/internal/models/user"
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
	user := &models.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	}

	err := r.db.QueryRowContext(ctx, createUserQuery, 
		user.ID, user.Email, user.PasswordHash, user.Role).
		Scan(&user.ID, &user.Email, &user.Role)

	return user, err
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowContext(ctx, getUserByEmailQuery, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &user, err
}