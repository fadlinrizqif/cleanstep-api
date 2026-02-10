package handlers

import "github.com/fadlinrizqif/cleanstep-api/internal/app"

type OrdersHandler struct {
	App *app.App
}

func NewOrdersHandler(app *app.App) *OrdersHandler {
	return &OrdersHandler{App: app}
}
