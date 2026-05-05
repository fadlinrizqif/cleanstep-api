package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fadlinrizqif/cleanstep-api/internal/app"
	"github.com/fadlinrizqif/cleanstep-api/internal/database"
	"github.com/fadlinrizqif/cleanstep-api/internal/handlers"
	"github.com/fadlinrizqif/cleanstep-api/internal/middlware"
	//"github.com/fadlinrizqif/cleanstep-api/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/midtrans/midtrans-go"

	_ "github.com/lib/pq"
)

func main() {
	router := gin.Default()
	router.SetTrustedProxies(nil)

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	serverSecret := os.Getenv("SEVER_SECRET")
	googleSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleID := os.Getenv("GOOGLE_CLIENT_ID")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	midtransKey := os.Getenv("MIDTRANS_SERVER_KEY")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	for i := range 5 {
		err := db.Ping()
		if err == nil {
			fmt.Print("Connection to database success")
			break
		}

		fmt.Printf("Still connecting to database..%d.\n", i)
		time.Sleep(2 * time.Second)
	}

	//hub := ws.NewHub()
	//go hub.Run()

	dbQueries := database.New(db)
	config := app.App{
		DB:           db,
		DBqueries:    dbQueries,
		SeverSecret:  serverSecret,
		GoogleSecret: googleSecret,
		GoogleID:     googleID,
		RedirectURL:  redirectURL,
		MidtransKey:  midtransKey,
		//Hub:          hub,
	}

	midtrans.ServerKey = midtransKey
	midtrans.Environment = midtrans.Sandbox

	userHandler := handlers.NewUserHandler(&config)
	productHandler := handlers.NewProductsHandler(&config)
	orderHandler := handlers.NewOrdersHandler(&config)

	router.POST("/api/signup", userHandler.CreateUser)
	router.POST("/api/login", userHandler.LoginUser)
	router.GET("/api/logout", userHandler.LogoutUser)

	router.GET("/auth/google/login", userHandler.OauthLogin)
	router.GET("/auth/google/callback", userHandler.OauthCallback)

	protected := router.Group("/api")
	protected.Use(middlware.AuthMiddleware(config.SeverSecret))
	{
		protected.POST("/admin/products", productHandler.CreateProducts)
		protected.POST("/admin/products/bulk", productHandler.CreateMassProducts)
		protected.GET("/products", productHandler.GetAllProducts)
		protected.GET("/products/{productID}", productHandler.GetProducts)

		protected.POST("/orders", orderHandler.CreateOrders)
		protected.GET("/orders/notification", orderHandler.NotificationToClient)

	}
	router.POST("/api/orders/callback/webhook", orderHandler.NotificationUrl)

	router.Run(":8080")
}
