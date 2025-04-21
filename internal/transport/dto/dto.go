package dto

import(
	pickup "github.com/nik-mLb/avito_task/internal/models/pickup_point"
	product "github.com/nik-mLb/avito_task/internal/models/product"
	reception "github.com/nik-mLb/avito_task/internal/models/reception"
)

// TokenResponse представляет ответ с JWT токеном
type TokenResponse struct {
	Token string `json:"token"`
}

// DummyLoginRequest представляет запрос для получения тестового токена
type DummyLoginRequest struct {
	Role string `json:"role"`
}

// LoginRequest представляет запрос на аутентификацию
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest представляет запрос на регистрацию
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// PickupPointRequest представляет запрос на создание ПВЗ
type PickupPointRequest struct {
	City string `json:"city"`
}

// PickupPointListResponse представляет ответ со списком ПВЗ и приемками
type PickupPointListResponse struct {
	PickupPoint        pickup.PickupPoint        `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}

// ReceptionWithProducts представляет приемку с товарами
type ReceptionWithProducts struct {
	Reception reception.Reception `json:"reception"`
	Products  []product.Product `json:"products"`
}

type ReceptionRequest struct {
	PickupPointID string `json:"pvzId"`
}

type ProductRequest struct {
	Type  string `json:"type"`
	PickupPointID string `json:"pvzId"`
}

// ErrorResponse представляет ответ с описанием ошибки
type ErrorResponse struct {
	Message string `json:"message"`
}