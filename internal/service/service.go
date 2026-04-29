package service

import (
	"database/sql"

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

	tx, err := orderReq.DB.Begin()
	if err != nil {
		return coreapi.ChargeResponse{}, err
	}

	defer tx.Rollback()

	qtx := orderReq.DBqueries.WithTx(tx)

	var totalPrice int32
	priceList := make(map[uuid.UUID]int32)
	for _, item := range orderReq.OrderParams.OrderItems {
		product, err := qtx.UpdateProduct(orderReq.Ctx, database.UpdateProductParams{
			ID:    item.ProductID,
			Stock: item.Quantity,
		})
		if err != sql.ErrNoRows {
			return coreapi.ChargeResponse{}, err
		} else if err != nil {
			return coreapi.ChargeResponse{}, err
		}

		priceList[product.ID] = product.Price

		totalPrice += product.Price * item.Quantity
	}

	newOrder, err := qtx.CreateOrder(orderReq.Ctx, database.CreateOrderParams{
		UserID:     orderReq.UserId,
		Status:     "PENDING",
		TotalItems: totalPrice,
	})
	if err != nil {
		return coreapi.ChargeResponse{}, err
	}

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

	if err := tx.Commit(); err != nil {
		return coreapi.ChargeResponse{}, err
	}

	chargeReq := &coreapi.ChargeReq{
		PaymentType: coreapi.PaymentTypeQris,
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  newOrder.ID.String(),
			GrossAmt: int64(totalPrice),
		},
	}

	coreApiRes, _ := coreapi.ChargeTransaction(chargeReq)
	return *coreApiRes, nil

}
