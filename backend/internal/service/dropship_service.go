// File: backend/internal/service/dropship_service.go

package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// DropshipRepoInterface defines the subset of DropshipRepo methods that the service needs.
// In production, you pass in *repository.DropshipRepo; in tests you pass a fake implementing this.
type DropshipRepoInterface interface {
	InsertDropshipPurchase(ctx context.Context, p *models.DropshipPurchase) error
	InsertDropshipPurchaseDetail(ctx context.Context, d *models.DropshipPurchaseDetail) error
	ExistsDropshipPurchase(ctx context.Context, kodePesanan string) (bool, error)
	ListDropshipPurchases(ctx context.Context, channel, store, date, month, year string, limit, offset int) ([]models.DropshipPurchase, int, error)
	SumDropshipPurchases(ctx context.Context, channel, store, date, month, year string) (float64, error)
	GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error)
	ListDropshipPurchaseDetails(ctx context.Context, kodePesanan string) ([]models.DropshipPurchaseDetail, error)
}

// DropshipService handles CSV‐import and any Dropship‐related business logic.
type DropshipJournalRepo interface {
	CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error)
	InsertJournalLine(ctx context.Context, l *models.JournalLine) error
}

// DropshipService handles CSV‐import and any Dropship‐related business logic.
type DropshipService struct {
	repo        DropshipRepoInterface
	journalRepo DropshipJournalRepo
}

// NewDropshipService constructs a DropshipService with the given repository.
func NewDropshipService(repo DropshipRepoInterface, jr DropshipJournalRepo) *DropshipService {
	return &DropshipService{repo: repo, journalRepo: jr}
}

// ImportFromCSV reads a Dumpsihp CSV file (with a header row) and inserts each purchase row.
// Expected CSV columns (example):
//
//	0: seller_username
//	1: purchase_id
//	2: order_id         (can be empty string if not linked yet)
//	3: sku
//	4: qty
//	5: purchase_price
//	6: purchase_fee
//	7: status
//	8: purchase_date    (YYYY-MM-DD)
//	9: supplier_name
//
// Any parse error aborts the import and returns it.
// ImportFromCSV inserts rows from a CSV reader and returns how many rows were inserted.
func (s *DropshipService) ImportFromCSV(ctx context.Context, r io.Reader) (int, error) {
	reader := csv.NewReader(r)
	if _, err := reader.Read(); err != nil {
		return 0, fmt.Errorf("read header: %w", err)
	}

	inserted := make(map[string]bool)
	skipped := make(map[string]bool)
	count := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return count, fmt.Errorf("read row: %w", err)
		}

		qty, err := strconv.Atoi(record[8])
		if err != nil {
			return count, fmt.Errorf("parse qty '%s': %w", record[8], err)
		}
		hargaProduk, _ := strconv.ParseFloat(record[7], 64)
		totalHargaProduk, _ := strconv.ParseFloat(record[9], 64)
		biayaLain, _ := strconv.ParseFloat(record[10], 64)
		biayaMitra, _ := strconv.ParseFloat(record[11], 64)
		totalTransaksi, _ := strconv.ParseFloat(record[12], 64)
		hargaChannel, _ := strconv.ParseFloat(record[13], 64)
		totalHargaChannel, _ := strconv.ParseFloat(record[14], 64)
		potensi, _ := strconv.ParseFloat(record[15], 64)

		waktuPesanan, err := time.Parse("02 January 2006, 15:04:05", record[1])
		if err != nil {
			return count, fmt.Errorf("parse waktu_pesanan '%s': %w", record[1], err)
		}
		waktuKirim, _ := time.Parse("02 January 2006, 15:04:05", record[24])

		header := &models.DropshipPurchase{
			KodePesanan:           record[3],
			KodeTransaksi:         record[4],
			WaktuPesananTerbuat:   waktuPesanan,
			StatusPesananTerakhir: record[2],
			BiayaLainnya:          biayaLain,
			BiayaMitraJakmall:     biayaMitra,
			TotalTransaksi:        totalTransaksi,
			DibuatOleh:            record[16],
			JenisChannel:          record[17],
			NamaToko:              record[18],
			KodeInvoiceChannel:    record[19],
			GudangPengiriman:      record[20],
			JenisEkspedisi:        record[21],
			Cashless:              record[22],
			NomorResi:             record[23],
			WaktuPengiriman:       waktuKirim,
			Provinsi:              record[25],
			Kota:                  record[26],
		}

		if !inserted[header.KodePesanan] && !skipped[header.KodePesanan] {
			exists, err := s.repo.ExistsDropshipPurchase(ctx, header.KodePesanan)
			if err != nil {
				return count, fmt.Errorf("check exists %s: %w", header.KodePesanan, err)
			}
			if exists {
				skipped[header.KodePesanan] = true
			} else {
				if err := s.repo.InsertDropshipPurchase(ctx, header); err != nil {
					return count, fmt.Errorf("insert header %s: %w", header.KodePesanan, err)
				}
				if err := s.createPendingSalesJournal(ctx, header); err != nil {
					return count, fmt.Errorf("journal %s: %w", header.KodePesanan, err)
				}
				inserted[header.KodePesanan] = true
			}
		}

		if skipped[header.KodePesanan] {
			continue
		}

		detail := &models.DropshipPurchaseDetail{
			KodePesanan:             header.KodePesanan,
			SKU:                     record[5],
			NamaProduk:              record[6],
			HargaProduk:             hargaProduk,
			Qty:                     qty,
			TotalHargaProduk:        totalHargaProduk,
			HargaProdukChannel:      hargaChannel,
			TotalHargaProdukChannel: totalHargaChannel,
			PotensiKeuntungan:       potensi,
		}
		if err := s.repo.InsertDropshipPurchaseDetail(ctx, detail); err != nil {
			return count, fmt.Errorf("insert detail %s: %w", header.KodePesanan, err)
		}
		count++
	}
	return count, nil
}

