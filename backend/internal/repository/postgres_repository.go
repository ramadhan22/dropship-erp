package repository

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Repository aggregates all sub-repositories for the application.
// It holds the shared *sqlx.DB connection and each repo instance.
type Repository struct {
	DB                   *sqlx.DB
	BatchRepo            *BatchRepo
	BatchDetailRepo      *BatchDetailRepo
	DropshipRepo         *DropshipRepo
	ShopeeRepo           *ShopeeRepo
	ReconcileRepo        *ReconcileRepo
	JournalRepo          *JournalRepo

	ChannelRepo          *ChannelRepo
	AccountRepo          *AccountRepo
	AdInvoiceRepo        *AdInvoiceRepo
	AssetAccountRepo     *AssetAccountRepo
	WithdrawalRepo       *WithdrawalRepo
	TaxRepo              *TaxRepo
	ShopeeAdjustmentRepo *ShopeeAdjustmentRepo
	OrderDetailRepo      *OrderDetailRepo
}

// NewPostgresRepository connects to Postgres via sqlx and constructs all repos.
// databaseURL should be a valid DSN, e.g. "postgres://user:pass@host:port/dbname?sslmode=disable".
func NewPostgresRepository(databaseURL string) (*Repository, error) {
	// Connect using sqlx and configure the connection pool
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	// Limit the connection pool so imports don't exhaust Postgres slots
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Instantiate sub-repositories
	batchRepo := NewBatchRepo(db)
	batchDetailRepo := NewBatchDetailRepo(db)
	dropshipRepo := NewDropshipRepo(db)
	shopeeRepo := NewShopeeRepo(db)
	reconcileRepo := NewReconcileRepo(db)
	journalRepo := NewJournalRepo(db)

	channelRepo := NewChannelRepo(db)
	accountRepo := NewAccountRepo(db)
	adInvoiceRepo := NewAdInvoiceRepo(db)
	assetAccountRepo := NewAssetAccountRepo(db)
	withdrawalRepo := NewWithdrawalRepo(db)
	taxRepo := NewTaxRepo(db)
	adjustmentRepo := NewShopeeAdjustmentRepo(db)
	orderDetailRepo := NewOrderDetailRepo(db)

	return &Repository{
		DB:                   db,
		BatchRepo:            batchRepo,
		BatchDetailRepo:      batchDetailRepo,
		DropshipRepo:         dropshipRepo,
		ShopeeRepo:           shopeeRepo,
		ReconcileRepo:        reconcileRepo,
		JournalRepo:          journalRepo,

		ChannelRepo:          channelRepo,
		AccountRepo:          accountRepo,
		AdInvoiceRepo:        adInvoiceRepo,
		AssetAccountRepo:     assetAccountRepo,
		WithdrawalRepo:       withdrawalRepo,
		TaxRepo:              taxRepo,
		ShopeeAdjustmentRepo: adjustmentRepo,
		OrderDetailRepo:      orderDetailRepo,
	}, nil
}

// Close closes the underlying DB connection.
func (r *Repository) Close() error {
	return r.DB.Close()
}
