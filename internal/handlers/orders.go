package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/fadlinrizqif/cleanstep-api/internal/app"
	"github.com/fadlinrizqif/cleanstep-api/internal/dto"
	"github.com/fadlinrizqif/cleanstep-api/internal/service"
	"github.com/fadlinrizqif/cleanstep-api/internal/ws"
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

	//parse userID from string to UUID type
	userID, err := uuid.Parse(fmt.Sprint(val))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Soemething wrong in server"})
		return
	}

	var orderParams dto.Params
	// Bind json to the variable
	if err := c.BindJSON(&orderParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// declare the variable for service function
	newReqOrder := dto.ReqOrderParams{
		Ctx:         c.Request.Context(),
		DB:          h.App.DB,
		DBqueries:   h.App.DBqueries,
		OrderParams: orderParams,
		UserId:      userID,
	}

	//this function create order to database and post order to midtrans
	newOrder, err := service.CreateNewOrder(newReqOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something wrong in server"})
		log.Fatal(err)
		return
	}

	totalItem, err := strconv.ParseFloat(newOrder.GrossAmount, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//this the data payment for client from midtrans including qr code string
	c.JSON(http.StatusCreated, dto.OrderResponse{
		ID:         newOrder.OrderID,
		UserID:     userID,
		Status:     newOrder.TransactionStatus,
		TotalItem:  int32(math.Round(totalItem)),
		Action:     newOrder.Actions,
		QrString:   newOrder.QRString,
		ExpiryTime: newOrder.ExpiryTime,
	})

}

func (h *OrdersHandler) NotificationUrl(c *gin.Context) {
	var notficationPayload map[string]interface{}

	//bind the json from midtrans to the notificationPaylod variable
	if err := c.BindJSON(&notficationPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//get transaction_id/orderiD from json payload
	orderId, exists := notficationPayload["order_id"].(string)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something wrong from midtrans server"})
		return
	}

	//parse from string to the UUID type
	orderDBId, _ := uuid.Parse(orderId)

	//check the transaction status
	transactionStatusResp, err := coreapi.CheckTransaction(orderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something wrong from midtrans server"})
		return
	}

	//get the data order from database using transaction_id from midtrans
	order, errDB := h.App.DBqueries.GetOrderByID(c.Request.Context(), orderDBId)
	if errDB != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order id not found"})
		return
	}

	//this is to authenticate the data from midtrans or not
	//rawSignature := order.ID.String() + string(order.TotalItems) + notficationPayload["status_code"].(string) + h.App.MidtransKey
	//hashKey := sha512.Sum512([]byte(rawSignature))
	//expectedSignature := hex.EncodeToString(hashKey[:])

	//if notficationPayload["signature_key"].(string) != expectedSignature {
	//	c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	//	return
	//}

	//checking status of transaction
	switch transactionStatusResp.TransactionStatus {
	case "challange":
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
			_, errDB = h.App.DBqueries.UpdateProduct(c.Request.Context(), order.ID)
			if errDB != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "something wrong from database"})
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
			c.JSON(http.StatusOK, "ok")
			return
		}
	}

	//h.App.Hub.EventCh <- ws.PaymentEvent{
	//	OrderId: orderDBId,
	//	UserId:  order.UserID,
	//	Status:  transactionStatusResp.TransactionStatus,
	//}

	//if transaction success no problem, send the status code to midtrans
	c.JSON(http.StatusOK, "ok")

}

func (h *OrdersHandler) NotificationToClient(c *gin.Context) {
	val, ok := c.Get("getUserID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, _ := uuid.Parse(fmt.Sprint(val))

	conn, err := ws.Upgrader.Upgrade(c.Writer, c.Request, c.Request.Header)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Connection failed"})
		return
	}

	defer conn.Close()
	defer h.App.Hub.Unregister(userID)
	h.App.Hub.Register(userID, conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