// ListDropshipPurchases proxies to the repository to fetch filtered purchases.
func (s *DropshipService) ListDropshipPurchases(
	ctx context.Context,
	channel, store, date, month, year string,
	limit, offset int,
) ([]models.DropshipPurchase, int, error) {
	return s.repo.ListDropshipPurchases(ctx, channel, store, date, month, year, limit, offset)
}

func (s *DropshipService) SumDropshipPurchases(
	ctx context.Context,
	channel, store, date, month, year string,
) (float64, error) {
	return s.repo.SumDropshipPurchases(ctx, channel, store, date, month, year)
}

func (s *DropshipService) GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error) {
	return s.repo.GetDropshipPurchaseByID(ctx, kodePesanan)
}

func (s *DropshipService) ListDropshipPurchaseDetails(ctx context.Context, kodePesanan string) ([]models.DropshipPurchaseDetail, error) {
	return s.repo.ListDropshipPurchaseDetails(ctx, kodePesanan)
}

func (s *DropshipService) createPendingSalesJournal(ctx context.Context, p *models.DropshipPurchase) error {
	if s.journalRepo == nil {
		return nil
	}
	je := &models.JournalEntry{
		EntryDate:    p.WaktuPesananTerbuat,
		Description:  ptrString("Pending sales " + p.KodePesanan),
		SourceType:   "pending_sales",
		SourceID:     p.KodePesanan,
		ShopUsername: p.NamaToko,
		Store:        p.NamaToko,
		CreatedAt:    time.Now(),
	}
	id, err := s.journalRepo.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}

	debit := pendingAccountID(p.NamaToko)
	credit := int64(4001)
	lines := []models.JournalLine{
		{
			JournalID: id,
			AccountID: debit,
			IsDebit:   true,
			Amount:    p.TotalTransaksi,
			Memo:      ptrString("Pending receivable " + p.KodePesanan),
		},
		{
			JournalID: id,
			AccountID: credit,
			IsDebit:   false,
			Amount:    p.TotalTransaksi,
			Memo:      ptrString("Sales " + p.KodePesanan),
		},
	}
	for i := range lines {
		if err := s.journalRepo.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}
	return nil
}

func pendingAccountID(store string) int64 {
	switch store {
	case "MR eStore Shopee":
		return 11010
	case "MR Barista Gear":
		return 11012
	default:
		return 11010
	}
}
