package repository

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Repository aggregates all sub-repositories for the application.
// It holds the shared *sqlx.DB connection and each repo instance.
type Repository struct {
	DB                   *sqlx.DB
	DropshipRepo         *DropshipRepo
	ShopeeRepo           *ShopeeRepo
	ReconcileRepo        *ReconcileRepo
	JournalRepo          *JournalRepo
	MetricRepo           *MetricRepo
	ChannelRepo          *ChannelRepo
	AccountRepo          *AccountRepo
	AdInvoiceRepo        *AdInvoiceRepo
	AssetAccountRepo     *AssetAccountRepo
	WithdrawalRepo       *WithdrawalRepo
	TaxRepo              *TaxRepo
	ShopeeAdjustmentRepo *ShopeeAdjustmentRepo
}

// NewPostgresRepository connects to Postgres via sqlx and constructs all repos.
// databaseURL should be a valid DSN, e.g. "postgres://user:pass@host:port/dbname?sslmode=disable".
func NewPostgresRepository(databaseURL string) (*Repository, error) {
	// Connect using sqlx
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	// Instantiate sub-repositories
	dropshipRepo := NewDropshipRepo(db)
	shopeeRepo := NewShopeeRepo(db)
	reconcileRepo := NewReconcileRepo(db)
	journalRepo := NewJournalRepo(db)
	metricRepo := NewMetricRepo(db)
	channelRepo := NewChannelRepo(db)
	accountRepo := NewAccountRepo(db)
	adInvoiceRepo := NewAdInvoiceRepo(db)
	assetAccountRepo := NewAssetAccountRepo(db)
	withdrawalRepo := NewWithdrawalRepo(db)
	taxRepo := NewTaxRepo(db)
	adjustmentRepo := NewShopeeAdjustmentRepo(db)

	return &Repository{
		DB:                   db,
		DropshipRepo:         dropshipRepo,
		ShopeeRepo:           shopeeRepo,
		ReconcileRepo:        reconcileRepo,
		JournalRepo:          journalRepo,
		MetricRepo:           metricRepo,
		ChannelRepo:          channelRepo,
		AccountRepo:          accountRepo,
		AdInvoiceRepo:        adInvoiceRepo,
		AssetAccountRepo:     assetAccountRepo,
		WithdrawalRepo:       withdrawalRepo,
		TaxRepo:              taxRepo,
		ShopeeAdjustmentRepo: adjustmentRepo,
	}, nil
}

// Close closes the underlying DB connection.
func (r *Repository) Close() error {
	return r.DB.Close()
}
