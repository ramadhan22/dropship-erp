package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/xuri/excelize/v2"
)

// ShopeeAdjustmentService imports adjustment data and records journals.
type AdjustmentRepo interface {
	Insert(ctx context.Context, a *models.ShopeeAdjustment) error
	Delete(ctx context.Context, order string, t time.Time, typ string) error
	List(ctx context.Context, from, to string) ([]models.ShopeeAdjustment, error)
	ListByOrder(ctx context.Context, order string) ([]models.ShopeeAdjustment, error)
	Get(ctx context.Context, id int64) (*models.ShopeeAdjustment, error)
	Update(ctx context.Context, a *models.ShopeeAdjustment) error
	DeleteByID(ctx context.Context, id int64) error
}

type ShopeeAdjustmentService struct {
	db          *sqlx.DB
	repo        AdjustmentRepo
	journalRepo ShopeeJournalRepo
}

func NewShopeeAdjustmentService(db *sqlx.DB, r AdjustmentRepo, jr ShopeeJournalRepo) *ShopeeAdjustmentService {
	return &ShopeeAdjustmentService{db: db, repo: r, journalRepo: jr}
}

func (s *ShopeeAdjustmentService) List(ctx context.Context, from, to string) ([]models.ShopeeAdjustment, error) {
	return s.repo.List(ctx, from, to)
}

func (s *ShopeeAdjustmentService) ImportXLSX(ctx context.Context, r io.Reader) (int, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return 0, err
	}
	sheet := ""
	for _, sh := range f.GetSheetList() {
		if strings.EqualFold(sh, "Adjustment") {
			sheet = sh
			break
		}
	}
	if sheet == "" {
		return 0, fmt.Errorf("sheet Adjustment not found")
	}
	username, _ := f.GetCellValue(sheet, "B2")
	store := formatNamaToko(username)
	rows, err := f.GetRows(sheet)
	if err != nil {
		return 0, err
	}
	start := 0
	for i, row := range rows {
		if len(row) > 0 && strings.Contains(strings.ToLower(row[0]), "rincian transaksi penyesuaian") {
			start = i + 2
			break
		}
	}
	inserted := 0
	for i := start; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 6 {
			continue
		}
		if strings.HasPrefix(strings.ToLower(row[0]), "total") {
			break
		}
		if strings.TrimSpace(row[0]) == "" {
			continue
		}
		t, err := parseDate(row[1])
		if err != nil {
			continue
		}
		amt, err := parseFloat(row[4])
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(row[2]), "bd marketing") || strings.Contains(strings.ToLower(row[3]), "bd marketing") {
			continue
		}
		adj := &models.ShopeeAdjustment{
			NamaToko:           store,
			TanggalPenyesuaian: t,
			TipePenyesuaian:    row[2],
			AlasanPenyesuaian:  row[3],
			BiayaPenyesuaian:   amt,
			NoPesanan:          row[5],
			CreatedAt:          time.Now(),
		}
		if err := s.repo.Delete(ctx, adj.NoPesanan, adj.TanggalPenyesuaian, adj.TipePenyesuaian); err != nil {
			return inserted, err
		}
		if s.journalRepo != nil {
			sid := fmt.Sprintf("%s-%s-%s", adj.NoPesanan, adj.TanggalPenyesuaian.Format("20060102"), sanitizeID(adj.TipePenyesuaian))
			if je, _ := s.journalRepo.GetJournalEntryBySource(ctx, "shopee_adjustment", sid); je != nil {
				_ = s.journalRepo.DeleteJournalEntry(ctx, je.JournalID)
			}
		}
		if err := s.repo.Insert(ctx, adj); err != nil {
			return inserted, err
		}
		if s.journalRepo != nil {
			if err := s.createJournal(ctx, s.journalRepo, adj); err != nil {
				return inserted, err
			}
		}
		inserted++
	}
	return inserted, nil
}

func (s *ShopeeAdjustmentService) createJournal(ctx context.Context, jr ShopeeJournalRepo, a *models.ShopeeAdjustment) error {
	je := &models.JournalEntry{
		EntryDate:    a.TanggalPenyesuaian,
		Description:  ptrString("Shopee adjustment " + a.NoPesanan),
		SourceType:   "shopee_adjustment",
		SourceID:     fmt.Sprintf("%s-%s-%s", a.NoPesanan, a.TanggalPenyesuaian.Format("20060102"), sanitizeID(a.TipePenyesuaian)),
		ShopUsername: a.NamaToko,
		Store:        a.NamaToko,
		CreatedAt:    time.Now(),
	}
	jid, err := jr.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	amt := a.BiayaPenyesuaian
	saldoAcc := saldoShopeeAccountID(a.NamaToko)
	if amt >= 0 {
		lines := []models.JournalLine{
			{JournalID: jid, AccountID: saldoAcc, IsDebit: true, Amount: amt},
			{JournalID: jid, AccountID: 4001, IsDebit: false, Amount: amt},
		}
		for i := range lines {
			if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
				return err
			}
		}
	} else {
		aamt := -amt
		lines := []models.JournalLine{
			{JournalID: jid, AccountID: 52002, IsDebit: true, Amount: aamt},
			{JournalID: jid, AccountID: saldoAcc, IsDebit: false, Amount: aamt},
		}
		for i := range lines {
			if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ShopeeAdjustmentService) Delete(ctx context.Context, id int64) error {
	adj, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteByID(ctx, id); err != nil {
		return err
	}
	if s.journalRepo != nil {
		sid := fmt.Sprintf("%s-%s-%s", adj.NoPesanan, adj.TanggalPenyesuaian.Format("20060102"), sanitizeID(adj.TipePenyesuaian))
		if je, _ := s.journalRepo.GetJournalEntryBySource(ctx, "shopee_adjustment", sid); je != nil {
			if err := s.journalRepo.DeleteJournalEntry(ctx, je.JournalID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *ShopeeAdjustmentService) Update(ctx context.Context, a *models.ShopeeAdjustment) error {
	old, err := s.repo.Get(ctx, a.ID)
	if err != nil {
		return err
	}
	if err := s.repo.Update(ctx, a); err != nil {
		return err
	}
	if s.journalRepo != nil {
		sid := fmt.Sprintf("%s-%s-%s", old.NoPesanan, old.TanggalPenyesuaian.Format("20060102"), sanitizeID(old.TipePenyesuaian))
		if je, _ := s.journalRepo.GetJournalEntryBySource(ctx, "shopee_adjustment", sid); je != nil {
			if err := s.journalRepo.DeleteJournalEntry(ctx, je.JournalID); err != nil {
				return err
			}
		}
		if err := s.createJournal(ctx, s.journalRepo, a); err != nil {
			return err
		}
	}
	return nil
}
