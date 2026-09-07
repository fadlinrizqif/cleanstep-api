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
	"github.com/fadlinrizqif/cleanstep-api/internal/service"
	"github.com/fadlinrizqif/cleanstep-api/internal/ws"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"

	//_ "github.com/lib/pq"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	for i := range 5 {
		err := db.Ping()
		if err == nil {
			fmt.Println("Connection to database success")
			break
		}

		fmt.Printf("Still connecting to database..%d.\n", i)
		time.Sleep(2 * time.Second)
	}

	hub := ws.NewHub()
	go hub.Run()

	dbQueries := database.New(db)
	config := app.App{
		DB:           db,
		DBqueries:    dbQueries,
		SeverSecret:  serverSecret,
		GoogleSecret: googleSecret,
		GoogleID:     googleID,
		RedirectURL:  redirectURL,
		MidtransKey:  midtransKey,
		Hub:          hub,
	}

	//midtrans.ServerKey = midtransKey
	//midtrans.Environment = midtrans.Sandbox
	fmt.Println(midtransKey)
	c := coreapi.Client{}
	c.New(midtransKey, midtrans.Sandbox)

	serviceHandler := service.NewOrderService(&config, c)

	userHandler := handlers.NewUserHandler(&config)
	productHandler := handlers.NewProductsHandler(&config)
	orderHandler := handlers.NewOrdersHandler(&config, serviceHandler)

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "Cookie"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.POST("/api/signup", userHandler.CreateUser)
	router.POST("/api/login", userHandler.LoginUser)
	router.GET("/api/logout", userHandler.LogoutUser)

	router.GET("/auth/google/login", userHandler.OauthLogin)
	router.GET("/auth/google/callback", userHandler.OauthCallback)

	router.GET("api/products", productHandler.GetAllProducts)

	protected := router.Group("/api")
	protected.Use(middlware.AuthMiddleware(&config))
	{
		protected.POST("/admin/products/bulk", productHandler.CreateMassProducts)
		protected.POST("/admin/products", productHandler.CreateProducts)
		protected.GET("/products/:productID", productHandler.GetProducts)

		protected.GET("/users", userHandler.GetUserById)

		protected.POST("/orders", orderHandler.CreateOrders)
		protected.GET("/orders/:orderID", orderHandler.GetPayment)

		protected.GET("/ws/payment", orderHandler.NotificationToClient)

	}
	router.POST("/api/orders/callback/webhook", orderHandler.NotificationUrl)

	router.Run(":8080")
}
