// File: backend/internal/service/dropship_service.go

package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// DropshipRepoInterface defines the subset of DropshipRepo methods that the service needs.
// In production, you pass in *repository.DropshipRepo; in tests you pass a fake implementing this.
type DropshipRepoInterface interface {
	InsertDropshipPurchase(ctx context.Context, p *models.DropshipPurchase) error
	InsertDropshipPurchaseDetail(ctx context.Context, d *models.DropshipPurchaseDetail) error
	ExistsDropshipPurchase(ctx context.Context, kodePesanan string) (bool, error)
	ListExistingPurchases(ctx context.Context, ids []string) (map[string]bool, error)
	ListDropshipPurchases(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.DropshipPurchase, int, error)
	SumDropshipPurchases(ctx context.Context, channel, store, from, to string) (float64, error)
	GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error)
	ListDropshipPurchaseDetails(ctx context.Context, kodePesanan string) ([]models.DropshipPurchaseDetail, error)
	TopProducts(ctx context.Context, channel, store, from, to string, limit int) ([]models.ProductSales, error)
	DailyTotals(ctx context.Context, channel, store, from, to string) ([]repository.DailyPurchaseTotal, error)
	MonthlyTotals(ctx context.Context, channel, store, from, to string) ([]repository.MonthlyPurchaseTotal, error)
	CancelledSummary(ctx context.Context, channel, store, from, to string) (repository.CancelledSummary, error)
}

// DropshipService handles CSV‐import and any Dropship‐related business logic.
type DropshipJournalRepo interface {
	CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error)
	InsertJournalLine(ctx context.Context, l *models.JournalLine) error
}

// DropshipServiceStoreRepo provides access to store credentials used when
// fetching Shopee order details.
type DropshipServiceStoreRepo interface {
	GetStoreByName(ctx context.Context, name string) (*models.Store, error)
	UpdateStore(ctx context.Context, s *models.Store) error
}

// DropshipServiceDetailRepo persists raw Shopee order details and items.
type DropshipServiceDetailRepo interface {
	SaveOrderDetail(ctx context.Context, detail *models.ShopeeOrderDetailRow, items []models.ShopeeOrderItemRow, packages []models.ShopeeOrderPackageRow) error
}

// DropshipService handles CSV‐import and any Dropship‐related business logic.
// DropshipService handles CSV-import and Shopee integration for Dropship
// purchases. It persists purchase rows and posts pending sales journal entries.
// When configured with a Shopee client and store repositories it also fetches
// order details from Shopee.
type DropshipService struct {
	db          *sqlx.DB
	repo        DropshipRepoInterface
	journalRepo DropshipJournalRepo
	storeRepo   DropshipServiceStoreRepo
	detailRepo  DropshipServiceDetailRepo
	client      *ShopeeClient
}

