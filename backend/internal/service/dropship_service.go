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
}

// DropshipService handles CSV‐import and any Dropship‐related business logic.
type DropshipService struct {
	repo DropshipRepoInterface
}

// NewDropshipService constructs a DropshipService with the given repository.
func NewDropshipService(repo DropshipRepoInterface) *DropshipService {
	return &DropshipService{repo: repo}
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
func (s *DropshipService) ImportFromCSV(ctx context.Context, r io.Reader) error {
	reader := csv.NewReader(r)
	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	inserted := make(map[string]bool)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return fmt.Errorf("read row: %w", err)
		}

		qty, err := strconv.Atoi(record[8])
		if err != nil {
			return fmt.Errorf("parse qty '%s': %w", record[8], err)
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
			return fmt.Errorf("parse waktu_pesanan '%s': %w", record[1], err)
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

		if !inserted[header.KodePesanan] {
			if err := s.repo.InsertDropshipPurchase(ctx, header); err != nil {
				return fmt.Errorf("insert header %s: %w", header.KodePesanan, err)
			}
			inserted[header.KodePesanan] = true
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
			return fmt.Errorf("insert detail %s: %w", header.KodePesanan, err)
		}
	}
	return nil
}
