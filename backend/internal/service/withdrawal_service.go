package service

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
	"github.com/xuri/excelize/v2"
)

// WithdrawalService handles import and creation of withdrawals.
type WithdrawalService struct {
	db          *sqlx.DB
	repo        *repository.WithdrawalRepo
	journalRepo *repository.JournalRepo
}

func NewWithdrawalService(db *sqlx.DB, r *repository.WithdrawalRepo, jr *repository.JournalRepo) *WithdrawalService {
	return &WithdrawalService{db: db, repo: r, journalRepo: jr}
}

func (s *WithdrawalService) List(ctx context.Context) ([]models.Withdrawal, error) {
	return s.repo.List(ctx, "date", "desc")
}

func (s *WithdrawalService) Create(ctx context.Context, w *models.Withdrawal) error {
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
		repo = repository.NewWithdrawalRepo(tx)
		jr = repository.NewJournalRepo(tx)
	}
	if w.CreatedAt.IsZero() {
		w.CreatedAt = time.Now()
	}
	if err := repo.Insert(ctx, w); err != nil {
		return err
	}
	if jr != nil {
		if err := createWithdrawalJournal(ctx, jr, w); err != nil {
			return err
		}
	}
	if tx != nil {
		return tx.Commit()
	}
	return nil
}

func createWithdrawalJournal(ctx context.Context, jr *repository.JournalRepo, w *models.Withdrawal) error {
	je := &models.JournalEntry{
		EntryDate:    w.Date,
		Description:  ptrString("Withdraw Shopee"),
		SourceType:   "withdrawal",
		SourceID:     fmt.Sprintf("%d", w.ID),
		ShopUsername: w.Store,
		Store:        w.Store,
		CreatedAt:    time.Now(),
	}
	jid, err := jr.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	lines := []models.JournalLine{
		{JournalID: jid, AccountID: 11014, IsDebit: true, Amount: w.Amount},
		{JournalID: jid, AccountID: saldoShopeeAccountID(w.Store), IsDebit: false, Amount: w.Amount},
	}
	for i := range lines {
		if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *WithdrawalService) ImportXLSX(ctx context.Context, r io.Reader) (int, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return 0, err
	}
	sheet := f.GetSheetList()[0]
	username, _ := f.GetCellValue(sheet, "B6")
	store := formatNamaToko(username)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return 0, err
	}
	headerRow := 17
	inserted := 0
	for i := headerRow; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 6 || row[0] == "Tanggal Transaksi" {
			continue
		}
		if strings.TrimSpace(row[1]) != "Penarikan Dana" {
			continue
		}
		t, err := time.Parse("2006-01-02 15:04:05", row[0])
		if err != nil {
			continue
		}
		amt, err := strconv.ParseFloat(row[5], 64)
		if err != nil {
			continue
		}
		w := &models.Withdrawal{Store: store, Date: t, Amount: -amt, CreatedAt: time.Now()}
		if err := s.Create(ctx, w); err != nil {
			return inserted, err
		}
		inserted++
	}
	return inserted, nil
}
