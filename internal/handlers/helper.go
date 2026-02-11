package handlers

import (
	"errors"

	"github.com/fadlinrizqif/cleanstep-api/internal/database"
	"github.com/google/uuid"
)

type orderDetail struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int32     `json:"quantity"`
}

func validateTransactions(productList map[uuid.UUID]database.GetAllPriceRow, orderList []orderDetail) (bool, error) {
	for _, order := range orderList {
		if productList[order.ProductID].Stock < order.Quantity {
			return false, errors.New(productList[order.ProductID].Name + " Out of stock")
		}

	}
	return true, nil
}
