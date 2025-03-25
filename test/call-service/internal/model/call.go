package model

import (
	"time"

	"github.com/google/uuid"
)

type Call struct {
	ID          uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	ClientName  string    `bun:"client_name,notnull" json:"client_name"`
	PhoneNumber string    `bun:"phone_number,notnull" json:"phone_number"`
	Description string    `bun:"description,notnull" json:"description"`
	Status      string    `bun:"status,notnull" json:"status"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UserID      uuid.UUID `bun:"user_id,notnull" json:"user_id"`
}

type CreateCallRequest struct {
	ClientName  string `json:"client_name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type UpdateCallStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
