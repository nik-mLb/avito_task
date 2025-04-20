package models

import (
	"time"
	
	"github.com/google/uuid"
)

type Reception struct {
	ID            uuid.UUID `json:"id"`
	ReceptionDate time.Time `json:"reception_date"`
	PickupPointID uuid.UUID `json:"pickup_point_id"`
	Status        string    `json:"status"` // "in_progress" или "close"
}