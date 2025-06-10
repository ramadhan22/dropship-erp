// File: backend/internal/models/models.go

package models

import "time"

// Account represents the D6 table: accounts
type Account struct {
	AccountID   int64     `db:"account_id"`
	AccountCode string    `db:"account_code"`
	AccountName string    `db:"account_name"`
	AccountType string    `db:"account_type"`
	ParentID    *int64    `db:"parent_id"` // NULLABLE
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// DropshipPurchase represents the D2 table: dropship_purchases
type DropshipPurchase struct {
	ID             int64     `db:"id"`
	SellerUsername string    `db:"seller_username"`
	PurchaseID     string    `db:"purchase_id"`
	OrderID        *string   `db:"order_id"` // NULLABLE
	SKU            string    `db:"sku"`
	Quantity       int       `db:"qty"`
	PurchasePrice  float64   `db:"purchase_price"`
	PurchaseFee    float64   `db:"purchase_fee"`
	Status         string    `db:"status"`
	PurchaseDate   time.Time `db:"purchase_date"`
	SupplierName   *string   `db:"supplier_name"` // NULLABLE
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// ShopeeSettledOrder represents the D1 table: shopee_settled_orders
type ShopeeSettledOrder struct {
	ID              int64     `db:"id"`
	OrderID         string    `db:"order_id"`
	NetIncome       float64   `db:"net_income"`
	ServiceFee      float64   `db:"service_fee"`
	CampaignFee     float64   `db:"campaign_fee"`
	CreditCardFee   float64   `db:"credit_card_fee"`
	ShippingSubsidy float64   `db:"shipping_subsidy"`
	TaxImportFee    float64   `db:"tax_and_import_fee"`
	SettledDate     time.Time `db:"settled_date"`
	SellerUsername  string    `db:"seller_username"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

// JournalEntry represents the D7 header table: journal_entries
type JournalEntry struct {
	JournalID    int64     `db:"journal_id"`
	EntryDate    time.Time `db:"entry_date"`
	Description  *string   `db:"description"` // NULLABLE
	SourceType   string    `db:"source_type"`
	SourceID     string    `db:"source_id"`
	ShopUsername string    `db:"shop_username"`
	CreatedAt    time.Time `db:"created_at"`
}

// JournalLine represents the D7 detail table: journal_lines
type JournalLine struct {
	LineID    int64   `db:"line_id"`
	JournalID int64   `db:"journal_id"` // FK → journal_entries(journal_id)
	AccountID int64   `db:"account_id"` // FK → accounts(account_id)
	IsDebit   bool    `db:"is_debit"`
	Amount    float64 `db:"amount"`
	Memo      *string `db:"memo"` // NULLABLE
}

// ReconciledTransaction represents the D3 table: reconciled_transactions
type ReconciledTransaction struct {
	ID           int64     `db:"id"`
	ShopUsername string    `db:"shop_username"`
	DropshipID   *string   `db:"dropship_id"` // NULLABLE
	ShopeeID     *string   `db:"shopee_id"`   // NULLABLE
	Status       string    `db:"status"`      // e.g., "matched", "unmatched"
	MatchedAt    time.Time `db:"matched_at"`
}

// CachedMetric represents the D5 table: cached_metrics
type CachedMetric struct {
	ID                int64     `db:"id"`
	ShopUsername      string    `db:"shop_username"`
	Period            string    `db:"period"` // e.g., "2025-05"
	SumRevenue        float64   `db:"sum_revenue"`
	SumCOGS           float64   `db:"sum_cogs"`
	SumFees           float64   `db:"sum_fees"`
	NetProfit         float64   `db:"net_profit"`
	EndingCashBalance float64   `db:"ending_cash_balance"`
	UpdatedAt         time.Time `db:"updated_at"`
}