// NewDropshipService constructs a DropshipService with the given repository.
func NewDropshipService(
	db *sqlx.DB,
	repo DropshipRepoInterface,
	jr DropshipJournalRepo,
	sr DropshipServiceStoreRepo,
	dr DropshipServiceDetailRepo,
	c *ShopeeClient,
) *DropshipService {
	return &DropshipService{
		db:          db,
		repo:        repo,
		journalRepo: jr,
		storeRepo:   sr,
		detailRepo:  dr,
		client:      c,
	}
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
func (s *DropshipService) ImportFromCSV(ctx context.Context, r io.Reader, channel string) (int, error) {
	log.Printf("ImportFromCSV channel=%s", channel)
	reader := csv.NewReader(r)
	reader.ReuseRecord = true
	if _, err := reader.Read(); err != nil {
		logutil.Errorf("ImportFromCSV header error: %v", err)
		return 0, fmt.Errorf("read header: %w", err)
	}

	var tx *sqlx.Tx
	repoTx := s.repo
	jrTx := s.journalRepo
	if s.db != nil {
		var err error
		tx, err = s.db.BeginTxx(ctx, nil)
		if err != nil {
			logutil.Errorf("ImportFromCSV tx begin error: %v", err)
			return 0, err
		}
		defer tx.Rollback()
		repoTx = repository.NewDropshipRepo(tx)
		jrTx = repository.NewJournalRepo(tx)
	}

	inserted := make(map[string]bool)
	skipped := make(map[string]bool)
	fetched := make(map[string]bool)
	var allRecords [][]string
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			logutil.Errorf("ImportFromCSV read row error: %v", err)
			return 0, fmt.Errorf("read row: %w", err)
		}
		cp := make([]string, len(rec))
		copy(cp, rec)
		allRecords = append(allRecords, cp)
	}
	unique := make(map[string]bool)
	for _, rec := range allRecords {
		if len(rec) > 3 {
			unique[rec[3]] = true
		}
	}
	ids := make([]string, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}
	existing, err := repoTx.ListExistingPurchases(ctx, ids)
	if err != nil {
		return 0, err
	}

	batches := make(map[string][]*models.DropshipPurchase)
	for _, record := range allRecords {
		h := &models.DropshipPurchase{
			KodePesanan:        record[3],
			NamaToko:           record[18],
			KodeInvoiceChannel: record[19],
			JenisChannel:       record[17],
		}
		if channel != "" && h.JenisChannel != channel {
			continue
		}
		if existing[h.KodePesanan] || fetched[h.KodePesanan] {
			continue
		}
		fetched[h.KodePesanan] = true
		batches[h.NamaToko] = append(batches[h.NamaToko], h)
	}

	apiTotals := make(map[string]float64)
	for store, list := range batches {
		for i := 0; i < len(list); i += 50 {
			end := i + 50
			if end > len(list) {
				end = len(list)
			}
			amtMap, err := s.fetchAndStoreDetailBatch(ctx, list[i:end])
			if err != nil {
				log.Printf("fetch batch detail store %s: %v", store, err)
				for _, h := range list[i:end] {
					skipped[h.KodePesanan] = true
				}
				continue
			}
			for k, v := range amtMap {
				apiTotals[k] = v
			}
		}
	}
	// track the header for each newly inserted purchase so we can
	// create the journal entry after all rows are processed
	headersMap := make(map[string]*models.DropshipPurchase)
	// accumulate product totals per purchase across multiple rows
	type totals struct {
		prod      float64
		prodCh    float64
		apiAmount float64
	}
	agg := make(map[string]*totals)
	count := 0

	for _, record := range allRecords {

		qty, err := strconv.Atoi(record[8])
		if err != nil {
			logutil.Errorf("ImportFromCSV parse qty error: %v", err)
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

		if channel != "" && header.JenisChannel != channel {
			continue
		}

		if !inserted[header.KodePesanan] && !skipped[header.KodePesanan] {
			if existing[header.KodePesanan] {
				skipped[header.KodePesanan] = true
				continue
			}

			apiAmt := apiTotals[header.KodePesanan]

			if err := repoTx.InsertDropshipPurchase(ctx, header); err != nil {
				return count, fmt.Errorf("insert header %s: %w", header.KodePesanan, err)
			}
			inserted[header.KodePesanan] = true
			headersMap[header.KodePesanan] = header
			if apiAmt > 0 {
				t := agg[header.KodePesanan]
				if t == nil {
					t = &totals{}
					agg[header.KodePesanan] = t
				}
				t.apiAmount = apiAmt
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
		if err := repoTx.InsertDropshipPurchaseDetail(ctx, detail); err != nil {
			return count, fmt.Errorf("insert detail %s: %w", header.KodePesanan, err)
		}
		// accumulate totals for journal creation later
		t, ok := agg[header.KodePesanan]
		if !ok {
			t = &totals{}
			agg[header.KodePesanan] = t
		}
		t.prod += totalHargaProduk
		t.prodCh += totalHargaChannel
		count++
	}
	// after processing all rows, create journal entries using summed totals
	for kode := range inserted {
		h := headersMap[kode]
		sum := agg[kode]
		var prod, prodCh, apiAmt float64
		if sum != nil {
			prod = sum.prod
			prodCh = sum.prodCh
			apiAmt = sum.apiAmount
		}
		pending := prodCh
		if apiAmt > 0 {
			pending = apiAmt
		}
		if err := s.createPendingSalesJournal(ctx, jrTx, h, prod, pending); err != nil {
			return count, fmt.Errorf("journal %s: %w", kode, err)
		}
	}
	if tx != nil {
		if err := tx.Commit(); err != nil {
			return count, err
		}
		log.Printf("ImportFromCSV committed %d rows", count)
	}
	log.Printf("ImportFromCSV done count=%d", count)
	return count, nil
}

// ListDropshipPurchases proxies to the repository to fetch filtered purchases.
func (s *DropshipService) ListDropshipPurchases(
	ctx context.Context,
	channel, store, from, to, orderNo, sortBy, dir string,
	limit, offset int,
) ([]models.DropshipPurchase, int, error) {
	return s.repo.ListDropshipPurchases(ctx, channel, store, from, to, orderNo, sortBy, dir, limit, offset)
}

func (s *DropshipService) SumDropshipPurchases(
	ctx context.Context,
	channel, store, from, to string,
) (float64, error) {
	return s.repo.SumDropshipPurchases(ctx, channel, store, from, to)
}

func (s *DropshipService) GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error) {
	return s.repo.GetDropshipPurchaseByID(ctx, kodePesanan)
}

func (s *DropshipService) ListDropshipPurchaseDetails(ctx context.Context, kodePesanan string) ([]models.DropshipPurchaseDetail, error) {
	return s.repo.ListDropshipPurchaseDetails(ctx, kodePesanan)
}

func (s *DropshipService) TopProducts(ctx context.Context, channel, store, from, to string, limit int) ([]models.ProductSales, error) {
	return s.repo.TopProducts(ctx, channel, store, from, to, limit)
}

func (s *DropshipService) DailyTotals(ctx context.Context, channel, store, from, to string) ([]repository.DailyPurchaseTotal, error) {
	return s.repo.DailyTotals(ctx, channel, store, from, to)
}

func (s *DropshipService) MonthlyTotals(ctx context.Context, channel, store, from, to string) ([]repository.MonthlyPurchaseTotal, error) {
	return s.repo.MonthlyTotals(ctx, channel, store, from, to)
}

func (s *DropshipService) CancelledSummary(ctx context.Context, channel, store, from, to string) (repository.CancelledSummary, error) {
	return s.repo.CancelledSummary(ctx, channel, store, from, to)
}

// fetchAndStoreDetail retrieves Shopee order detail and persists it. It returns
// the summed original item prices multiplied by quantities for journal posting.
func (s *DropshipService) fetchAndStoreDetail(ctx context.Context, header *models.DropshipPurchase) (float64, error) {
	if s.storeRepo == nil || s.client == nil {
		return 0, fmt.Errorf("shopee client not configured")
	}
	st, err := s.storeRepo.GetStoreByName(ctx, header.NamaToko)
	if err != nil || st == nil || st.AccessToken == nil || st.ShopID == nil {
		return 0, fmt.Errorf("fetch store %s: %w", header.NamaToko, err)
	}
	if err := s.ensureStoreTokenValid(ctx, st); err != nil {
		return 0, err
	}
	det, err := s.client.FetchShopeeOrderDetail(ctx, *st.AccessToken, *st.ShopID, header.KodeInvoiceChannel)
	if err != nil && strings.Contains(err.Error(), "invalid_access_token") {
		if e := s.ensureStoreTokenValid(ctx, st); e == nil {
			det, err = s.client.FetchShopeeOrderDetail(ctx, *st.AccessToken, *st.ShopID, header.KodeInvoiceChannel)
		}
	}
	if err != nil {
		return 0, err
	}
	var total float64
	row, items, packages := normalizeOrderDetail(header.KodeInvoiceChannel, header.NamaToko, *det)
	for _, it := range items {
		if it.ModelOriginalPrice != nil && it.ModelQuantityPurchased != nil {
			total += *it.ModelOriginalPrice * float64(*it.ModelQuantityPurchased)
		}
	}
	if s.detailRepo != nil {
		if err := s.detailRepo.SaveOrderDetail(ctx, row, items, packages); err != nil {
			log.Printf("save order detail %s: %v", header.KodeInvoiceChannel, err)
		}
	}
	return total, nil
}

// fetchAndStoreDetailBatch retrieves Shopee order details for multiple orders in a single API call.
// It returns a map keyed by order_sn with summed original item prices for journal posting.
func (s *DropshipService) fetchAndStoreDetailBatch(ctx context.Context, headers []*models.DropshipPurchase) (map[string]float64, error) {
	res := make(map[string]float64)
	if len(headers) == 0 {
		return res, nil
	}
	if s.storeRepo == nil || s.client == nil {
		return nil, fmt.Errorf("shopee client not configured")
	}
	storeName := headers[0].NamaToko
	st, err := s.storeRepo.GetStoreByName(ctx, storeName)
	if err != nil || st == nil || st.AccessToken == nil || st.ShopID == nil {
		return nil, fmt.Errorf("fetch store %s: %w", storeName, err)
	}
	if err := s.ensureStoreTokenValid(ctx, st); err != nil {
		return nil, err
	}

	sns := make([]string, len(headers))
	hdrMap := make(map[string]*models.DropshipPurchase, len(headers))
	for i, h := range headers {
		sns[i] = h.KodeInvoiceChannel
		hdrMap[h.KodeInvoiceChannel] = h
	}

	details, err := s.client.FetchShopeeOrderDetails(ctx, *st.AccessToken, *st.ShopID, sns)
	if err != nil && strings.Contains(err.Error(), "invalid_access_token") {
		if e := s.ensureStoreTokenValid(ctx, st); e == nil {
			details, err = s.client.FetchShopeeOrderDetails(ctx, *st.AccessToken, *st.ShopID, sns)
		}
	}
	if err != nil {
		return nil, err
	}

	for _, det := range details {
		snVal, _ := det["order_sn"].(string)
		h := hdrMap[snVal]
		row, items, packages := normalizeOrderDetail(snVal, storeName, det)
		var total float64
		for _, it := range items {
			if it.ModelOriginalPrice != nil && it.ModelQuantityPurchased != nil {
				total += *it.ModelOriginalPrice * float64(*it.ModelQuantityPurchased)
			}
		}
		if s.detailRepo != nil {
			if err := s.detailRepo.SaveOrderDetail(ctx, row, items, packages); err != nil {
				log.Printf("save order detail %s: %v", h.KodeInvoiceChannel, err)
			}
		}
		if h != nil {
			res[h.KodePesanan] = total
		}
	}
	return res, nil
}

// ensureStoreTokenValid refreshes the access token for the given store if it has
// expired. The updated token is saved back to the storeRepo.
func (s *DropshipService) ensureStoreTokenValid(ctx context.Context, st *models.Store) error {
	if s.client == nil || s.storeRepo == nil {
		return fmt.Errorf("missing client or store repo")
	}
	log.Printf("ensureStoreTokenValid for store %d", st.StoreID)
	loc, _ := time.LoadLocation("Asia/Jakarta")
	reinterpreted := time.Date(
		st.LastUpdated.Year(), st.LastUpdated.Month(), st.LastUpdated.Day(),
		st.LastUpdated.Hour(), st.LastUpdated.Minute(), st.LastUpdated.Second(), st.LastUpdated.Nanosecond(),
		loc,
	)
	exp := reinterpreted.Add(time.Duration(*st.ExpireIn) * time.Second)
	if st.RefreshToken == nil {
		log.Fatalf("ensureStoreTokenValid: missing refresh token for store %d", st.StoreID)
		return fmt.Errorf("missing refresh token")
	}
	if st.ShopID == nil || *st.ShopID == "" {
		log.Fatalf("ensureStoreTokenValid: missing shop id for store %d", st.StoreID)
		return fmt.Errorf("missing shop id")
	}
	if st.ExpireIn != nil && st.LastUpdated != nil {
		if time.Now().Before(exp.Local()) {
			log.Printf("Token for store %d is still valid until %v", st.StoreID, exp)
			return nil
		}
	}
	s.client.ShopID = *st.ShopID
	s.client.RefreshToken = *st.RefreshToken
	resp, err := s.client.RefreshAccessToken(ctx)
	if err != nil {
		return err
	}
	st.AccessToken = &resp.Response.AccessToken
	if resp.Response.RefreshToken != "" {
		st.RefreshToken = &resp.Response.RefreshToken
	}
	st.ExpireIn = &resp.Response.ExpireIn
	st.RequestID = &resp.Response.RequestID
	now := time.Now()
	st.LastUpdated = &now
	if uerr := s.storeRepo.UpdateStore(ctx, st); uerr != nil {
		log.Printf("update store token: %v", uerr)
	}
	return nil
}

// createPendingSalesJournal records pending receivable and sales using the
// amounts derived from Shopee order detail when available.
func (s *DropshipService) createPendingSalesJournal(ctx context.Context, jr DropshipJournalRepo, p *models.DropshipPurchase, totalProduk, pendingAmount float64) error {
	if jr == nil {
		return nil
	}
	je := &models.JournalEntry{
		EntryDate:    p.WaktuPesananTerbuat,
		Description:  ptrString("Pending sales " + p.KodeInvoiceChannel),
		SourceType:   "pending_sales",
		SourceID:     p.KodeInvoiceChannel,
		ShopUsername: p.NamaToko,
		Store:        p.NamaToko,
		CreatedAt:    time.Now(),
	}
	id, err := jr.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}

	debit := pendingAccountID(p.NamaToko)
	credit := int64(4001)
	jakmall := int64(11009)
	cogs := int64(5001)
	mitra := int64(52007)

	saldoJakmall := totalProduk + p.BiayaMitraJakmall
	lines := []models.JournalLine{
		{
			JournalID: id,
			AccountID: jakmall,
			IsDebit:   false,
			Amount:    saldoJakmall,
			Memo:      ptrString("Saldo Jakmall " + p.KodeInvoiceChannel),
		},
		{
			JournalID: id,
			AccountID: cogs,
			IsDebit:   true,
			Amount:    totalProduk,
			Memo:      ptrString("HPP " + p.KodeInvoiceChannel),
		},
		{
			JournalID: id,
			AccountID: mitra,
			IsDebit:   true,
			Amount:    p.BiayaMitraJakmall,
			Memo:      ptrString("Biaya Mitra Jakmall " + p.KodeInvoiceChannel),
		},
		{
			JournalID: id,
			AccountID: debit,
			IsDebit:   true,
			Amount:    pendingAmount,
			Memo:      ptrString("Pending receivable " + p.KodeInvoiceChannel),
		},
		{
			JournalID: id,
			AccountID: credit,
			IsDebit:   false,
			Amount:    pendingAmount,
			Memo:      ptrString("Sales " + p.KodeInvoiceChannel),
		},
		// previously additional HPP and Saldo Jakmall journal lines were
		// inserted using the order code. These duplicated the invoice
		// lines above and caused double recognition of the accounts.
	}
	for i := range lines {
		if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
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

func saldoShopeeAccountID(store string) int64 {
	switch store {
	case "MR eStore Shopee":
		return 11011
	case "MR Barista Gear":
		return 11013
	default:
		return 11011
	}
}
