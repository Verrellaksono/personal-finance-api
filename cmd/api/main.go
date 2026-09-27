package main

import (
	"log"
	"net/http"
	"time"

	"personal-finance/internal/account"
	"personal-finance/internal/auth"
	"personal-finance/internal/platform/database"
	"personal-finance/internal/transaction"
)

func main() {
	// Konfigurasi Database
	cfg := database.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "root",
		DBName:   "personal_finance",
		SSLMode:  "disable",
	}

	// Database Connection
	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	// Inisialisasi AUTH
	jwtSecret := "supersecretjwtkey12345"
	jwtExpires := 24 * time.Hour
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, jwtSecret, jwtExpires)
	authHandler := auth.NewHandler(authService)

	// Middleware
	authMiddleware := auth.NewMiddleware(authService)

	// Dependency Injection Transaction
	transactionRepo := transaction.NewRepository()
	transactionService := transaction.NewService(db, transactionRepo)
	transactionHandler := transaction.NewHandler(transactionService)

	// Dependency Injection Account
	accRepo := account.NewRepository(db)
	accService := account.NewService(accRepo)
	accHandler := account.NewHandler(accService)

	// Routing
	mux := http.NewServeMux()

	// Route Auth
	mux.HandleFunc("POST /api/v1/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/login", authHandler.Login)

	// Route Transaction
	mux.HandleFunc("POST /api/v1/transactions", authMiddleware.RequireAuth(transactionHandler.CreateTransaction))

	// Route Account
	mux.HandleFunc("POST /api/v1/accounts", authMiddleware.RequireAuth(accHandler.CreateAccount))
	mux.HandleFunc("GET /api/v1/accounts/{id}", authMiddleware.RequireAuth(accHandler.GetAccount))

	// Konfigurasi HTTP Server
	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server started on localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Serve failed to start %v", err)
	}
}
