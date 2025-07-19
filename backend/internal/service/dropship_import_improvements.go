// File: backend/internal/service/dropship_import_improvements.go

package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// OrderGroup represents a complete order with header and all its details
type OrderGroup struct {
	Header  *models.DropshipPurchase
	Details []*models.DropshipPurchaseDetail
	Records [][]string // original CSV records for this order
}

// groupRecordsByOrder groups CSV records by kode_pesanan and creates OrderGroup objects
func groupRecordsByOrder(allRecords [][]string, channel string) (map[string]*OrderGroup, error) {
	groups := make(map[string]*OrderGroup)
	
	for _, record := range allRecords {
		if len(record) < 27 {
			return nil, fmt.Errorf("insufficient columns in CSV record: expected at least 27, got %d", len(record))
		}
		
		kodePesanan := record[3]
		if kodePesanan == "" {
			continue // skip empty order codes
		}
		
		// Parse header fields once per order
		if groups[kodePesanan] == nil {
			waktuPesanan, err := time.Parse("02 January 2006, 15:04:05", record[1])
			if err != nil {
				return nil, fmt.Errorf("parse waktu_pesanan for %s: %w", kodePesanan, err)
			}
			
			biayaLain, _ := strconv.ParseFloat(record[10], 64)
			biayaMitra, _ := strconv.ParseFloat(record[11], 64)
			totalTransaksi, _ := strconv.ParseFloat(record[12], 64)
			waktuKirim, _ := time.Parse("02 January 2006, 15:04:05", record[24])
			
			header := &models.DropshipPurchase{
				KodePesanan:           kodePesanan,
				KodeTransaksi:         record[4],
				WaktuPesananTerbuat:   waktuPesanan,
				StatusPesananTerakhir: record[2],
				BiayaLainnya:          biayaLain,
				BiayaMitraJakmall:     biayaMitra,
				TotalTransaksi:        totalTransaksi,
				DibuatOleh:            record[16],
				JenisChannel:          record[17],
				NamaToko:              record[18],
				KodeInvoiceChannel:    strings.TrimPrefix(record[19], "'"),
				GudangPengiriman:      record[20],
				JenisEkspedisi:        record[21],
				Cashless:              record[22],
				NomorResi:             record[23],
				WaktuPengiriman:       waktuKirim,
				Provinsi:              record[25],
				Kota:                  record[26],
			}
			
			// Apply channel filter
			if channel != "" && header.JenisChannel != channel {
				continue
			}
			
			groups[kodePesanan] = &OrderGroup{
				Header:  header,
				Details: []*models.DropshipPurchaseDetail{},
				Records: [][]string{},
			}
		}
		
		// Skip if channel filter doesn't match
		if channel != "" && record[17] != channel {
			continue
		}
		
		// Parse detail fields for each record
		qty, err := strconv.Atoi(record[8])
		if err != nil {
			return nil, fmt.Errorf("parse qty for %s: %w", kodePesanan, err)
		}
		
		hargaProduk, _ := strconv.ParseFloat(record[7], 64)
		totalHargaProduk, _ := strconv.ParseFloat(record[9], 64)
		hargaChannel, _ := strconv.ParseFloat(record[13], 64)
		totalHargaChannel, _ := strconv.ParseFloat(record[14], 64)
		potensi, _ := strconv.ParseFloat(record[15], 64)
		
		detail := &models.DropshipPurchaseDetail{
			KodePesanan:             kodePesanan,
			SKU:                     record[5],
			NamaProduk:              record[6],
			HargaProduk:             hargaProduk,
			Qty:                     qty,
			TotalHargaProduk:        totalHargaProduk,
			HargaProdukChannel:      hargaChannel,
			TotalHargaProdukChannel: totalHargaChannel,
			PotensiKeuntungan:       potensi,
		}
		
		groups[kodePesanan].Details = append(groups[kodePesanan].Details, detail)
		groups[kodePesanan].Records = append(groups[kodePesanan].Records, record)
	}
	
	return groups, nil
}

// processOrderDetails processes all detail records for a single order, handling individual failures gracefully
func (s *DropshipService) processOrderDetails(ctx context.Context, repoTx DropshipRepoInterface, header *models.DropshipPurchase, allDetails []*models.DropshipPurchaseDetail, batchID int64) (int, error) {
	var successCount int
	var firstError error
	
	for _, detail := range allDetails {
		if err := repoTx.InsertDropshipPurchaseDetail(ctx, detail); err != nil {
			logutil.Errorf("Failed to insert detail for order %s, SKU %s: %v", header.KodePesanan, detail.SKU, err)
			if s.batchSvc != nil && batchID != 0 {
				d := &models.BatchHistoryDetail{
					BatchID:   batchID,
					Reference: header.KodeInvoiceChannel,
					Store:     header.NamaToko,
					Status:    "failed",
					ErrorMsg:  fmt.Sprintf("Detail insert failed for SKU %s: %v", detail.SKU, err),
				}
				_ = s.batchSvc.CreateDetail(ctx, d)
			}
			// Store first error but continue processing other details
			if firstError == nil {
				firstError = err
			}
			continue
		}
		successCount++
	}
	
	log.Printf("Processed %d/%d details for order %s", successCount, len(allDetails), header.KodePesanan)
	
	// Return first error if any, but don't fail the entire import for partial success
	return successCount, firstError
}

// validateTransactionTotals validates that total_transaksi equals the sum of all product totals plus fees
func validateTransactionTotals(header *models.DropshipPurchase, details []*models.DropshipPurchaseDetail) error {
	var productTotal float64
	for _, detail := range details {
		productTotal += detail.TotalHargaProduk
	}
	
	expectedTotal := productTotal + header.BiayaLainnya + header.BiayaMitraJakmall
	actualTotal := header.TotalTransaksi
	
	// Allow for small floating point precision differences
	tolerance := 0.01
	diff := actualTotal - expectedTotal
	if diff < -tolerance || diff > tolerance {
		return fmt.Errorf("total transaction validation failed: expected %.2f (products: %.2f + biaya_lain: %.2f + biaya_mitra: %.2f), got %.2f", 
			expectedTotal, productTotal, header.BiayaLainnya, header.BiayaMitraJakmall, actualTotal)
	}
	
	return nil
}