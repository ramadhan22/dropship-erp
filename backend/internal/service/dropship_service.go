// File: backend/internal/service/dropship_service.go

package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
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
	batchSvc    *BatchService
	client      *ShopeeClient
	cache       Cache // Add cache interface
	maxThreads  int
	batchSize   int
}

// Cache interface for dropship service
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// NewDropshipService constructs a DropshipService with the given repository.
func NewDropshipService(
	db *sqlx.DB,
	repo DropshipRepoInterface,
	jr DropshipJournalRepo,
	sr DropshipServiceStoreRepo,
	dr DropshipServiceDetailRepo,
	bs *BatchService,
	c *ShopeeClient,
	cache Cache,
	maxThreads int,
	batchSize int,
) *DropshipService {
	return &DropshipService{
		db:          db,
		repo:        repo,
		journalRepo: jr,
		storeRepo:   sr,
		detailRepo:  dr,
		batchSvc:    bs,
		client:      c,
		cache:       cache,
		maxThreads:  maxThreads,
		batchSize:   batchSize,
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
func (s *DropshipService) ImportFromCSV(ctx context.Context, r io.Reader, channel string, batchID int64) (int, error) {
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
			KodeInvoiceChannel: strings.TrimPrefix(record[19], "'"),
			JenisChannel:       record[17],
		}
		if channel != "" && h.JenisChannel != channel {
			continue
		}
		if existing[h.KodePesanan] || fetched[h.KodePesanan] {
			continue
		}
		fetched[h.KodePesanan] = true
		if !strings.EqualFold(h.NamaToko, "MR eStore Free Sample") {
			batches[h.NamaToko] = append(batches[h.NamaToko], h)
		}
	}

	apiTotals := make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup
	limit := s.maxThreads
	if limit <= 0 {
		limit = 5
	}
	sem := make(chan struct{}, limit)

	for store, list := range batches {
		if strings.EqualFold(store, "MR eStore Free Sample") {
			continue
		}
		for i := 0; i < len(list); i += 50 {
			end := i + 50
			if end > len(list) {
				end = len(list)
			}
			subset := list[i:end]
			wg.Add(1)
			sem <- struct{}{}
			go func(st string, batch []*models.DropshipPurchase) {
				defer func() {
					<-sem
					wg.Done()
				}()
				amtMap, err := s.fetchAndStoreDetailBatch(ctx, batch)
				mu.Lock()
				if err != nil {
					log.Printf("fetch batch detail store %s: %v", st, err)
					for _, h := range batch {
						skipped[h.KodePesanan] = true
						if s.batchSvc != nil && batchID != 0 {
							d := &models.BatchHistoryDetail{
								BatchID:   batchID,
								Reference: h.KodeInvoiceChannel,
								Store:     h.NamaToko,
								Status:    "failed",
								ErrorMsg:  err.Error(),
							}
							_ = s.batchSvc.CreateDetail(ctx, d)
						}
					}
					mu.Unlock()
					return
				}
				for k, v := range amtMap {
					apiTotals[k] = v
				}
				mu.Unlock()
			}(store, subset)
		}
	}
	wg.Wait()
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
		if record[18] == "MR eStore Free Sample" {
			log.Printf("processing MR eStore Free Sample")
		}

		qty, err := strconv.Atoi(record[8])
		if err != nil {
			logutil.Errorf("ImportFromCSV parse qty error: %v", err)
			if s.batchSvc != nil && batchID != 0 {
				d := &models.BatchHistoryDetail{
					BatchID:   batchID,
					Reference: record[19],
					Store:     record[18],
					Status:    "failed",
					ErrorMsg:  fmt.Sprintf("parse qty: %v", err),
				}
				_ = s.batchSvc.CreateDetail(ctx, d)
			}
			continue
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
			if s.batchSvc != nil && batchID != 0 {
				d := &models.BatchHistoryDetail{
					BatchID:   batchID,
					Reference: record[19],
					Store:     record[18],
					Status:    "failed",
					ErrorMsg:  fmt.Sprintf("parse waktu_pesanan: %v", err),
				}
				_ = s.batchSvc.CreateDetail(ctx, d)
			}
			continue
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
			KodeInvoiceChannel:    strings.TrimPrefix(record[19], "'"),
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
				if s.batchSvc != nil && batchID != 0 {
					d := &models.BatchHistoryDetail{
						BatchID:   batchID,
						Reference: header.KodeInvoiceChannel,
						Store:     header.NamaToko,
						Status:    "failed",
						ErrorMsg:  "data already exist",
					}
					_ = s.batchSvc.CreateDetail(ctx, d)
				}
				continue
			}

			apiAmt := apiTotals[header.KodePesanan]

			if err := repoTx.InsertDropshipPurchase(ctx, header); err != nil {
				log.Printf("ImportFromCSV insert purchase %s error: %v", header.KodePesanan, err)
				if s.batchSvc != nil && batchID != 0 {
					d := &models.BatchHistoryDetail{
						BatchID:   batchID,
						Reference: header.KodeInvoiceChannel,
						Store:     header.NamaToko,
						Status:    "failed",
						ErrorMsg:  err.Error(),
					}
					_ = s.batchSvc.CreateDetail(ctx, d)
				}
				skipped[header.KodePesanan] = true
				continue
			}
			inserted[header.KodePesanan] = true
			headersMap[header.KodePesanan] = header
			if s.batchSvc != nil && batchID != 0 {
				d := &models.BatchHistoryDetail{
					BatchID:   batchID,
					Reference: header.KodeInvoiceChannel,
					Store:     header.NamaToko,
					Status:    "success",
				}
				_ = s.batchSvc.CreateDetail(ctx, d)
			}
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
			if s.batchSvc != nil && batchID != 0 {
				d := &models.BatchHistoryDetail{
					BatchID:   batchID,
					Reference: header.KodeInvoiceChannel,
					Store:     header.NamaToko,
					Status:    "failed",
					ErrorMsg:  err.Error(),
				}
				_ = s.batchSvc.CreateDetail(ctx, d)
			}
			skipped[header.KodePesanan] = true
			continue
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
		if s.batchSvc != nil && batchID != 0 {
			_ = s.batchSvc.UpdateDone(ctx, batchID, count)
		}
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
		if strings.EqualFold(h.NamaToko, "MR eStore Free Sample") {
			log.Printf("creating free sample journal for %s", kode)
			if err := s.createFreeSampleJournal(ctx, jrTx, h, prod); err != nil {
				log.Printf("journal %s: %v", kode, err)
				if s.batchSvc != nil && batchID != 0 {
					d := &models.BatchHistoryDetail{
						BatchID:   batchID,
						Reference: h.KodeInvoiceChannel,
						Store:     h.NamaToko,
						Status:    "failed",
						ErrorMsg:  err.Error(),
					}
					_ = s.batchSvc.CreateDetail(ctx, d)
				}
				continue
			}
			if repoUp, ok := repoTx.(interface {
				UpdatePurchaseStatus(context.Context, string, string) error
			}); ok {
				if err := repoUp.UpdatePurchaseStatus(ctx, h.KodePesanan, "Pesanan selesai"); err != nil {
					log.Printf("update status %s: %v", h.KodePesanan, err)
				}
			}
			continue
		}
		if err := s.createPendingSalesJournal(ctx, jrTx, h, prod, pending); err != nil {
			log.Printf("journal %s: %v", kode, err)
			if s.batchSvc != nil && batchID != 0 {
				d := &models.BatchHistoryDetail{
					BatchID:   batchID,
					Reference: h.KodeInvoiceChannel,
					Store:     h.NamaToko,
					Status:    "failed",
					ErrorMsg:  err.Error(),
				}
				_ = s.batchSvc.CreateDetail(ctx, d)
			}
			continue
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

func freeSampleAccountID() int64 { return 55007 }

func (s *DropshipService) createFreeSampleJournal(ctx context.Context, jr DropshipJournalRepo, p *models.DropshipPurchase, totalProduk float64) error {
	if jr == nil {
		return nil
	}
	je := &models.JournalEntry{
		EntryDate:    p.WaktuPesananTerbuat,
		Description:  ptrString("Free sample " + p.KodeInvoiceChannel),
		SourceType:   "free_sample",
		SourceID:     p.KodeInvoiceChannel,
		ShopUsername: p.NamaToko,
		Store:        p.NamaToko,
		CreatedAt:    time.Now(),
	}
	id, err := jr.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	amt := totalProduk + p.BiayaLainnya + p.BiayaMitraJakmall
	lines := []models.JournalLine{
		{JournalID: id, AccountID: 11009, IsDebit: false, Amount: amt, Memo: ptrString("Saldo Jakmall " + p.KodeInvoiceChannel)},
		{JournalID: id, AccountID: freeSampleAccountID(), IsDebit: true, Amount: amt, Memo: ptrString("Free Sample " + p.KodeInvoiceChannel)},
	}
	for i := range lines {
		if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}
	return nil
}

// ========== Performance Optimization Methods ==========

// BatchInsertPurchases processes purchases in batches to improve performance
func (s *DropshipService) BatchInsertPurchases(ctx context.Context, purchases []*models.DropshipPurchase) error {
	if len(purchases) == 0 {
		return nil
	}

	log.Printf("BatchInsertPurchases: processing %d purchases in batches of %d", len(purchases), s.batchSize)

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Process in batches to avoid memory issues
	for i := 0; i < len(purchases); i += s.batchSize {
		end := i + s.batchSize
		if end > len(purchases) {
			end = len(purchases)
		}

		batch := purchases[i:end]
		log.Printf("Processing batch %d/%d (items %d-%d)", i/s.batchSize+1, (len(purchases)-1)/s.batchSize+1, i, end-1)

		for _, purchase := range batch {
			if err := s.repo.InsertDropshipPurchase(ctx, purchase); err != nil {
				return fmt.Errorf("failed to insert purchase %s: %w", purchase.KodePesanan, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully processed %d purchases", len(purchases))
	return nil
}

// GetCachedPurchaseData retrieves purchase data from cache or database with cache-aside pattern
func (s *DropshipService) GetCachedPurchaseData(ctx context.Context, channel, store, from, to string, limit, offset int) ([]models.DropshipPurchase, int, error) {
	if s.cache == nil {
		// Fallback to direct database query if cache is not available
		return s.repo.ListDropshipPurchases(ctx, channel, store, from, to, "", "kode_pesanan", "asc", limit, offset)
	}

	// Generate cache key
	cacheKey := fmt.Sprintf("purchases:%s:%s:%s:%s:%d:%d", channel, store, from, to, limit, offset)

	// Try to get from cache first
	if data, err := s.cache.Get(ctx, cacheKey); err == nil {
		var cached struct {
			Purchases []models.DropshipPurchase `json:"purchases"`
			Total     int                       `json:"total"`
		}
		if err := json.Unmarshal(data, &cached); err == nil {
			log.Printf("Cache hit for key: %s", cacheKey)
			return cached.Purchases, cached.Total, nil
		}
	}

	// Cache miss - fetch from database
	log.Printf("Cache miss for key: %s", cacheKey)
	purchases, total, err := s.repo.ListDropshipPurchases(ctx, channel, store, from, to, "", "kode_pesanan", "asc", limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Store in cache for next time
	cached := struct {
		Purchases []models.DropshipPurchase `json:"purchases"`
		Total     int                       `json:"total"`
	}{
		Purchases: purchases,
		Total:     total,
	}

	if data, err := json.Marshal(cached); err == nil {
		// Cache for 5 minutes by default
		if err := s.cache.Set(ctx, cacheKey, data, 5*time.Minute); err != nil {
			log.Printf("Failed to cache data for key %s: %v", cacheKey, err)
		}
	}

	return purchases, total, nil
}

// InvalidatePurchaseCache removes cached purchase data when data changes
func (s *DropshipService) InvalidatePurchaseCache(ctx context.Context, patterns ...string) error {
	if s.cache == nil {
		return nil
	}

	// For now, we'll invalidate based on patterns
	// In a production system, you might want to use Redis SCAN or maintain cache key sets
	for _, pattern := range patterns {
		if err := s.cache.Delete(ctx, pattern); err != nil {
			log.Printf("Failed to invalidate cache pattern %s: %v", pattern, err)
		}
	}
	return nil
}

// GetPurchaseSummaryCache retrieves cached purchase summary data
func (s *DropshipService) GetPurchaseSummaryCache(ctx context.Context, channel, store, from, to string) (float64, error) {
	if s.cache == nil {
		return s.repo.SumDropshipPurchases(ctx, channel, store, from, to)
	}

	cacheKey := fmt.Sprintf("purchase_summary:%s:%s:%s:%s", channel, store, from, to)

	// Try cache first
	if data, err := s.cache.Get(ctx, cacheKey); err == nil {
		if total, err := strconv.ParseFloat(string(data), 64); err == nil {
			return total, nil
		}
	}

	// Cache miss - fetch from database
	total, err := s.repo.SumDropshipPurchases(ctx, channel, store, from, to)
	if err != nil {
		return 0, err
	}

	// Cache the result
	if err := s.cache.Set(ctx, cacheKey, []byte(fmt.Sprintf("%.2f", total)), 5*time.Minute); err != nil {
		log.Printf("Failed to cache purchase summary: %v", err)
	}

	return total, nil
}
