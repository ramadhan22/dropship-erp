// File: backend/cmd/api/main.go

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/ramadhan22/dropship-erp/backend/internal/cache"
	"github.com/ramadhan22/dropship-erp/backend/internal/config"
	"github.com/ramadhan22/dropship-erp/backend/internal/handlers"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/middleware"
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

	// 2) Initialize repositories (Postgres DB connection with optimized pool settings)
	repo, err := repository.NewPostgresRepository(cfg.Database.URL)
	if err != nil {
		logutil.Fatalf("DB connection failed: %v", err)
	}

	// Apply optimized connection pool settings
	repo.DB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	repo.DB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	if connLifetime := parseDuration(cfg.Database.ConnMaxLifetime, time.Hour); connLifetime > 0 {
		repo.DB.SetConnMaxLifetime(connLifetime)
	}
	if err := migrations.Run(repo.DB.DB); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Printf("DB migrations: %v", err)
		} else {
			logutil.Fatalf("DB migrations failed: %v", err)
		}
	}

	// 3) Initialize cache
	var cacheInstance cache.Cache
	if cfg.Cache.Enabled {
		log.Printf("Initializing Redis cache...")
		redisCache, err := cache.NewRedisCache(cache.CacheConfig{
			RedisURL:     cfg.Cache.RedisURL,
			Password:     cfg.Cache.Password,
			DB:           cfg.Cache.DB,
			MaxRetries:   cfg.Cache.MaxRetries,
			DialTimeout:  parseDuration(cfg.Cache.DialTimeout, 5*time.Second),
			ReadTimeout:  parseDuration(cfg.Cache.ReadTimeout, 3*time.Second),
			WriteTimeout: parseDuration(cfg.Cache.WriteTimeout, 3*time.Second),
			DefaultTTL:   parseDuration(cfg.Cache.DefaultTTL, 5*time.Minute),
		})
		if err != nil {
			log.Printf("Failed to initialize Redis cache, falling back to no-op cache: %v", err)
			cacheInstance = cache.NewNoopCache()
		} else {
			cacheInstance = redisCache
			log.Printf("Redis cache initialized successfully")
		}
	} else {
		log.Printf("Cache disabled, using no-op cache")
		cacheInstance = cache.NewNoopCache()
	}

	// 4) Initialize services with the appropriate repo interfaces
	shClient := service.NewShopeeClient(cfg.Shopee)
	batchSvc := service.NewBatchService(repo.BatchRepo, repo.BatchDetailRepo)
	dropshipSvc := service.NewDropshipService(
		repo.DB,
		repo.DropshipRepo,
		repo.JournalRepo,
		repo.ChannelRepo,
		repo.OrderDetailRepo,
		batchSvc,
		shClient,
		cacheInstance,
		cfg.MaxThreads,
		cfg.Performance.BatchSize,
	)
	// Initialize enhanced import components
	streamingConfig := service.DefaultStreamingImportConfig()
	streamingProcessor := service.NewStreamingImportProcessor(dropshipSvc, streamingConfig)

	// Initialize memory optimizer
	memoryOptimizer := service.NewMemoryOptimizer(1024, 10*time.Second) // 1GB max, check every 10s
	memoryOptimizer.StartMonitoring(context.Background())

	// Initialize enhanced scheduler
	enhancedScheduler := service.NewEnhancedImportScheduler(
		batchSvc, dropshipSvc, streamingProcessor, time.Minute, cfg.MaxThreads,
	)
	enhancedScheduler.Start()

	// Keep original scheduler for backward compatibility
	service.NewDropshipImportScheduler(batchSvc, dropshipSvc, time.Minute).Start(context.Background())
	shopeeSvc := service.NewShopeeService(repo.DB, repo.ShopeeRepo, repo.DropshipRepo, repo.JournalRepo, repo.ShopeeAdjustmentRepo, repo.ChannelRepo, cfg.Shopee)
	reconSvc := service.NewReconcileService(
		repo.DB,
		repo.DropshipRepo, repo.ShopeeRepo, repo.JournalRepo, repo.ReconcileRepo,
		repo.ChannelRepo,
		repo.OrderDetailRepo,
		repo.ShopeeAdjustmentRepo,
		shClient,
		batchSvc,
		repo.FailedReconciliationRepo,
		repo.ShippingDiscrepancyRepo,
		cfg.MaxThreads,
		nil, // Use default reconciliation config
	)
	service.NewReconcileBatchScheduler(batchSvc, reconSvc, time.Minute).Start(context.Background())

	// Start background scheduler for reconcile batch creation
	service.NewReconcileBatchCreationScheduler(batchSvc, reconSvc, time.Minute).Start(context.Background())

	// Start background scheduler for Shopee detail fetching
	shopeeDetailBgSvc := service.NewShopeeDetailBackgroundService(reconSvc, batchSvc, repo.OrderDetailRepo, repo.DropshipRepo, repo.ChannelRepo, shClient)
	service.NewShopeeDetailBackgroundScheduler(shopeeDetailBgSvc, time.Minute).Start(context.Background())
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
	pbSvc := service.NewPendingBalanceService(shClient, repo.ChannelRepo)
	walletSvc := service.NewWalletTransactionService(repo.ChannelRepo, shClient)
	adsTopupSvc := service.NewAdsTopupService(walletSvc, repo.JournalRepo)
	walletWdSvc := service.NewWalletWithdrawalService(walletSvc, repo.JournalRepo)
	assetSvc := service.NewAssetAccountService(repo.AssetAccountRepo, repo.JournalRepo)
	withdrawalSvc := service.NewWithdrawalService(repo.DB, repo.WithdrawalRepo, repo.JournalRepo)
	adjustSvc := service.NewShopeeAdjustmentService(repo.DB, repo.ShopeeAdjustmentRepo, repo.JournalRepo)
	orderDetailSvc := service.NewOrderDetailService(repo.OrderDetailRepo)
	shippingDiscrepancySvc := service.NewShippingDiscrepancyService(repo.DB, repo.ShippingDiscrepancyRepo)
	adsPerformanceSvc := service.NewAdsPerformanceService(repo.DB, cfg.Shopee, repo)
	adsPerformanceBatchScheduler := service.NewAdsPerformanceBatchScheduler(batchSvc, adsPerformanceSvc, time.Minute)
	adsPerformanceBatchScheduler.Start(context.Background())
	
	// Initialize forecast service (without obsolete shopee repo)
	forecastSvc := service.NewForecastService(
		repo.DropshipRepo, repo.JournalRepo,
	)
	// 4) Setup performance monitoring
	if cfg.Performance.EnableMetrics {
		// Set slow query threshold
		middleware.SetSlowQueryThreshold(parseDuration(cfg.Performance.SlowQueryThreshold, 2*time.Second))
	}

	// 5) Setup Gin router and API routes
	router := gin.Default()

	// Add performance monitoring middleware
	if cfg.Performance.EnableMetrics {
		router.Use(middleware.PerformanceMiddleware())
		router.Use(middleware.MetricsMiddleware())
	}
	// CORS configuration – origins can be configured via config.yaml
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.CorsOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Skip-Loading"},
		AllowCredentials: true,
		AllowOriginFunc:  func(origin string) bool { return true },
	}))

	apiGroup := router.Group("/api")
	{
		dh := handlers.NewDropshipHandler(dropshipSvc, batchSvc)
		apiGroup.POST("/dropship/import", dh.HandleImport)
		apiGroup.GET("/dropship/purchases", dh.HandleList)
		apiGroup.GET("/dropship/purchases/filtered", middleware.FilterMiddleware(), dh.HandleListFiltered)
		apiGroup.GET("/dropship/purchases/summary", dh.HandleSum)
		apiGroup.GET("/dropship/purchases/daily", dh.HandleDailyTotals)
		apiGroup.GET("/dropship/purchases/monthly", dh.HandleMonthlyTotals)
		apiGroup.GET("/dropship/cancellations/summary", dh.HandleCancelledSummary)
		apiGroup.GET("/dropship/purchases/:id/details", dh.HandleListDetails)
		apiGroup.GET("/dropship/top-products", dh.HandleTopProducts)

		// Enhanced bulk import endpoints
		bulkImportHandler := handlers.NewBulkImportHandler(dropshipSvc, batchSvc, streamingProcessor, enhancedScheduler)
		apiGroup.POST("/dropship/bulk-import", bulkImportHandler.HandleBulkImport)
		apiGroup.GET("/dropship/import-status/:batch_id", bulkImportHandler.HandleImportStatus)
		apiGroup.GET("/dropship/bulk-import-status", bulkImportHandler.HandleBulkImportStatus)
		apiGroup.POST("/dropship/force-process/:batch_id", bulkImportHandler.HandleForceProcessBatch)
		apiGroup.GET("/dropship/import-recommendations", bulkImportHandler.HandleImportRecommendations)

		// Memory and performance monitoring endpoints
		apiGroup.GET("/memory-stats", func(c *gin.Context) {
			stats := memoryOptimizer.GetMemoryStats()
			c.JSON(200, gin.H{
				"memory_stats":    stats,
				"recommendations": memoryOptimizer.GetMemoryRecommendations(),
			})
		})

		apiGroup.GET("/performance", func(c *gin.Context) {
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			c.JSON(200, gin.H{
				"system_stats": gin.H{
					"allocated_memory": memStats.Alloc,
					"total_allocated":  memStats.TotalAlloc,
					"system_memory":    memStats.Sys,
					"gc_count":         memStats.NumGC,
					"goroutines":       runtime.NumGoroutine(),
				},
				"streaming_stats": streamingProcessor.GetStats(),
				"queue_status":    enhancedScheduler.GetQueueStatus(),
			})
		})
		shHandler := handlers.NewShopeeHandler(shopeeSvc)
		apiGroup.POST("/shopee/import", shHandler.HandleImport)
		apiGroup.POST("/shopee/affiliate", shHandler.HandleImportAffiliate)
		apiGroup.POST("/shopee/settle/:order_sn", shHandler.HandleConfirmSettle)
		apiGroup.GET("/shopee/affiliate", shHandler.HandleListAffiliate)
		apiGroup.GET("/shopee/affiliate/summary", shHandler.HandleSumAffiliate)
		apiGroup.GET("/shopee/settled", shHandler.HandleListSettled)
		apiGroup.GET("/shopee/settled/:order_sn", shHandler.HandleGetSettleDetail)
		apiGroup.GET("/shopee/settled/summary", shHandler.HandleSumSettled)
		apiGroup.GET("/shopee/returns", shHandler.HandleGetReturnList)
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
		handlers.NewShippingDiscrepancyHandler(shippingDiscrepancySvc).RegisterRoutes(apiGroup)
		handlers.NewAdsPerformanceHandler(adsPerformanceSvc, adsPerformanceBatchScheduler).RegisterRoutes(apiGroup)
		handlers.NewConfigHandler(cfg).RegisterRoutes(apiGroup)
		dashSvc := service.NewDashboardService(repo.DropshipRepo, repo.JournalRepo, plReportSvc)
		handlers.NewDashboardHandler(dashSvc).RegisterRoutes(apiGroup)

		// Forecast endpoints
		forecastHandler := handlers.NewForecastHandler(forecastSvc)
		apiGroup.POST("/forecast/generate", forecastHandler.HandleGenerateForecast)
		apiGroup.GET("/forecast/params", forecastHandler.HandleGetForecastParams)
		apiGroup.GET("/forecast/summary", forecastHandler.HandleGetForecastSummary)

		// Performance metrics endpoint (system monitoring)
		// Note: /api/performance is already registered above at line 214
	}

	// 6) Start the HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		logutil.Fatalf("Server failed: %v", err)
	}
}

// parseDuration parses a duration string and returns a default value if parsing fails
func parseDuration(durationStr string, defaultValue time.Duration) time.Duration {
	if durationStr == "" {
		return defaultValue
	}
	if duration, err := time.ParseDuration(durationStr); err == nil {
		return duration
	}
	return defaultValue
}
