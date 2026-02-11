package handlers

import (
	"database/sql"
	"net/http"

	"github.com/fadlinrizqif/cleanstep-api/internal/app"
	"github.com/fadlinrizqif/cleanstep-api/internal/database"
	"github.com/fadlinrizqif/cleanstep-api/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrdersHandler struct {
	App *app.App
}

func NewOrdersHandler(app *app.App) *OrdersHandler {
	return &OrdersHandler{App: app}
}

func (h *ProductsHandler) CreateOrders(c *gin.Context) {
	type OrderDetail struct {
		ProductID uuid.UUID `json:"product_id"`
		Quantity  int32     `json:"quantity"`
	}

	type Params struct {
		UserID     uuid.UUID     `json:"user_id"`
		OrderItems []orderDetail `json:"order_detail"`
	}

	var orderParams Params

	if err := c.BindJSON(&orderParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.App.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	defer tx.Rollback()

	qtx := h.App.DBqueries.WithTx(tx)

	var totalPrice int32

	for _, item := range orderParams.OrderItems {
		product, err := qtx.UpdateProduct(c.Request.Context(), database.UpdateProductParams{
			ID:    item.ProductID,
			Stock: item.Quantity,
		})
		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		totalPrice += product.Price * item.Quantity

	}

	newOrder, err := qtx.CreateOrder(c.Request.Context(), database.CreateOrderParams{
		UserID:     orderParams.UserID,
		Status:     "PENDING",
		TotalItems: totalPrice,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, item := range orderParams.OrderItems {

		_, err = qtx.CreateOrderItems(c.Request.Context(), database.CreateOrderItemsParams{
			ProductID: item.ProductID,
			OrderID:   newOrder.ID,
			Quantity:  item.Quantity,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.OrderResponse{
		ID:        newOrder.ID,
		UserID:    orderParams.UserID,
		Status:    newOrder.Status,
		TotalItem: newOrder.TotalItems,
	})

}
