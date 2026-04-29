package dto

import (
	"context"
	"database/sql"
	"time"

	"github.com/fadlinrizqif/cleanstep-api/internal/database"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go/coreapi"
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

type OrderDetail struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int32     `json:"quantity"`
}

type Params struct {
	OrderItems []OrderDetail `json:"order_detail"`
}

type ReqOrderParams struct {
	Ctx         context.Context
	DB          *sql.DB
	DBqueries   *database.Queries
	OrderParams Params
	UserId      uuid.UUID
	MidtransKey string
}

type ActionUser struct {
	Name   string   `json:"name"`
	Method string   `json:"method"`
	URL    string   `json:"url"`
	Fields []string `json:"fields"`
}

type OrderResponse struct {
	ID         string           `json:"id"`
	UserID     uuid.UUID        `json:"user_id"`
	Status     string           `json:"status"`
	TotalItem  int32            `json:"total_item"`
	Action     []coreapi.Action `json:"action"`
	QrString   string           `json:"qr_string"`
	ExpiryTime string           `json:"expiry_time"`
}
