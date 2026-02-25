package handlers

import (
	"database/sql"
	"fmt"
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

func (h *OrdersHandler) CreateOrders(c *gin.Context) {
	val, ok := c.Get("getUserID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	userID, err := uuid.Parse(fmt.Sprint(val))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Soemething wrong in server"})
	}

	type OrderDetail struct {
		ProductID uuid.UUID `json:"product_id"`
		Quantity  int32     `json:"quantity"`
	}

	type Params struct {
		OrderItems []orderDetail `json:"order_detail"`
	}

	var orderParams Params

	// Bind json to the struct
	if err := c.BindJSON(&orderParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// begin transaction
	tx, err := h.App.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//rollback if there is something wrong
	defer tx.Rollback()

	qtx := h.App.DBqueries.WithTx(tx)

	var totalPrice int32

	for _, item := range orderParams.OrderItems {
		//decrease the stock from db
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

		//sum up the price of each product's price times with order's quantity
		totalPrice += product.Price * item.Quantity

	}

	//make order the order to db
	newOrder, err := qtx.CreateOrder(c.Request.Context(), database.CreateOrderParams{
		UserID:     userID,
		Status:     "PENDING",
		TotalItems: totalPrice,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, item := range orderParams.OrderItems {

		//store each item to the db with foreign from order
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

	//if there is no error the change to the database saved
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//send respond to the frontend
	c.JSON(http.StatusOK, dto.OrderResponse{
		ID:        newOrder.ID,
		UserID:    userID,
		Status:    newOrder.Status,
		TotalItem: newOrder.TotalItems,
	})

}
