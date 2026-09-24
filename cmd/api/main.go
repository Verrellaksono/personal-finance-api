package main

import (
	"log"
	"net/http"
	"time"

	"personal-finance/internal/account"
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

	// Akun Dummy
	// var accountID int64
	// _ = db.QueryRow("SELECT id FROM accounts LIMIT 1").Scan(&accountID)
	// if accountID == 0 {
	// 	queryInitAccount := `
	// 		INSERT INTO accounts (user_id, name, type, currency, balance)
	// 		VALUES (1, 'BCA Rekening Utama', 'BANK', 'IDR', 100000)
	// 		RETURNING id;
	// 	`
	// 	_ = db.QueryRow(queryInitAccount).Scan(&accountID)
	// 	log.Printf("Akun dummy berhasil dibuat dengan ID: %d", accountID)
	// }

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

	// Route Transaction
	mux.HandleFunc("/api/v1/transactions", transactionHandler.CreateTransaction)

	// Route Account
	mux.HandleFunc("/api/v1/accounts", accHandler.CreateAccount)
	mux.HandleFunc("/api/v1/accounts/{id}", accHandler.GetAccount)

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
