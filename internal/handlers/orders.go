package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fadlinrizqif/cleanstep-api/internal/app"
	"github.com/fadlinrizqif/cleanstep-api/internal/dto"
	"github.com/fadlinrizqif/cleanstep-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go/coreapi"
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
		return
	}

	userID, err := uuid.Parse(fmt.Sprint(val))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Soemething wrong in server"})
		return
	}

	var orderParams dto.Params
	// Bind json to the struct
	if err := c.BindJSON(&orderParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newReqOrder := dto.ReqOrderParams{
		Ctx:         c.Request.Context(),
		DB:          h.App.DB,
		DBqueries:   h.App.DBqueries,
		OrderParams: orderParams,
		UserId:      userID,
	}

	newOrder, err := service.CreateNewOrder(newReqOrder)

	totalItem, err := strconv.Atoi(newOrder.GrossAmount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.OrderResponse{
		ID:         newOrder.OrderID,
		UserID:     userID,
		Status:     newOrder.TransactionStatus,
		TotalItem:  int32(totalItem),
		Action:     newOrder.Actions,
		QrString:   newOrder.QRString,
		ExpiryTime: newOrder.ExpiryTime,
	})

}

func (h *OrdersHandler) NotificationUrl(c *gin.Context) {
	var notficationPayload map[string]interface{}

	if err := c.BindJSON(&notficationPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderId, exists := notficationPayload["order_id"].(string)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something wrong from midtrans server"})
		return
	}

	orderDBId, _ := uuid.Parse(orderId)

	transactionStatusResp, err := coreapi.CheckTransaction(orderId)
	if err.RawError != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something wrong from midtrans server"})
		return
	}

	switch transactionStatusResp.TransactionStatus {
	case "capture":
		if transactionStatusResp.FraudStatus == "capture" {
			err := UpdateDBOrder(h.App.DBqueries, c.Request.Context(), "DENIED", orderDBId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "something wrong from databse"})
				return
			}
		} else if transactionStatusResp.FraudStatus == "accept" {
			err := UpdateDBOrder(h.App.DBqueries, c.Request.Context(), "SUCCESS", orderDBId)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "something wrong from databse"})
				return
			}
		}
	case "settlement":
		err := UpdateDBOrder(h.App.DBqueries, c.Request.Context(), "SUCCESS", orderDBId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something wrong from database"})
			return
		}
	case "deny", "cancel", "expire":
		err := UpdateDBOrder(h.App.DBqueries, c.Request.Context(), "FAILED", orderDBId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something wrong from database"})
			return
		}
	case "pending":
		err := UpdateDBOrder(h.App.DBqueries, c.Request.Context(), "PENDING", orderDBId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something wrong from database"})
			return
		}
	}

	c.JSON(http.StatusAccepted, "ok")

}
