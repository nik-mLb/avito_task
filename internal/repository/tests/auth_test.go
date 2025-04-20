package tests

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	user "github.com/nik-mLb/avito_task/internal/models/user"
	repository "github.com/nik-mLb/avito_task/internal/repository/auth"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := repository.New(db)

	tests := []struct {
		name        string
		email       string
		password    []byte
		role        string
		mock        func()
		expected    *user.User
		expectedErr bool
	}{
		{
			name:     "Success",
			email:    "test@example.com",
			password: []byte("hashed_password"),
			role:     "user",
			mock: func() {
				expectedID := uuid.New()
				rows := sqlmock.NewRows([]string{"id", "email", "role"}).
					AddRow(expectedID, "test@example.com", "user")
				
				mock.ExpectQuery(`INSERT INTO "user"`).
					WithArgs(sqlmock.AnyArg(), "test@example.com", []byte("hashed_password"), "user").
					WillReturnRows(rows)
			},
			expected: &user.User{
				Email: "test@example.com",
				Role:  "user",
			},
			expectedErr: false,
		},
		{
			name:     "Database Error",
			email:    "test@example.com",
			password: []byte("hashed_password"),
			role:     "user",
			mock: func() {
				mock.ExpectQuery(`INSERT INTO "user"`).
					WithArgs(sqlmock.AnyArg(), "test@example.com", []byte("hashed_password"), "user").
					WillReturnError(errors.New("database error"))
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			user, err := repo.CreateUser(context.Background(), tt.email, tt.password, tt.role)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Email, user.Email)
				assert.Equal(t, tt.expected.Role, user.Role)
				assert.NotEqual(t, uuid.Nil, user.ID)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := repository.New(db)

	tests := []struct {
		name        string
		email       string
		mock        func()
		expected    *user.User
		expectedErr bool
	}{
		{
			name:  "Success",
			email: "test@example.com",
			mock: func() {
				expectedID := uuid.New()
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
					AddRow(expectedID, "test@example.com", []byte("hashed_password"), "user")
				
				mock.ExpectQuery(`SELECT id, email, password_hash, role FROM "user"`).
					WithArgs("test@example.com").
					WillReturnRows(rows)
			},
			expected: &user.User{
				Email:        "test@example.com",
				PasswordHash: []byte("hashed_password"),
				Role:         "user",
			},
			expectedErr: false,
		},
		{
			name:  "Not Found",
			email: "notfound@example.com",
			mock: func() {
				mock.ExpectQuery(`SELECT id, email, password_hash, role FROM "user"`).
					WithArgs("notfound@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			expected:    nil,
			expectedErr: false,
		},
		{
			name:  "Database Error",
			email: "test@example.com",
			mock: func() {
				mock.ExpectQuery(`SELECT id, email, password_hash, role FROM "user"`).
					WithArgs("test@example.com").
					WillReturnError(errors.New("database error"))
			},
			expected:    nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mock()

			user, err := repo.GetUserByEmail(context.Background(), tt.email)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected == nil {
					assert.Nil(t, user)
				} else {
					assert.Equal(t, tt.expected.Email, user.Email)
					assert.Equal(t, tt.expected.Role, user.Role)
					assert.Equal(t, tt.expected.PasswordHash, user.PasswordHash)
					assert.NotEqual(t, uuid.Nil, user.ID)
				}
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}