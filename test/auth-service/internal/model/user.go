package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Username     string    `bun:"username,notnull,unique"`
	PasswordHash string    `bun:"password_hash,notnull"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp"`
}
