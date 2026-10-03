package main

import (
	"log"
	"net/http"
	"time"

	"personal-finance/internal/account"
	"personal-finance/internal/auth"
	"personal-finance/internal/budget"
	"personal-finance/internal/category"
	"personal-finance/internal/platform/database"
	"personal-finance/internal/report"
	"personal-finance/internal/transaction"
	"personal-finance/internal/transfer"
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

	// Dependency injection Category
	categoryRepo := category.NewRepository(db)
	categoryService := category.NewService(categoryRepo)
	categoryHandler := category.NewHandler(categoryService)

	// Dependency Injection Budget
	budgetRepo := budget.NewRepository(db)
	budgetService := budget.NewService(budgetRepo)
	budgetHandler := budget.NewHandler(budgetService)

	// Dependency Injection Budget
	reportRepo := report.NewRepository(db)
	reportService := report.NewService(reportRepo)
	reportHandler := report.NewHandler(reportService)

	// Dependency injection Transfer
	transferRepo := transfer.NewRepository(db)
	transferService := transfer.NewService(transferRepo)
	transferHandler := transfer.NewHandler(transferService)

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

	// Route Categories
	mux.HandleFunc("POST /api/v1/categories", authMiddleware.RequireAuth(categoryHandler.CreateCategory))
	mux.HandleFunc("GET /api/v1/categories", authMiddleware.RequireAuth(categoryHandler.ListCategories))

	// Route Cudget
	mux.HandleFunc("POST /api/v1/budgets", authMiddleware.RequireAuth(budgetHandler.SetBudget))
	mux.HandleFunc("GET /api/v1/budgets/progress", authMiddleware.RequireAuth(budgetHandler.GetBudgetProgress))

	// Route Report
	mux.HandleFunc("GET /api/v1/reports/monthly-summary", authMiddleware.RequireAuth(reportHandler.MonthlySummary))

	// Route Transfer
	mux.HandleFunc("POST /api/v1/transfers", authMiddleware.RequireAuth(transferHandler.Transfer))

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
