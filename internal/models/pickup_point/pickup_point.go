package models

import (
	"github.com/google/uuid"
)

// PickupPoint представляет пункт выдачи заказов
// @model PickupPoint
type PickupPoint struct {
	ID        uuid.UUID `json:"id"`
	City      string    `json:"city"`
	RegistrationDate string `json:"registrationDate"`
}