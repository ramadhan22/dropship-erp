package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/ramadhan22/dropship-erp/backend/internal/config"
	"github.com/xuri/excelize/v2"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	"github.com/ramadhan22/dropship-erp/backend/internal/repository"
)

// expectedHeaders lists column names (excluding the leading "No." column)
// that must appear in the header row at index 4. These were derived from the
// provided sample file.
var expectedHeaders = []string{
	"No. Pesanan",
	"No. Pengajuan",
	"Username (Pembeli)",
	"Waktu Pesanan Dibuat",
	"Metode pembayaran pembeli",
	"Tanggal Dana Dilepaskan",
	"Harga Asli Produk",
	"Total Diskon Produk",
	"Jumlah Pengembalian Dana ke Pembeli",
	"Diskon Produk dari Shopee",
	"Diskon Voucher Ditanggung Penjual",
	"Cashback Koin yang Ditanggung Penjual",
	"Ongkir Dibayar Pembeli",
	"Diskon Ongkir Ditanggung Jasa Kirim",
	"Gratis Ongkir dari Shopee",
	"Ongkir yang Diteruskan oleh Shopee ke Jasa Kirim",
	"Ongkos Kirim Pengembalian Barang",
	"Pengembalian Biaya Kirim",
	"Biaya Komisi AMS",
	"Biaya Administrasi",
	"Biaya Layanan (termasuk PPN 11%)",
	"Premi",
	"Biaya Program",
	"Biaya Kartu Kredit",
	"Biaya Kampanye",
	"Bea Masuk, PPN & PPh",
	"Total Penghasilan",
	"Kode Voucher",
	"Kompensasi",
	"Promo Gratis Ongkir dari Penjual",
	"Jasa Kirim",
	"Nama Kurir",
	"",
	"Pengembalian Dana ke Pembeli",
	"Pro-rata Koin yang Ditukarkan untuk Pengembalian Barang",
	"Pro-rata Voucher Shopee untuk Pengembalian Barang",
	"Pro-rated Bank Payment Channel Promotion  for return refund Items",
	"Pro-rated Shopee Payment Channel Promotion  for return refund Items",
}

// ShopeeRepoInterface defines methods used by ShopeeService.
type ShopeeRepoInterface interface {
	InsertShopeeSettled(ctx context.Context, s *models.ShopeeSettled) error
	InsertShopeeAffiliateSale(ctx context.Context, s *models.ShopeeAffiliateSale) error
	ListShopeeSettled(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.ShopeeSettled, int, error)
	SumShopeeSettled(ctx context.Context, channel, store, from, to string) (*models.ShopeeSummary, error)
	ExistsShopeeSettled(ctx context.Context, noPesanan string) (bool, error)
	ExistsShopeeAffiliateSale(ctx context.Context, orderID, productCode, komisiID string) (bool, error)
	DeleteShopeeAffiliateSale(ctx context.Context, orderID, productCode, komisiID string) error
	ListShopeeAffiliateSales(ctx context.Context, noPesanan, from, to string, limit, offset int) ([]models.ShopeeAffiliateSale, int, error)
	SumShopeeAffiliateSales(ctx context.Context, noPesanan, from, to string) (*models.ShopeeAffiliateSummary, error)
	GetAffiliateExpenseByOrder(ctx context.Context, kodePesanan string) (float64, error)
	MarkMismatch(ctx context.Context, orderSN string, mismatch bool) error
	ConfirmSettle(ctx context.Context, orderSN string) error
	GetBySN(ctx context.Context, orderSN string) (*models.ShopeeSettled, error)
	ListSalesProfit(ctx context.Context, channel, store, from, to, orderNo, sortBy, dir string, limit, offset int) ([]models.SalesProfit, int, error)
}

type ShopeeDropshipRepo interface {
	GetDropshipPurchaseByInvoice(ctx context.Context, kodeInvoice string) (*models.DropshipPurchase, error)
	GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error)
	GetDropshipPurchaseByTransaction(ctx context.Context, kodeTransaksi string) (*models.DropshipPurchase, error)
	SumDetailByInvoice(ctx context.Context, kodeInvoice string) (float64, error)
	UpdateDropshipStatus(ctx context.Context, kodePesanan, status string) error
}

