package models

import (
	"time"
	
	"github.com/google/uuid"
)

type Reception struct {
	ID            uuid.UUID `json:"id"`
	ReceptionDate time.Time `json:"dateTime"`
	PickupPointID uuid.UUID `json:"pvzId"`
	Status        string    `json:"status"` // "in_progress" или "close"
}