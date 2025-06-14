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
	dropshipSvc := service.NewDropshipService(repo.DB, repo.DropshipRepo, repo.JournalRepo)
	shopeeSvc := service.NewShopeeService(repo.ShopeeRepo)
	reconSvc := service.NewReconcileService(
		repo.DB,
		repo.DropshipRepo, repo.ShopeeRepo, repo.JournalRepo, repo.ReconcileRepo,
	)
	metricSvc := service.NewMetricService(
		repo.DropshipRepo, repo.ShopeeRepo, repo.JournalRepo, repo.MetricRepo,
	)
	expenseSvc := service.NewExpenseService(repo.DB, repository.NewExpenseRepo(repo.DB), repo.JournalRepo)
	balanceSvc := service.NewBalanceService(repo.JournalRepo)
	channelSvc := service.NewChannelService(repo.ChannelRepo)
	accountSvc := service.NewAccountService(repo.AccountRepo)
	journalSvc := service.NewJournalService(repo.DB, repo.JournalRepo)
	plSvc := service.NewPLService(repo.MetricRepo)
	glSvc := service.NewGLService(repo.JournalRepo)

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
		apiGroup.GET("/dropship/purchases", handlers.NewDropshipHandler(dropshipSvc).HandleList)
		apiGroup.GET("/dropship/purchases/summary", handlers.NewDropshipHandler(dropshipSvc).HandleSum)
		apiGroup.GET("/dropship/purchases/:id/details", handlers.NewDropshipHandler(dropshipSvc).HandleListDetails)
		apiGroup.POST("/shopee/import", handlers.NewShopeeHandler(shopeeSvc).HandleImport)
		apiGroup.GET("/shopee/settled", handlers.NewShopeeHandler(shopeeSvc).HandleListSettled)
		apiGroup.GET("/shopee/settled/summary", handlers.NewShopeeHandler(shopeeSvc).HandleSumSettled)
		apiGroup.POST("/reconcile", handlers.NewReconcileHandler(reconSvc).HandleMatchAndJournal)
		apiGroup.POST("/metrics", handlers.NewMetricHandler(metricSvc).HandleCalculateMetrics)
		apiGroup.GET("/metrics", handlers.NewMetricHandler(metricSvc).HandleGetMetrics)
		apiGroup.GET("/balancesheet", handlers.NewBalanceHandler(balanceSvc).HandleGetBalanceSheet)
		apiGroup.POST("/jenis-channels", handlers.NewChannelHandler(channelSvc).HandleCreateJenisChannel)
		apiGroup.POST("/stores", handlers.NewChannelHandler(channelSvc).HandleCreateStore)
		apiGroup.GET("/stores", handlers.NewChannelHandler(channelSvc).HandleListStoresByName)
		apiGroup.GET("/jenis-channels", handlers.NewChannelHandler(channelSvc).HandleListJenisChannels)
		apiGroup.GET("/jenis-channels/:id/stores", handlers.NewChannelHandler(channelSvc).HandleListStores)

		accHandler := handlers.NewAccountHandler(accountSvc)
		apiGroup.POST("/accounts", accHandler.HandleCreateAccount)
		apiGroup.GET("/accounts", accHandler.HandleListAccounts)
		apiGroup.GET("/accounts/:id", accHandler.HandleGetAccount)
		apiGroup.PUT("/accounts/:id", accHandler.HandleUpdateAccount)
		apiGroup.DELETE("/accounts/:id", accHandler.HandleDeleteAccount)

		expHandler := handlers.NewExpenseHandler(expenseSvc)
		expHandler.RegisterRoutes(apiGroup)

		jHandler := handlers.NewJournalHandler(journalSvc)
		jHandler.RegisterRoutes(apiGroup)

		handlers.NewPLHandler(plSvc).RegisterRoutes(apiGroup)
		handlers.NewGLHandler(glSvc).RegisterRoutes(apiGroup)
		handlers.NewReconcileExtraHandler(reconSvc).RegisterRoutes(apiGroup)
	}

	// 5) Start the HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