// ShopeeService handles import of settled Shopee orders from XLSX files.
type ShopeeJournalRepo interface {
	CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error)
	InsertJournalLine(ctx context.Context, l *models.JournalLine) error
	GetJournalEntryBySource(ctx context.Context, sourceType, sourceID string) (*models.JournalEntry, error)
	GetLinesByJournalID(ctx context.Context, id int64) ([]repository.JournalLineDetail, error)
	UpdateJournalLineAmount(ctx context.Context, lineID int64, amount float64) error
	DeleteJournalEntry(ctx context.Context, id int64) error
}

// ShopeeService handles import of settled Shopee orders from XLSX files.
type ShopeeService struct {
	db           *sqlx.DB
	repo         ShopeeRepoInterface
	dropshipRepo ShopeeDropshipRepo
	journalRepo  ShopeeJournalRepo
}

// NewShopeeService constructs a ShopeeService.
func NewShopeeService(db *sqlx.DB, r ShopeeRepoInterface, dr ShopeeDropshipRepo, jr ShopeeJournalRepo) *ShopeeService {
	return &ShopeeService{db: db, repo: r, dropshipRepo: dr, journalRepo: jr}
}

// ImportSettledOrdersXLSX reads an XLSX file and inserts rows into shopee_settled.
// It returns the count of successfully inserted rows.
func (s *ShopeeService) ImportSettledOrdersXLSX(ctx context.Context, r io.Reader) (int, []string, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return 0, nil, fmt.Errorf("open xlsx: %w", err)
	}
	sheets := f.GetSheetList()
	if len(sheets) < 2 {
		return 0, nil, fmt.Errorf("second sheet not found")
	}
	sheet := sheets[1]

	rows, err := f.GetRows(sheet)
	if err != nil {
		return 0, nil, fmt.Errorf("read rows: %w", err)
	}

	storeUsername, _ := f.GetCellValue(sheet, "A2")
	namaToko := formatNamaToko(storeUsername)

	headerIndex := -1
	for i, row := range rows {
		if len(row) > 1 && strings.TrimSpace(row[1]) == expectedHeaders[0] {
			headerIndex = i
			break
		}
	}
	if headerIndex == -1 {
		return 0, nil, fmt.Errorf("header row not found")
	}
	header := rows[headerIndex]
	if len(header) < len(expectedHeaders)+1 { // +1 for the \"No.\" column
		return 0, nil, fmt.Errorf("invalid header length")
	}
	for i, name := range expectedHeaders {
		if strings.TrimSpace(header[i+1]) != name {
			return 0, nil, fmt.Errorf("unexpected header %q at column %d", header[i+1], i+2)
		}
	}

	entries := []*models.ShopeeSettled{}
	orderNos := []string{}
	for i := headerIndex + 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 37 {
			continue
		}
		if strings.TrimSpace(row[1]) == "" || strings.Contains(strings.ToLower(row[1]), "total") || strings.Contains(strings.ToLower(row[1]), "summary") {
			continue
		}

		entry, err := parseShopeeRow(row, namaToko)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
		orderNos = append(orderNos, entry.NoPesanan)
	}

	existing, err := s.existingOrders(ctx, orderNos)
	if err != nil {
		return 0, nil, fmt.Errorf("check existing: %w", err)
	}

	inserted := 0
	mismatches := []string{}
	for _, entry := range entries {
		if existing[entry.NoPesanan] {
			continue
		}
		if err := s.repo.InsertShopeeSettled(ctx, entry); err != nil {
			return inserted, mismatches, fmt.Errorf("insert %s: %w", entry.NoPesanan, err)
		}
		var sum float64
		if s.db != nil {
			_ = s.db.GetContext(ctx, &sum,
				`SELECT COALESCE(SUM(d.total_harga_produk_channel),0)
                                FROM dropship_purchase_details d
                                JOIN dropship_purchases p ON d.kode_pesanan = p.kode_pesanan
                                WHERE p.kode_invoice_channel=$1`,
				entry.NoPesanan)
		} else if s.dropshipRepo != nil {
			sum, _ = s.dropshipRepo.SumDetailByInvoice(ctx, entry.NoPesanan)
		}
		mismatch := sum != entry.HargaAsliProduk
		_ = s.repo.MarkMismatch(ctx, entry.NoPesanan, mismatch)
		if err := s.ConfirmSettle(ctx, entry.NoPesanan); err != nil {
			if mismatch {
				mismatches = append(mismatches, entry.NoPesanan)
			} else {
				return inserted, mismatches, fmt.Errorf("auto settle %s: %w", entry.NoPesanan, err)
			}
		} else if mismatch {
			_ = s.repo.MarkMismatch(ctx, entry.NoPesanan, false)
		}
		// Update related dropship purchase status if applicable
		if s.dropshipRepo != nil && entry.NoPengajuan != "" {
			if dp, _ := s.dropshipRepo.GetDropshipPurchaseByTransaction(ctx, entry.NoPengajuan); dp != nil {
				if dp.StatusPesananTerakhir != "Pesanan selesai" && dp.StatusPesananTerakhir != "Pesanan dibatalkan" {
					_ = s.dropshipRepo.UpdateDropshipStatus(ctx, dp.KodePesanan, "Pesanan selesai")
				}
			}
		}
		inserted++
	}
	return inserted, mismatches, nil
}

