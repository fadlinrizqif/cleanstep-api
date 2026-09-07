package service

import (
	//"errors"

	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/fadlinrizqif/cleanstep-api/internal/app"
	"github.com/fadlinrizqif/cleanstep-api/internal/database"
	"github.com/fadlinrizqif/cleanstep-api/internal/dto"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/sqlc-dev/pqtype"
)

type OrderDetail struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int32     `json:"quantity"`
}

type Params struct {
	OrderItems []OrderDetail `json:"order_detail"`
}

type PaymentGateway interface {
	ChargeTransaction(req *coreapi.ChargeReq) (*coreapi.ChargeResponse, *midtrans.Error)
	CheckTransaction(param string) (*coreapi.TransactionStatusResponse, *midtrans.Error)
}

type OrderService struct {
	App     *app.App
	Payment PaymentGateway
}

func NewOrderService(app *app.App, payment PaymentGateway) *OrderService {
	return &OrderService{
		App:     app,
		Payment: payment,
	}
}

func (s OrderService) CreateNewOrder(orderReq dto.ReqOrderParams) (coreapi.ChargeResponse, error) {

	// begin the transaction
	tx, err := orderReq.DB.Begin()
	if err != nil {
		return coreapi.ChargeResponse{}, err
	}

	//if something wrong rollback to this state
	defer tx.Rollback()

	qtx := orderReq.DBqueries.WithTx(tx)

	var totalPrice int32
	priceList := make(map[uuid.UUID]int32)
	//this foor lop to accumulate total amount from client's order items
	//and check the available stock from database
	for _, item := range orderReq.OrderParams.OrderItems {
		//product, err := qtx.GetProduct(orderReq.Ctx, item.ProductID)
		//if err != nil {
		//	return coreapi.ChargeResponse{}, err
		//}

		//if product.Stock < item.Quantity {
		//	return coreapi.ChargeResponse{}, errors.New(product.Name + "out of stock")
		//}
		product, err := qtx.UpdateReservedStock(orderReq.Ctx, database.UpdateReservedStockParams{
			StockReserved: item.Quantity,
			ID:            item.ProductID,
		})

		if err != nil {
			return coreapi.ChargeResponse{}, err
		}

		priceList[product.ID] = product.Price

		totalPrice += product.Price * item.Quantity
	}

	//create order and put to the database with status PENDING
	newOrder, err := qtx.CreateOrder(orderReq.Ctx, database.CreateOrderParams{
		UserID:     orderReq.UserId,
		Status:     "PENDING",
		TotalItems: totalPrice,
	})
	if err != nil {
		return coreapi.ChargeResponse{}, err
	}

	//this for loop to store order items per product
	for _, item := range orderReq.OrderParams.OrderItems {
		_, err := qtx.CreateOrderItems(orderReq.Ctx, database.CreateOrderItemsParams{
			ProductID: item.ProductID,
			OrderID:   newOrder.ID,
			Quantity:  item.Quantity,
			Price:     priceList[item.ProductID],
		})
		if err != nil {
			return coreapi.ChargeResponse{}, err
		}
	}

	//this is the end of transaction
	if err := tx.Commit(); err != nil {
		return coreapi.ChargeResponse{}, err
	}

	//initiate the userID and total amount to the midtrans variable
	chargeReq := &coreapi.ChargeReq{
		PaymentType: coreapi.PaymentTypeQris,
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  newOrder.ID.String(),
			GrossAmt: int64(totalPrice),
		},
	}

	//from midtrans varibale before put to he function to get the bill from midtrans
	coreApiRes, _ := s.Payment.ChargeTransaction(chargeReq)

	orderId, err := uuid.Parse(fmt.Sprint(coreApiRes.OrderID))
	if err != nil {
		return coreapi.ChargeResponse{}, err
	}

	transactionId, err := uuid.Parse(fmt.Sprint(coreApiRes.OrderID))
	if err != nil {
		return coreapi.ChargeResponse{}, err
	}

	totalAmount, err := strconv.ParseFloat(coreApiRes.GrossAmount, 64)
	if err != nil {
		return coreapi.ChargeResponse{}, err
	}

	payloadByte, err := json.Marshal(coreApiRes)
	if err != nil {
		return coreapi.ChargeResponse{}, err
	}

	payloadJSON := pqtype.NullRawMessage{
		RawMessage: json.RawMessage(payloadByte),
		Valid:      true,
	}

	layout := "2006-01-02 15:04:05"
	parsedTime, err := time.Parse(layout, coreApiRes.ExpiryTime)
	if err != nil {
		return coreapi.ChargeResponse{}, err
	}

	nullTime := sql.NullTime{
		Time:  parsedTime,
		Valid: true,
	}

	paymentParams := database.CreatePaymentParams{
		OrderID:     orderId,
		ExternalID:  transactionId,
		Method:      coreApiRes.PaymentType,
		Acquirer:    coreApiRes.Acquirer,
		Currency:    coreApiRes.Currency,
		Status:      coreApiRes.TransactionStatus,
		FraudStatus: coreApiRes.FraudStatus,
		Amount:      int32(totalAmount),
		QrString:    coreApiRes.QRString,
		UrlImage:    coreApiRes.Actions[0].URL,
		Payload:     payloadJSON,
		ExpireAt:    nullTime,
	}

	_, errPayment := s.App.DBqueries.CreatePayment(orderReq.Ctx, paymentParams)
	if errPayment != nil {
		return coreapi.ChargeResponse{}, err
	}

	return *coreApiRes, nil

}
