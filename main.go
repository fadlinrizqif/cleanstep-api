package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/fadlinrizqif/cleanstep-api/internal/app"
	"github.com/fadlinrizqif/cleanstep-api/internal/database"
	"github.com/fadlinrizqif/cleanstep-api/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	router := gin.Default()
	router.SetTrustedProxies(nil)

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	serverSecret := os.Getenv("SEVER_SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)
	config := app.App{
		DB:          db,
		DBqueries:   dbQueries,
		SeverSecret: serverSecret,
	}

	userHandler := handlers.NewUserHandler(&config)
	productHandler := handlers.NewProductsHandler(&config)
	orderHandler := handlers.NewOrdersHandler(&config)

	router.POST("/api/signup", userHandler.CreateUser)
	router.POST("/api/login", userHandler.LoginUser)

	router.POST("/api/admin/products", productHandler.CreateProducts)
	router.GET("/api/products", productHandler.GetAllProducts)
	router.GET("/api/products/{productID}", productHandler.GetProducts)
	router.POST("/api/products/bulk", productHandler.CreateMassProducts)

	router.POST("/api/orders", orderHandler.CreateOrders)

	router.Run(":8080")
}
