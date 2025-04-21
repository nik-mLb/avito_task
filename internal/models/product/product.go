package models

import (
	"time"
	
	"github.com/google/uuid"
)

type ProductType string

const (
	Electronics ProductType = "электроника"
	Clothing    ProductType = "одежда"
	Shoes       ProductType = "обувь"
)

type Product struct {
	ID            uuid.UUID   `json:"id"`
	ReceptionDate time.Time   `json:"dateTime"`
	ReceptionID   uuid.UUID   `json:"receptionId"`
	ProductType   ProductType `json:"type"`
}