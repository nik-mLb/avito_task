package models

import (
	"github.com/google/uuid"
)

type PickupPoint struct {
	ID        uuid.UUID `json:"id"`
	City      string    `json:"city"`
	RegistrationDate string `json:"registrationDate"`
}