package dto

import(
	pickup "github.com/nik-mLb/avito_task/internal/models/pickup_point"
	product "github.com/nik-mLb/avito_task/internal/models/product"
	reception "github.com/nik-mLb/avito_task/internal/models/reception"
)

//для auth
type TokenResponse struct {
	Token string `json:"token"`
}

type DummyLoginRequest struct {
	Role string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}


type PickupPointRequest struct {
	City string `json:"city"`
}

type PickupPointListResponse struct {
	PickupPoint        pickup.PickupPoint        `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}

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

type ErrorResponse struct {
	Message string `json:"message"`
}