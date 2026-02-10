package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserDetail struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}

type CreateUser struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
}

type LoginUser struct {
	Token        string     `json:"token"`
	RefreshToken string     `json:"refresh_token"`
	User         UserDetail `json:"user"`
}

type ProductResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Category string    `json:"category"`
	Price    int32     `json:"price"`
	Stock    int32     `json:"stock"`
}

type GetProductResponse struct {
	Data  []ProductResponse `json:"data"`
	Total int32             `json:"total"`
}
