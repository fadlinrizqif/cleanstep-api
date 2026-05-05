package service

import (
	"errors"

	"github.com/fadlinrizqif/cleanstep-api/internal/database"
	"github.com/fadlinrizqif/cleanstep-api/internal/dto"
	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
)

type OrderDetail struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int32     `json:"quantity"`
}

type Params struct {
	OrderItems []OrderDetail `json:"order_detail"`
}

func CreateNewOrder(orderReq dto.ReqOrderParams) (coreapi.ChargeResponse, error) {

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
		product, err := qtx.GetProduct(orderReq.Ctx, item.ProductID)
		if err != nil {
			return coreapi.ChargeResponse{}, err
		}

		if product.Stock < item.Quantity {
			return coreapi.ChargeResponse{}, errors.New(product.Name + "out of stock")
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
	coreApiRes, _ := coreapi.ChargeTransaction(chargeReq)
	return *coreApiRes, nil

}
