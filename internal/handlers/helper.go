package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	//	"github.com/fadlinrizqif/cleanstep-api/internal/database"
	"github.com/fadlinrizqif/cleanstep-api/internal/database"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type orderDetail struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int32     `json:"quantity"`
}

//func validateTransactions(productList map[uuid.UUID]database.GetAllPriceRow, orderList []orderDetail) (bool, error) {
//	for _, order := range orderList {
//		if productList[order.ProductID].Stock < order.Quantity {
//			return false, errors.New(productList[order.ProductID].Name + " Out of stock")
//		}
//
//	}
//	return true, nil
//}

func validateCategory(category string) (string, error) {
	theCategory := strings.ToLower(category)
	categories := map[string]struct{}{
		"service": {},
		"bundle":  {},
		"cleaner": {},
		"tool":    {},
	}

	if _, ok := categories[category]; !ok {
		return "", errors.New("wrong category")
	}

	return theCategory, nil
}

func UpdateDBOrder(dbQuery *database.Queries, ctx context.Context, status string, orderId uuid.UUID) error {
	err := dbQuery.UpdateStatusOrder(ctx, database.UpdateStatusOrderParams{
		Status: status,
		ID:     orderId,
	})
	if err != nil {
		return err
	}

	timePayAt := sql.NullTime{}

	if status == "SUCCESS" {
		timePayAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	errPayment := dbQuery.UpdatePayment(ctx, database.UpdatePaymentParams{
		OrderID: orderId,
		Status:  status,
		PaidAt:  timePayAt,
	})

	if errPayment != nil {
		return err
	}

	return nil
}

func UnmarshalFeatures(rawFeatures pqtype.NullRawMessage) (map[string]string, error) {
	if !rawFeatures.Valid || len(rawFeatures.RawMessage) == 0 {
		return make(map[string]string), nil
	}
	var featuresMap map[string]string
	err := json.Unmarshal(rawFeatures.RawMessage, &featuresMap)
	if err != nil {
		return nil, err
	}
	return featuresMap, nil
}
