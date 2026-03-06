package app

import (
	"database/sql"

	"github.com/fadlinrizqif/cleanstep-api/internal/database"
)

type App struct {
	DB           *sql.DB
	DBqueries    *database.Queries
	SeverSecret  string
	GoogleSecret string
	GoogleID     string
	RedirectURL  string
}
