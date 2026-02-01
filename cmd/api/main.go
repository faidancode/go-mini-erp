package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool" // Gunakan pgxpool untuk performa lebih baik
	"github.com/joho/godotenv"

	"go-mini-erp/internal/auth"
	dbgen "go-mini-erp/internal/shared/database/sqlc"
	"go-mini-erp/internal/shared/middleware"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 1. Database Connection (Menggunakan pgxpool)
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	dbPool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatal("Cannot connect to database pool:", err)
	}
	defer dbPool.Close()

	// Ping database untuk memastikan koneksi aktif
	if err := dbPool.Ping(ctx); err != nil {
		log.Fatal("Database ping failed:", err)
	}

	// sqlc generator sekarang menggunakan dbPool
	queries := dbgen.New(dbPool)

	// 2. Gin Setup
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})
	jwtManager := auth.NewJWTManager(os.Getenv("JWT_SECRET"))

	// 3. Routes Grouping
	v1 := router.Group("/api/v1")
	{
		// Sesuai requirement Anda: sertakan penempatan folder/logic per module
		authRepo := auth.NewRepository(queries)
		authService := auth.NewService(authRepo, queries, jwtManager)
		authHandler := auth.NewHandler(authService)
		authHandler.RegisterRoutes(v1)
	}

	// 4. HTTP Server Setup
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Start Server with Graceful Shutdown
	go func() {
		log.Printf("🚀 Server running on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %v", err)
		}
	}()

	// Menunggu signal interrupt
	<-ctx.Done()

	log.Println("⏳ Shutting down server...")

	// Memberikan waktu 5 detik untuk menyelesaikan request yang sedang berjalan
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
