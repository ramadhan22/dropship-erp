// File: backend/cmd/api/main.go

package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/ramadhan22/dropship-erp/backend/internal/config"
	"github.com/ramadhan22/dropship-erp/backend/internal/handlers"
	"github.com/ramadhan22/dropship-erp/backend/internal/migrations"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

func main() {
	// 1) Load configuration (from config.yaml and environment)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Fatal error loading config: %v", err)
	}

	// 2) Initialize repositories (Postgres DB connection)
	repo, err := repository.NewPostgresRepository(cfg.Database.URL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	if err := migrations.Run(repo.DB.DB); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Printf("DB migrations: %v", err)
		} else {
			log.Fatalf("DB migrations failed: %v", err)
		}
	}

	// 3) Initialize services with the appropriate repo interfaces
	dropshipSvc := service.NewDropshipService(repo.DropshipRepo)
	shopeeSvc := service.NewShopeeService(repo.ShopeeRepo)
	reconSvc := service.NewReconcileService(
		repo.DropshipRepo, repo.ShopeeRepo, repo.JournalRepo, repo.ReconcileRepo,
	)
	metricSvc := service.NewMetricService(
		repo.DropshipRepo, repo.ShopeeRepo, repo.JournalRepo, repo.MetricRepo,
	)
	balanceSvc := service.NewBalanceService(repo.JournalRepo)

	// 4) Setup Gin router and API routes
	router := gin.Default()
	// CORS configuration – allow your Vite dev server origin
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5175"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	apiGroup := router.Group("/api")
	{
		apiGroup.POST("/dropship/import", handlers.NewDropshipHandler(dropshipSvc).HandleImport)
		apiGroup.POST("/shopee/import", handlers.NewShopeeHandler(shopeeSvc).HandleImport)
		apiGroup.POST("/reconcile", handlers.NewReconcileHandler(reconSvc).HandleMatchAndJournal)
		apiGroup.POST("/metrics", handlers.NewMetricHandler(metricSvc).HandleCalculateMetrics)
		apiGroup.GET("/metrics", handlers.NewMetricHandler(metricSvc).HandleGetMetrics)
		apiGroup.GET("/balancesheet", handlers.NewBalanceHandler(balanceSvc).HandleGetBalanceSheet)
	}

	// 5) Start the HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
