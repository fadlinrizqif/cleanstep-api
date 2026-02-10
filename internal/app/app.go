package app

import "github.com/fadlinrizqif/cleanstep-api/internal/database"

type App struct {
	DB          *database.Queries
	SeverSecret string
}