// ImportAffiliateCSV reads a CSV file of affiliate sales and inserts rows.
func (s *ShopeeService) ImportAffiliateCSV(ctx context.Context, r io.Reader) (int, error) {
	reader := csv.NewReader(r)
	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("read header: %w", err)
	}
	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\ufeff")
	}
	allowedStatus := map[string]bool{
		"Belum Dibayar":   true,
		"Sedang Diproses": true,
		"Selesai":         true,
		"Dibatalkan":      true,
	}

	inserted := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return inserted, fmt.Errorf("read row: %w", err)
		}
		if len(row) < 39 {
			continue
		}
		row[0] = strings.TrimSpace(row[0])
		if row[0] == "" {
			continue
		}
		entry, err := parseAffiliateRow(row)
		if err != nil {
			continue
		}
		if !allowedStatus[entry.StatusPesanan] {
			continue
		}
		exists, err := s.repo.ExistsShopeeAffiliateSale(ctx, entry.KodePesanan, entry.KodeProduk, entry.IDKomisiPesanan)
		if err != nil {
			return inserted, fmt.Errorf("check existing: %w", err)
		}
		if exists {
			if err := s.repo.DeleteShopeeAffiliateSale(ctx, entry.KodePesanan, entry.KodeProduk, entry.IDKomisiPesanan); err != nil {
				return inserted, fmt.Errorf("delete existing: %w", err)
			}
			if s.journalRepo != nil {
				if je, err := s.journalRepo.GetJournalEntryBySource(ctx, "shopee_affiliate", fmt.Sprintf("%s-%s", entry.KodePesanan, entry.KodeProduk)); err == nil && je != nil {
					_ = s.journalRepo.DeleteJournalEntry(ctx, je.JournalID)
				}
			}
		}
		orderExists, err := s.repo.ExistsShopeeSettled(ctx, entry.KodePesanan)
		if err != nil {
			return inserted, fmt.Errorf("check order: %w", err)
		}
		if !orderExists && s.dropshipRepo != nil {
			if dp, _ := s.dropshipRepo.GetDropshipPurchaseByInvoice(ctx, entry.KodePesanan); dp != nil {
				entry.NamaToko = dp.NamaToko
			}
		} else if orderExists && entry.NamaToko == "" && s.dropshipRepo != nil {
			if dp, _ := s.dropshipRepo.GetDropshipPurchaseByInvoice(ctx, entry.KodePesanan); dp != nil {
				entry.NamaToko = dp.NamaToko
			} else if dp, _ := s.dropshipRepo.GetDropshipPurchaseByID(ctx, entry.KodePesanan); dp != nil {
				entry.NamaToko = dp.NamaToko
			}
		}

		if err := s.repo.InsertShopeeAffiliateSale(ctx, entry); err != nil {
			return inserted, fmt.Errorf("insert: %w", err)
		}
		if orderExists && strings.EqualFold(entry.StatusTerverifikasi, "Sah") {
			if err := s.addAffiliateToJournal(ctx, entry); err != nil {
				return inserted, fmt.Errorf("journal: %w", err)
			}
		}
		inserted++
	}
	return inserted, nil
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	layouts := []string{"2006-01-02", "02/01/2006", "2/1/2006"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %s", s)
}

