package main

import (
	_ "biblio/docs"
	bookHttp "biblio/internal/book/adapters/primary/http"
	bookPostgres "biblio/internal/book/adapters/secondary/postgres"
	bookEntities "biblio/internal/book/core/entities"
	bookUsecases "biblio/internal/book/core/usecases"
	"biblio/internal/infrastructure"
	"biblio/internal/infrastructure/config"
	loanHttp "biblio/internal/loan/adapters/primary/http"
	loanPostgres "biblio/internal/loan/adapters/secondary/postgres"
	loanEntities "biblio/internal/loan/core/entities"
	loanUsecases "biblio/internal/loan/core/usecases"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

// @title           Biblio API - Smart Library Backend
// @version         1.0
// @description     Enterprise backend API for smart physical library management and book tracking.
// @contact.name    Biblio Engineering Team
// @host            localhost:3000
// @BasePath        /
func main() {
	// 1. Load Environment Configuration (.env)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Establish PostgreSQL Database Connection Pool
	db, err := infrastructure.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 3. Execute Automatic Database Schema Migrations
	log.Println("Running automatic database migrations...")
	if err := db.AutoMigrate(
		&bookEntities.Book{},
		&bookEntities.BookCopy{},
		&loanEntities.Loan{},
		&loanEntities.LoanItem{},
	); err != nil {
		log.Fatalf("Failed to execute database migration: %v", err)
	}

	// 4. Manual Dependency Injection (Wiring 4-Layer Hexagonal Architecture)
	// --- Book Module ---
	bookRepo := bookPostgres.NewBookRepository(db)
	bookService := bookUsecases.NewBookService(bookRepo)
	bookAPI := bookHttp.NewBookAPI(bookService)
	bookHandler := bookHttp.NewBookHandler(bookAPI)

	// --- Loan Module ---
	loanRepo := loanPostgres.NewLoanRepository(db)
	loanService := loanUsecases.NewLoanService(loanRepo, bookRepo)
	loanAPI := loanHttp.NewLoanAPI(loanService)
	loanHandler := loanHttp.NewLoanHandler(loanAPI)

	// 5. Initialize Go Fiber Application & Standard Middleware Stack
	app := fiber.New(fiber.Config{
		AppName:      "Biblio API v1.0",
		ServerHeader: "Fiber",
	})

	app.Use(logger.New())  // Request logging in terminal
	app.Use(recover.New()) // Panic recovery: protects the server against unexpected runtime crashes
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// 6. Base Routes & Interactive Swagger Documentation
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"app":     "Biblio API",
			"time":    time.Now().Format(time.RFC3339),
			"version": "1.0.0",
		})
	})
	app.Get("/swagger/*", swagger.HandlerDefault)

	// 7. Register Module Routes under /api/v1
	apiV1 := app.Group("/api/v1")
	bookHandler.RegisterRoutes(apiV1)
	loanHandler.RegisterRoutes(apiV1)

	// 8. Graceful Shutdown Setup
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	serverPort := fmt.Sprintf(":%s", cfg.AppPort)

	// Launch Fiber server in an asynchronous Goroutine
	go func() {
		log.Printf("🚀 Biblio Server is running on http://localhost%s", serverPort)
		log.Printf("📚 Swagger UI documentation available at: http://localhost%s/swagger/index.html", serverPort)

		if err := app.Listen(serverPort); err != nil {
			log.Printf("Server stopped listening: %v", err)
		}
	}()

	// Block thread until interrupt or termination signal is received
	<-shutdownChan
	log.Println("\n⚠️ Shutdown signal received! Starting graceful cleanup...")

	// Allow up to 10 seconds for in-flight requests to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Error shutting down Fiber server: %v", err)
	}

	log.Println("🛑 Server cleanly shut down. Goodbye!")
}
