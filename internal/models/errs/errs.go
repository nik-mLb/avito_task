package models

import "errors"

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrNoActiveReception = errors.New("no active reception found")
	ErrNoProductsToDelete = errors.New("no products to delete in active reception")
	ErrActiveReceptionExists = errors.New("active reception already exists")
	ErrNoActiveReceptionToClose = errors.New("no active reception to close")
	ErrCityNotAllowed = errors.New("city not allowed")
	ErrRoleNotAllowed = errors.New("role not allowed")
)