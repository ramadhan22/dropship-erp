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
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/migrations"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
	"github.com/ramadhan22/dropship-erp/backend/internal/service"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// 1) Load configuration (from config.yaml and environment)
	cfg, err := config.LoadConfig()
	if err != nil {
		logutil.Fatalf("Fatal error loading config: %v", err)
	}

	// 2) Initialize repositories (Postgres DB connection)
	repo, err := repository.NewPostgresRepository(cfg.Database.URL)
	if err != nil {
		logutil.Fatalf("DB connection failed: %v", err)
	}
	if err := migrations.Run(repo.DB.DB); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Printf("DB migrations: %v", err)
		} else {
			logutil.Fatalf("DB migrations failed: %v", err)
		}
	}

	// 3) Initialize services with the appropriate repo interfaces
	dropshipSvc := service.NewDropshipService(repo.DB, repo.DropshipRepo, repo.JournalRepo)
	shopeeSvc := service.NewShopeeService(repo.DB, repo.ShopeeRepo, repo.DropshipRepo, repo.JournalRepo)
	shClient := service.NewShopeeClient(cfg.Shopee)
	reconSvc := service.NewReconcileService(
		repo.DB,
		repo.DropshipRepo, repo.ShopeeRepo, repo.JournalRepo, repo.ReconcileRepo,
		shClient,
	)
	metricSvc := service.NewMetricService(
		repo.DropshipRepo, repo.ShopeeRepo, repo.JournalRepo, repo.MetricRepo,
	)
	expenseSvc := service.NewExpenseService(repo.DB, repository.NewExpenseRepo(repo.DB), repo.JournalRepo)
	balanceSvc := service.NewBalanceService(repo.JournalRepo)
	channelSvc := service.NewChannelService(repo.ChannelRepo)
	accountSvc := service.NewAccountService(repo.AccountRepo)
	adsSvc := service.NewAdInvoiceService(repo.DB, repo.AdInvoiceRepo, repo.JournalRepo)
	journalSvc := service.NewJournalService(repo.DB, repo.JournalRepo)
	plSvc := service.NewPLService(repo.MetricRepo, metricSvc)
	plReportSvc := service.NewProfitLossReportService(repo.JournalRepo)
	glSvc := service.NewGLService(repo.JournalRepo)
	pbSvc := service.NewPendingBalanceService(shClient)
	assetSvc := service.NewAssetAccountService(repo.AssetAccountRepo, repo.JournalRepo)

	// 4) Setup Gin router and API routes
	router := gin.Default()
	// CORS configuration – origins can be configured via config.yaml
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.CorsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		AllowOriginFunc:  func(origin string) bool { return true },
	}))

	apiGroup := router.Group("/api")
	{
		apiGroup.POST("/dropship/import", handlers.NewDropshipHandler(dropshipSvc).HandleImport)
		apiGroup.GET("/dropship/purchases", handlers.NewDropshipHandler(dropshipSvc).HandleList)
		apiGroup.GET("/dropship/purchases/summary", handlers.NewDropshipHandler(dropshipSvc).HandleSum)
		apiGroup.GET("/dropship/purchases/daily", handlers.NewDropshipHandler(dropshipSvc).HandleDailyTotals)
		apiGroup.GET("/dropship/purchases/:id/details", handlers.NewDropshipHandler(dropshipSvc).HandleListDetails)
		apiGroup.GET("/dropship/top-products", handlers.NewDropshipHandler(dropshipSvc).HandleTopProducts)
		shHandler := handlers.NewShopeeHandler(shopeeSvc)
		apiGroup.POST("/shopee/import", shHandler.HandleImport)
		apiGroup.POST("/shopee/affiliate", shHandler.HandleImportAffiliate)
		apiGroup.POST("/shopee/settle/:order_sn", shHandler.HandleConfirmSettle)
		apiGroup.GET("/shopee/affiliate", shHandler.HandleListAffiliate)
		apiGroup.GET("/shopee/affiliate/summary", shHandler.HandleSumAffiliate)
		apiGroup.GET("/shopee/settled", shHandler.HandleListSettled)
		apiGroup.GET("/shopee/settled/:order_sn", shHandler.HandleGetSettleDetail)
		apiGroup.GET("/shopee/settled/summary", shHandler.HandleSumSettled)
		apiGroup.GET("/sales", shHandler.HandleListSalesProfit)
		apiGroup.POST("/reconcile", handlers.NewReconcileHandler(reconSvc).HandleMatchAndJournal)
		apiGroup.POST("/metrics", handlers.NewMetricHandler(metricSvc).HandleCalculateMetrics)
		apiGroup.GET("/metrics", handlers.NewMetricHandler(metricSvc).HandleGetMetrics)
		apiGroup.GET("/balancesheet", handlers.NewBalanceHandler(balanceSvc).HandleGetBalanceSheet)
		apiGroup.POST("/jenis-channels", handlers.NewChannelHandler(channelSvc).HandleCreateJenisChannel)
		apiGroup.POST("/stores", handlers.NewChannelHandler(channelSvc).HandleCreateStore)
		chH := handlers.NewChannelHandler(channelSvc)
		apiGroup.GET("/stores", chH.HandleListStoresByName)
		apiGroup.GET("/stores/all", chH.HandleListAllStores)
		apiGroup.GET("/stores/:id", chH.HandleGetStore)
		apiGroup.PUT("/stores/:id", chH.HandleUpdateStore)
		apiGroup.DELETE("/stores/:id", chH.HandleDeleteStore)
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

		adsHandler := handlers.NewAdInvoiceHandler(adsSvc)
		adsHandler.RegisterRoutes(apiGroup)

		jHandler := handlers.NewJournalHandler(journalSvc)
		jHandler.RegisterRoutes(apiGroup)

		handlers.NewPLHandler(plSvc).RegisterRoutes(apiGroup)
		handlers.NewProfitLossReportHandler(plReportSvc).RegisterRoutes(apiGroup)
		handlers.NewGLHandler(glSvc).RegisterRoutes(apiGroup)
		handlers.NewReconcileExtraHandler(reconSvc).RegisterRoutes(apiGroup)
		handlers.NewPendingBalanceHandler(pbSvc).RegisterRoutes(apiGroup)
		handlers.NewAssetAccountHandler(assetSvc).RegisterRoutes(apiGroup)
		handlers.NewWithdrawHandler(shopeeSvc).RegisterRoutes(apiGroup)
		handlers.NewConfigHandler(cfg).RegisterRoutes(apiGroup)
	}

	// 5) Start the HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		logutil.Fatalf("Server failed: %v", err)
	}
}
