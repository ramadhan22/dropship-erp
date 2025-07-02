package service

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

const (
	taxExpenseAcctID int64 = 54001 // PPh Final UMKM expense account id
	bankAcctID       int64 = 11002 // Bank account id
)

// TaxService handles UMKM tax payments.
type TaxRepoInterface interface {
	Get(ctx context.Context, store, periodType, periodValue string) (*models.TaxPayment, error)
	Create(ctx context.Context, tp *models.TaxPayment) error
	MarkPaid(ctx context.Context, id string, paidAt time.Time) error
}

type RevenueFetcher interface {
	GetRevenue(ctx context.Context, store, periodType, periodValue string) (float64, error)
}

type TaxService struct {
	db          *sqlx.DB
	repo        TaxRepoInterface
	journalRepo JournalRepoInterface
	metricSvc   RevenueFetcher
}

func NewTaxService(db *sqlx.DB, repo TaxRepoInterface, jr JournalRepoInterface, metricSvc RevenueFetcher) *TaxService {
	return &TaxService{db: db, repo: repo, journalRepo: jr, metricSvc: metricSvc}
}

func (s *TaxService) ComputeTax(ctx context.Context, store, periodType, periodValue string) (*models.TaxPayment, error) {
	rev, err := s.metricSvc.GetRevenue(ctx, store, periodType, periodValue)
	if err != nil {
		return nil, err
	}
	rate := 0.005
	amt := rev * rate
	tp := &models.TaxPayment{
		Store:       store,
		PeriodType:  periodType,
		PeriodValue: periodValue,
		Revenue:     rev,
		TaxRate:     rate,
		TaxAmount:   amt,
	}
	return tp, nil
}

func (s *TaxService) PayTax(ctx context.Context, tp *models.TaxPayment) error {
	var tx *sqlx.Tx
	repo := s.repo
	jr := s.journalRepo
	if s.db != nil {
		var err error
		tx, err = s.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		repo = repository.NewTaxRepo(tx)
		jr = repository.NewJournalRepo(tx)
	}

	if tp.ID == "" {
		if err := repo.Create(ctx, tp); err != nil {
			return err
		}
	}

	tp.IsPaid = true
	tp.PaidAt = time.Now()

	je := &models.JournalEntry{
		EntryDate:    tp.PaidAt,
		Description:  ptrString("UMKM Tax " + tp.PeriodValue),
		SourceType:   "tax_payment",
		SourceID:     tp.ID,
		ShopUsername: tp.Store,
		Store:        tp.Store,
		CreatedAt:    time.Now(),
	}
	jid, err := jr.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	lines := []models.JournalLine{
		{JournalID: jid, AccountID: taxExpenseAcctID, IsDebit: true, Amount: tp.TaxAmount},
		{JournalID: jid, AccountID: bankAcctID, IsDebit: false, Amount: tp.TaxAmount},
	}
	for i := range lines {
		if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}
	if err := repo.MarkPaid(ctx, tp.ID, tp.PaidAt); err != nil {
		return err
	}
	if tx != nil {
		return tx.Commit()
	}
	return nil
}