func parseDateTime(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	layouts := []string{"2006-01-02 15:04:05", "2006/01/02 15:04", "02/01/2006 15:04:05"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid datetime %s", s)
}

func parsePercent(s string) (float64, error) {
	s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
	return parseFloat(s)
}

func parseFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func parseShopeeRow(row []string, namaToko string) (*models.ShopeeSettled, error) {
	var err error
	res := &models.ShopeeSettled{NamaToko: namaToko}
	res.NoPesanan = row[1]
	res.NoPengajuan = row[2]
	res.UsernamePembeli = row[3]
	if res.WaktuPesananDibuat, err = parseDate(row[4]); err != nil {
		return nil, err
	}
	res.MetodePembayaranPembeli = row[5]
	if res.TanggalDanaDilepaskan, err = parseDate(row[6]); err != nil {
		return nil, err
	}
	if res.HargaAsliProduk, err = parseFloat(row[7]); err != nil {
		return nil, err
	}
	if res.TotalDiskonProduk, err = parseFloat(row[8]); err != nil {
		return nil, err
	}
	if res.JumlahPengembalianDanaKePembeli, err = parseFloat(row[9]); err != nil {
		return nil, err
	}
	if res.KomisiShopee, err = parseFloat(row[10]); err != nil {
		return nil, err
	}
	if res.BiayaAdminShopee, err = parseFloat(row[11]); err != nil {
		return nil, err
	}
	if res.BiayaLayanan, err = parseFloat(row[12]); err != nil {
		return nil, err
	}
	if res.BiayaLayananEkstra, err = parseFloat(row[13]); err != nil {
		return nil, err
	}
	if res.BiayaPenyediaPembayaran, err = parseFloat(row[14]); err != nil {
		return nil, err
	}
	if res.Asuransi, err = parseFloat(row[15]); err != nil {
		return nil, err
	}
	if res.TotalBiayaTransaksi, err = parseFloat(row[16]); err != nil {
		return nil, err
	}
	if res.BiayaPengiriman, err = parseFloat(row[17]); err != nil {
		return nil, err
	}
	if res.TotalDiskonPengiriman, err = parseFloat(row[18]); err != nil {
		return nil, err
	}
	if res.PromoGratisOngkirShopee, err = parseFloat(row[19]); err != nil {
		return nil, err
	}
	if res.PromoGratisOngkirPenjual, err = parseFloat(row[20]); err != nil {
		return nil, err
	}
	if res.PromoDiskonShopee, err = parseFloat(row[21]); err != nil {
		return nil, err
	}
	if res.PromoDiskonPenjual, err = parseFloat(row[22]); err != nil {
		return nil, err
	}
	if res.CashbackShopee, err = parseFloat(row[23]); err != nil {
		return nil, err
	}
	if res.CashbackPenjual, err = parseFloat(row[24]); err != nil {
		return nil, err
	}
	if res.KoinShopee, err = parseFloat(row[25]); err != nil {
		return nil, err
	}
	if res.PotonganLainnya, err = parseFloat(row[26]); err != nil {
		return nil, err
	}
	if res.TotalPenerimaan, err = parseFloat(row[27]); err != nil {
		return nil, err
	}
	// column 28 is "Kode Voucher" and is ignored
	if res.Kompensasi, err = parseFloat(row[29]); err != nil {
		return nil, err
	}
	if res.PromoGratisOngkirDariPenjual, err = parseFloat(row[30]); err != nil {
		return nil, err
	}
	res.JasaKirim = row[31]
	res.NamaKurir = row[32]
	// column 33 is blank
	if res.PengembalianDanaKePembeli, err = parseFloat(row[34]); err != nil {
		return nil, err
	}
	if res.ProRataKoinYangDitukarkanUntukPengembalianBarang, err = parseFloat(row[35]); err != nil {
		return nil, err
	}
	if res.ProRataVoucherShopeeUntukPengembalianBarang, err = parseFloat(row[36]); err != nil {
		return nil, err
	}
	if res.ProRatedBankPaymentChannelPromotionForReturns, err = parseFloat(row[37]); err != nil {
		return nil, err
	}
	if res.ProRatedShopeePaymentChannelPromotionForReturns, err = parseFloat(row[38]); err != nil {
		return nil, err
	}
	return res, nil
}

func parseAffiliateRow(row []string) (*models.ShopeeAffiliateSale, error) {
	var err error
	res := &models.ShopeeAffiliateSale{}
	res.KodePesanan = row[0]
	res.StatusPesanan = row[1]
	res.StatusTerverifikasi = row[2]
	if res.WaktuPesanan, err = parseDateTime(row[3]); err != nil {
		return nil, err
	}
	if res.WaktuPesananSelesai, err = parseDateTime(row[4]); err != nil {
		return nil, err
	}
	if res.WaktuPesananTerverifikasi, err = parseDateTime(row[5]); err != nil {
		return nil, err
	}
	res.KodeProduk = row[6]
	res.NamaProduk = row[7]
	res.IDModel = row[8]
	res.L1KategoriGlobal = row[9]
	res.L2KategoriGlobal = row[10]
	res.L3KategoriGlobal = row[11]
	res.KodePromo = row[12]
	if res.Harga, err = parseFloat(row[13]); err != nil {
		return nil, err
	}
	if res.Jumlah, err = strconv.Atoi(row[14]); err != nil {
		return nil, err
	}
	res.NamaAffiliate = row[15]
	res.UsernameAffiliate = row[16]
	res.MCNTerhubung = row[17]
	res.IDKomisiPesanan = row[18]
	res.PartnerPromo = row[19]
	res.JenisPromo = row[20]
	if res.NilaiPembelian, err = parseFloat(row[21]); err != nil {
		return nil, err
	}
	if res.JumlahPengembalian, err = parseFloat(row[22]); err != nil {
		return nil, err
	}
	res.TipePesanan = row[23]
	if res.EstimasiKomisiPerProduk, err = parseFloat(row[24]); err != nil {
		return nil, err
	}
	if res.EstimasiKomisiAffiliatePerProduk, err = parseFloat(row[25]); err != nil {
		return nil, err
	}
	if res.PersentaseKomisiAffiliatePerProduk, err = parsePercent(row[26]); err != nil {
		return nil, err
	}
	if res.EstimasiKomisiMCNPerProduk, err = parseFloat(row[27]); err != nil {
		return nil, err
	}
	if res.PersentaseKomisiMCNPerProduk, err = parsePercent(row[28]); err != nil {
		return nil, err
	}
	if res.EstimasiKomisiPerPesanan, err = parseFloat(row[29]); err != nil {
		return nil, err
	}
	if res.EstimasiKomisiAffiliatePerPesanan, err = parseFloat(row[30]); err != nil {
		return nil, err
	}
	if res.EstimasiKomisiMCNPerPesanan, err = parseFloat(row[31]); err != nil {
		return nil, err
	}
	res.CatatanProduk = row[32]
	res.Platform = row[33]
	if res.TingkatKomisi, err = parsePercent(row[34]); err != nil {
		return nil, err
	}
	if res.Pengeluaran, err = parseFloat(row[35]); err != nil {
		return nil, err
	}
	res.StatusPemotongan = row[36]
	res.MetodePemotongan = row[37]
	if res.WaktuPemotongan, err = parseDateTime(row[38]); err != nil {
		return nil, err
	}
	return res, nil
}

func formatNamaToko(username string) string {
	u := strings.ToLower(strings.TrimSpace(username))
	if u == "" {
		return ""
	}
	if u == "mrest0re" {
		return "MR eStore Shopee"
	}
	u = strings.ReplaceAll(u, ".", " ")
	return CapitalizeWords(u)
}

func (s *ShopeeService) existingOrders(ctx context.Context, ids []string) (map[string]bool, error) {
	m := make(map[string]bool, len(ids))
	if len(ids) == 0 {
		return m, nil
	}
	if s.db != nil {
		var existing []string
		query := `SELECT no_pesanan FROM shopee_settled WHERE no_pesanan = ANY($1)`
		if err := s.db.SelectContext(ctx, &existing, query, pq.Array(ids)); err != nil {
			return nil, err
		}
		for _, id := range existing {
			m[id] = true
		}
		return m, nil
	}
	for _, id := range ids {
		ok, err := s.repo.ExistsShopeeSettled(ctx, id)
		if err != nil {
			return nil, err
		}
		if ok {
			m[id] = true
		}
	}
	return m, nil
}

// ListSettled proxies to the repository for fetching settled orders with filters.
func (s *ShopeeService) ListSettled(
	ctx context.Context,
	channel, store, from, to, orderNo, sortBy, dir string,
	limit, offset int,
) ([]models.ShopeeSettled, int, error) {
	return s.repo.ListShopeeSettled(ctx, channel, store, from, to, orderNo, sortBy, dir, limit, offset)
}

func (s *ShopeeService) SumShopeeSettled(
	ctx context.Context,
	channel, store, from, to string,
) (*models.ShopeeSummary, error) {
	return s.repo.SumShopeeSettled(ctx, channel, store, from, to)
}

// ListAffiliate proxies to the repository for fetching affiliate sales with filters.
func (s *ShopeeService) ListAffiliate(
	ctx context.Context,
	noPesanan, from, to string,
	limit, offset int,
) ([]models.ShopeeAffiliateSale, int, error) {
	return s.repo.ListShopeeAffiliateSales(ctx, noPesanan, from, to, limit, offset)
}

func (s *ShopeeService) SumAffiliate(
	ctx context.Context,
	noPesanan, from, to string,
) (*models.ShopeeAffiliateSummary, error) {
	return s.repo.SumShopeeAffiliateSales(ctx, noPesanan, from, to)
}

// ListSalesProfit fetches sales with cost and profit calculations.
func (s *ShopeeService) ListSalesProfit(
	ctx context.Context,
	channel, store, from, to, orderNo, sortBy, dir string,
	limit, offset int,
) ([]models.SalesProfit, int, error) {
	list, total, err := s.repo.ListSalesProfit(ctx, channel, store, from, to, orderNo, sortBy, dir, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	for i := range list {
		p := list[i]
		profit := p.AmountSales - (p.ModalPurchase + p.BiayaMitraJakmall + p.BiayaAdministrasi + p.BiayaLayanan + p.BiayaVoucher + p.BiayaAffiliate + p.Discount)
		list[i].Profit = profit
		if p.AmountSales == 0 {
			list[i].ProfitPercent = 0
		} else {
			list[i].ProfitPercent = profit / p.AmountSales * 100
		}
	}
	return list, total, nil
}

func (s *ShopeeService) GetSettleDetail(ctx context.Context, orderSN string) (*models.ShopeeSettled, float64, error) {
	o, err := s.repo.GetBySN(ctx, orderSN)
	if err != nil {
		return nil, 0, err
	}
	var sum float64
	if s.dropshipRepo != nil {
		sum, _ = s.dropshipRepo.SumDetailByInvoice(ctx, orderSN)
	}
	return o, sum, nil
}

func (s *ShopeeService) createSettlementJournal(ctx context.Context, jr ShopeeJournalRepo, repo ShopeeRepoInterface, entry *models.ShopeeSettled) error {
	if jr == nil {
		return nil
	}
	affiliate, _ := repo.GetAffiliateExpenseByOrder(ctx, entry.NoPesanan)
	netSale := entry.HargaAsliProduk + entry.TotalDiskonProduk
	disc := abs(entry.TotalDiskonProduk)
	if je, err := jr.GetJournalEntryBySource(ctx, "shopee_discount", entry.NoPesanan+"-discount"); err == nil && je != nil {
		disc = 0
	}
	voucher := abs(entry.BiayaAdminShopee)
	admin := abs(entry.PromoGratisOngkirPenjual)
	layanan := abs(entry.PromoDiskonShopee)
	affiliateAmt := abs(affiliate)
	saldo := netSale - disc - voucher - admin - layanan - affiliateAmt

	je := &models.JournalEntry{
		EntryDate:    entry.TanggalDanaDilepaskan,
		Description:  ptrString("Shopee settled " + entry.NoPesanan),
		SourceType:   "shopee_settled",
		SourceID:     entry.NoPesanan,
		ShopUsername: entry.NamaToko,
		Store:        entry.NamaToko,
		CreatedAt:    time.Now(),
	}
	id, err := jr.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	lines := []models.JournalLine{
		{JournalID: id, AccountID: pendingAccountID(entry.NamaToko), IsDebit: false, Amount: netSale, Memo: ptrString("Pending " + entry.NoPesanan)},
		{JournalID: id, AccountID: 52002, IsDebit: true, Amount: disc, Memo: ptrString("Discount " + entry.NoPesanan)},
		{JournalID: id, AccountID: 52003, IsDebit: true, Amount: voucher, Memo: ptrString("Voucher " + entry.NoPesanan)},
		{JournalID: id, AccountID: 52006, IsDebit: true, Amount: admin, Memo: ptrString("Biaya Administrasi " + entry.NoPesanan)},
		{JournalID: id, AccountID: 52004, IsDebit: true, Amount: layanan, Memo: ptrString("Biaya Layanan " + entry.NoPesanan)},
		{JournalID: id, AccountID: 52005, IsDebit: true, Amount: affiliateAmt, Memo: ptrString("Biaya Affiliate " + entry.NoPesanan)},
		{JournalID: id, AccountID: saldoShopeeAccountID(entry.NamaToko), IsDebit: true, Amount: saldo, Memo: ptrString("Saldo Shopee " + entry.NoPesanan)},
	}
	for i := range lines {
		if lines[i].Amount == 0 {
			continue
		}
		if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}
	return nil
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// addAffiliateToJournal creates a new journal entry for the given affiliate
// sale. The entry debits Biaya Affiliate and credits the Saldo Shopee account
// for the related store.
func (s *ShopeeService) addAffiliateToJournal(ctx context.Context, sale *models.ShopeeAffiliateSale) error {
	if s.journalRepo == nil || sale == nil || sale.Pengeluaran == 0 {
		return nil
	}

	entryDate := sale.WaktuPemotongan
	if entryDate.IsZero() {
		entryDate = time.Now()
	}
	je := &models.JournalEntry{
		EntryDate:    entryDate,
		Description:  ptrString("Shopee affiliate " + sale.KodePesanan),
		SourceType:   "shopee_affiliate",
		SourceID:     fmt.Sprintf("%s-%s", sale.KodePesanan, sale.KodeProduk),
		ShopUsername: sale.NamaToko,
		Store:        sale.NamaToko,
		CreatedAt:    time.Now(),
	}
	jid, err := s.journalRepo.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	lines := []models.JournalLine{
		{JournalID: jid, AccountID: 52005, IsDebit: true, Amount: sale.Pengeluaran, Memo: ptrString("Biaya Affiliate " + sale.KodePesanan)},
		{JournalID: jid, AccountID: saldoShopeeAccountID(sale.NamaToko), IsDebit: false, Amount: sale.Pengeluaran, Memo: ptrString("Saldo Shopee " + sale.KodePesanan)},
	}
	for i := range lines {
		if lines[i].Amount == 0 {
			continue
		}
		if err := s.journalRepo.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *ShopeeService) handleMismatch(ctx context.Context, jr ShopeeJournalRepo, o *models.ShopeeSettled) error {
	if s.dropshipRepo == nil {
		return fmt.Errorf("dropship repo nil")
	}
	sum, err := s.dropshipRepo.SumDetailByInvoice(ctx, o.NoPesanan)
	if err != nil {
		return err
	}
	if math.Abs((o.HargaAsliProduk+o.TotalDiskonProduk)-sum) > 0.01 {
		return fmt.Errorf("amount mismatch")
	}
	diff := o.HargaAsliProduk - sum
	disc := abs(o.TotalDiskonProduk)
	if diff == 0 && disc == 0 {
		return nil
	}
	if diff != 0 {
		if err := s.createGrossUpJournal(ctx, jr, o, diff); err != nil {
			return err
		}
	}
	if disc != 0 {
		if err := s.createDiscountJournal(ctx, jr, o, disc); err != nil {
			return err
		}
	}
	return nil
}

func (s *ShopeeService) createGrossUpJournal(ctx context.Context, jr ShopeeJournalRepo, o *models.ShopeeSettled, diff float64) error {
	je := &models.JournalEntry{
		EntryDate:    o.TanggalDanaDilepaskan,
		Description:  ptrString("Gross Up " + o.NoPesanan),
		SourceType:   "shopee_grossup",
		SourceID:     o.NoPesanan + "-grossup",
		ShopUsername: o.NamaToko,
		Store:        o.NamaToko,
		CreatedAt:    time.Now(),
	}
	id, err := jr.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	lines := []models.JournalLine{
		{JournalID: id, AccountID: pendingAccountID(o.NamaToko), IsDebit: true, Amount: diff},
		{JournalID: id, AccountID: 4001, IsDebit: false, Amount: diff},
	}
	for i := range lines {
		if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *ShopeeService) createDiscountJournal(ctx context.Context, jr ShopeeJournalRepo, o *models.ShopeeSettled, disc float64) error {
	je := &models.JournalEntry{
		EntryDate:    o.TanggalDanaDilepaskan,
		Description:  ptrString("Discount " + o.NoPesanan),
		SourceType:   "shopee_discount",
		SourceID:     o.NoPesanan + "-discount",
		ShopUsername: o.NamaToko,
		Store:        o.NamaToko,
		CreatedAt:    time.Now(),
	}
	id, err := jr.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}
	lines := []models.JournalLine{
		{JournalID: id, AccountID: pendingAccountID(o.NamaToko), IsDebit: false, Amount: disc},
		{JournalID: id, AccountID: 52002, IsDebit: true, Amount: disc},
	}
	for i := range lines {
		if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
			return err
		}
	}
	return nil
}

// ConfirmSettle posts journal entries for the given order if data is valid.
func (s *ShopeeService) ConfirmSettle(ctx context.Context, orderSN string) error {
	o, err := s.repo.GetBySN(ctx, orderSN)
	if err != nil {
		return err
	}
	if o.IsSettledConfirmed {
		return fmt.Errorf("cannot settle")
	}

	jr := s.journalRepo
	if s.db != nil {
		tx, err := s.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		jr = repository.NewJournalRepo(tx)

		if o.IsDataMismatch {
			if err := s.handleMismatch(ctx, jr, o); err != nil {
				return err
			}
		}

		if err := s.createSettlementJournal(ctx, jr, s.repo, o); err != nil {
			return err
		}
		if err := s.repo.ConfirmSettle(ctx, orderSN); err != nil {
			return err
		}
		return tx.Commit()
	}

	if jr != nil {
		if o.IsDataMismatch {
			if err := s.handleMismatch(ctx, jr, o); err != nil {
				return err
			}
		}
		if err := s.createSettlementJournal(ctx, jr, s.repo, o); err != nil {
			return err
		}
	}
	return s.repo.ConfirmSettle(ctx, orderSN)
}

func CapitalizeWords(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		lowerWord := strings.ToLower(word)
		if lowerWord == "mr" {
			words[i] = "MR"
		} else {
			runes := []rune(lowerWord)
			if len(runes) > 0 {
				runes[0] = unicode.ToUpper(runes[0])
			}
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

// withdrawResp models the payout API response.
type withdrawResp struct {
	Fee float64 `json:"fee"`
}

// WithdrawShopeeBalance moves funds from Shopee balance to bank account.
func (s *ShopeeService) WithdrawShopeeBalance(ctx context.Context, store string, amount float64) error {
	cfg := config.MustLoadConfig()

	form := url.Values{}
	form.Set("shopid", cfg.Shopee.ShopID)
	form.Set("amount", fmt.Sprintf("%.2f", amount))
	form.Set("access_token", cfg.Shopee.AccessToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.Shopee.BaseURL+"/api/v2/shop/withdraw", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("withdraw status %d", resp.StatusCode)
	}

	var out withdrawResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}

	jr := s.journalRepo
	if s.db != nil {
		tx, err := s.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		jr = repository.NewJournalRepo(tx)

		if err := createWithdrawJournal(ctx, jr, store, amount, out.Fee); err != nil {
			return err
		}
		return tx.Commit()
	}
	return createWithdrawJournal(ctx, jr, store, amount, out.Fee)
}

func createWithdrawJournal(ctx context.Context, jr ShopeeJournalRepo, store string, amount, fee float64) error {
	je := &models.JournalEntry{
		EntryDate:    time.Now(),
		Description:  ptrString("Withdraw Shopee Balance"),
		SourceType:   "withdraw",
		SourceID:     fmt.Sprintf("%s-%d", store, time.Now().UnixNano()),
		ShopUsername: store,
		Store:        store,
		CreatedAt:    time.Now(),
	}
	jid, err := jr.CreateJournalEntry(ctx, je)
	if err != nil {
		return err
	}

	bankAcc := int64(1002) // account_code 1.1.2
	saldoAcc := saldoShopeeAccountID(store)
	if fee > 0 {
		net := amount - fee
		lines := []models.JournalLine{
			{JournalID: jid, AccountID: bankAcc, IsDebit: true, Amount: net},
			{JournalID: jid, AccountID: 52008, IsDebit: true, Amount: fee},
			{JournalID: jid, AccountID: saldoAcc, IsDebit: false, Amount: amount},
		}
		for i := range lines {
			if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
				return err
			}
		}
	} else {
		lines := []models.JournalLine{
			{JournalID: jid, AccountID: bankAcc, IsDebit: true, Amount: amount},
			{JournalID: jid, AccountID: saldoAcc, IsDebit: false, Amount: amount},
		}
		for i := range lines {
			if err := jr.InsertJournalLine(ctx, &lines[i]); err != nil {
				return err
			}
		}
	}
	return nil
}
