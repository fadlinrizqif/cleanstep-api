package app

import (
	"database/sql"

	"github.com/fadlinrizqif/cleanstep-api/internal/database"
	"github.com/fadlinrizqif/cleanstep-api/internal/ws"
)

type App struct {
	DB           *sql.DB
	DBqueries    *database.Queries
	SeverSecret  string
	GoogleSecret string
	GoogleID     string
	RedirectURL  string
	MidtransKey  string
	Hub          *ws.Hub
}
