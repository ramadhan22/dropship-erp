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
	w, err := logutil.NewDailyFileWriter(cfg.Logging.Dir)
	if err != nil {
		logutil.Fatalf("open log file: %v", err)
	}
	defer w.Close()
	log.SetOutput(w)

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
	shClient := service.NewShopeeClient(cfg.Shopee)
	dropshipSvc := service.NewDropshipService(
		repo.DB,
		repo.DropshipRepo,
		repo.JournalRepo,
		repo.ChannelRepo,
		repo.OrderDetailRepo,
		shClient,
	)
	shopeeSvc := service.NewShopeeService(repo.DB, repo.ShopeeRepo, repo.DropshipRepo, repo.JournalRepo, repo.ShopeeAdjustmentRepo)
	reconSvc := service.NewReconcileService(
		repo.DB,
		repo.DropshipRepo, repo.ShopeeRepo, repo.JournalRepo, repo.ReconcileRepo,
		repo.ChannelRepo,
		repo.OrderDetailRepo,
		repo.ShopeeAdjustmentRepo,
		shClient,
	)
	metricSvc := service.NewMetricService(
		repo.DropshipRepo, repo.ShopeeRepo, repo.JournalRepo, repo.MetricRepo,
	)
	taxSvc := service.NewTaxService(repo.DB, repo.TaxRepo, repo.JournalRepo, metricSvc)
	expenseSvc := service.NewExpenseService(repo.DB, repository.NewExpenseRepo(repo.DB), repo.JournalRepo)
	balanceSvc := service.NewBalanceService(repo.JournalRepo)
	channelSvc := service.NewChannelService(repo.ChannelRepo, shClient)
	accountSvc := service.NewAccountService(repo.AccountRepo)
	adsSvc := service.NewAdInvoiceService(repo.DB, repo.AdInvoiceRepo, repo.JournalRepo)
	journalSvc := service.NewJournalService(repo.DB, repo.JournalRepo)
	plSvc := service.NewPLService(repo.MetricRepo, metricSvc)
	plReportSvc := service.NewProfitLossReportService(repo.JournalRepo)
	glSvc := service.NewGLService(repo.JournalRepo)
	pbSvc := service.NewPendingBalanceService(shClient)
	walletSvc := service.NewWalletTransactionService(repo.ChannelRepo, shClient)
	adsTopupSvc := service.NewAdsTopupService(walletSvc, repo.JournalRepo)
	walletWdSvc := service.NewWalletWithdrawalService(walletSvc, repo.JournalRepo)
	assetSvc := service.NewAssetAccountService(repo.AssetAccountRepo, repo.JournalRepo)
	withdrawalSvc := service.NewWithdrawalService(repo.DB, repo.WithdrawalRepo, repo.JournalRepo)
	adjustSvc := service.NewShopeeAdjustmentService(repo.DB, repo.ShopeeAdjustmentRepo, repo.JournalRepo)
	orderDetailSvc := service.NewOrderDetailService(repo.OrderDetailRepo)
	batchSvc := service.NewBatchService(repo.BatchRepo)

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
		dh := handlers.NewDropshipHandler(dropshipSvc, batchSvc)
		apiGroup.POST("/dropship/import", dh.HandleImport)
		apiGroup.GET("/dropship/purchases", dh.HandleList)
		apiGroup.GET("/dropship/purchases/summary", dh.HandleSum)
		apiGroup.GET("/dropship/purchases/daily", dh.HandleDailyTotals)
		apiGroup.GET("/dropship/purchases/monthly", dh.HandleMonthlyTotals)
		apiGroup.GET("/dropship/cancellations/summary", dh.HandleCancelledSummary)
		apiGroup.GET("/dropship/purchases/:id/details", dh.HandleListDetails)
		apiGroup.GET("/dropship/top-products", dh.HandleTopProducts)
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
		handlers.NewTaxHandler(taxSvc).Register(apiGroup)

		handlers.NewPLHandler(plSvc).RegisterRoutes(apiGroup)
		handlers.NewProfitLossReportHandler(plReportSvc).RegisterRoutes(apiGroup)
		handlers.NewGLHandler(glSvc).RegisterRoutes(apiGroup)
		handlers.NewReconcileExtraHandler(reconSvc).RegisterRoutes(apiGroup)
		handlers.NewPendingBalanceHandler(pbSvc).RegisterRoutes(apiGroup)
		handlers.NewWalletHandler(walletSvc).RegisterRoutes(apiGroup)
		handlers.NewAdsTopupHandler(adsTopupSvc).RegisterRoutes(apiGroup)
		handlers.NewWalletWithdrawalHandler(walletWdSvc).RegisterRoutes(apiGroup)
		handlers.NewAssetAccountHandler(assetSvc).RegisterRoutes(apiGroup)
		handlers.NewBatchHandler(batchSvc).RegisterRoutes(apiGroup)
		handlers.NewWithdrawHandler(shopeeSvc).RegisterRoutes(apiGroup)
		handlers.NewWithdrawalHandler(withdrawalSvc).RegisterRoutes(apiGroup)
		handlers.NewShopeeAdjustmentHandler(adjustSvc).RegisterRoutes(apiGroup)
		handlers.NewOrderDetailHandler(orderDetailSvc).RegisterRoutes(apiGroup)
		handlers.NewConfigHandler(cfg).RegisterRoutes(apiGroup)
		handlers.NewDashboardHandler().RegisterRoutes(apiGroup)
	}

	// 5) Start the HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		logutil.Fatalf("Server failed: %v", err)
	}
}
