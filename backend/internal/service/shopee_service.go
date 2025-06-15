package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/xuri/excelize/v2"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
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
	ListShopeeSettled(ctx context.Context, channel, store, from, to string, limit, offset int) ([]models.ShopeeSettled, int, error)
	SumShopeeSettled(ctx context.Context, channel, store, from, to string) (*models.ShopeeSummary, error)
	ExistsShopeeSettled(ctx context.Context, noPesanan string) (bool, error)
	ExistsShopeeAffiliateSale(ctx context.Context, orderID, productCode string) (bool, error)
	ListShopeeAffiliateSales(ctx context.Context, date, month, year string, limit, offset int) ([]models.ShopeeAffiliateSale, int, error)
	SumShopeeAffiliateSales(ctx context.Context, date, month, year string) (*models.ShopeeAffiliateSummary, error)
	GetAffiliateExpenseByOrder(ctx context.Context, kodePesanan string) (float64, error)
}

type ShopeeDropshipRepo interface {
	GetDropshipPurchaseByInvoice(ctx context.Context, kodeInvoice string) (*models.DropshipPurchase, error)
	GetDropshipPurchaseByID(ctx context.Context, kodePesanan string) (*models.DropshipPurchase, error)
}

// ShopeeService handles import of settled Shopee orders from XLSX files.
type ShopeeJournalRepo interface {
	CreateJournalEntry(ctx context.Context, e *models.JournalEntry) (int64, error)
	InsertJournalLine(ctx context.Context, l *models.JournalLine) error
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
func (s *ShopeeService) ImportSettledOrdersXLSX(ctx context.Context, r io.Reader) (int, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return 0, fmt.Errorf("open xlsx: %w", err)
	}
	sheets := f.GetSheetList()
	if len(sheets) < 2 {
		return 0, fmt.Errorf("second sheet not found")
	}
	sheet := sheets[1]

	rows, err := f.GetRows(sheet)
	if err != nil {
		return 0, fmt.Errorf("read rows: %w", err)
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
		return 0, fmt.Errorf("header row not found")
	}
	header := rows[headerIndex]
	if len(header) < len(expectedHeaders)+1 { // +1 for the \"No.\" column
		return 0, fmt.Errorf("invalid header length")
	}
	for i, name := range expectedHeaders {
		if strings.TrimSpace(header[i+1]) != name {
			return 0, fmt.Errorf("unexpected header %q at column %d", header[i+1], i+2)
		}
	}

	inserted := 0
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
		exists, err := s.repo.ExistsShopeeSettled(ctx, entry.NoPesanan)
		if err != nil {
			return inserted, fmt.Errorf("check existing row %d: %w", i+1, err)
		}
		if exists {
			continue
		}
		if err := s.repo.InsertShopeeSettled(ctx, entry); err != nil {
			return inserted, fmt.Errorf("insert row %d: %w", i+1, err)
		}
		if err := s.createSettlementJournal(ctx, s.journalRepo, s.repo, entry); err != nil {
			return inserted, fmt.Errorf("journal row %d: %w", i+1, err)
		}
		inserted++
	}
	return inserted, nil
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
		if s.dropshipRepo != nil {
			if dp, _ := s.dropshipRepo.GetDropshipPurchaseByInvoice(ctx, entry.KodePesanan); dp != nil {
				entry.NamaToko = dp.NamaToko
			} else if dp, _ := s.dropshipRepo.GetDropshipPurchaseByID(ctx, entry.KodePesanan); dp != nil {
				entry.NamaToko = dp.NamaToko
			}
		}
		exists, err := s.repo.ExistsShopeeAffiliateSale(ctx, entry.KodePesanan, entry.KodeProduk)
		if err != nil {
			return inserted, fmt.Errorf("check existing: %w", err)
		}
		if exists {
			continue
		}
		if err := s.repo.InsertShopeeAffiliateSale(ctx, entry); err != nil {
			return inserted, fmt.Errorf("insert: %w", err)
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
	return strings.ToUpper(u[:1]) + u[1:]
}

// ListSettled proxies to the repository for fetching settled orders with filters.
func (s *ShopeeService) ListSettled(
	ctx context.Context,
	channel, store, from, to string,
	limit, offset int,
) ([]models.ShopeeSettled, int, error) {
	return s.repo.ListShopeeSettled(ctx, channel, store, from, to, limit, offset)
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
	date, month, year string,
	limit, offset int,
) ([]models.ShopeeAffiliateSale, int, error) {
	return s.repo.ListShopeeAffiliateSales(ctx, date, month, year, limit, offset)
}

func (s *ShopeeService) SumAffiliate(
	ctx context.Context,
	date, month, year string,
) (*models.ShopeeAffiliateSummary, error) {
	return s.repo.SumShopeeAffiliateSales(ctx, date, month, year)
}

func (s *ShopeeService) createSettlementJournal(ctx context.Context, jr ShopeeJournalRepo, repo ShopeeRepoInterface, entry *models.ShopeeSettled) error {
	if jr == nil {
		return nil
	}
	affiliate, _ := repo.GetAffiliateExpenseByOrder(ctx, entry.NoPesanan)
	netSale := entry.HargaAsliProduk - entry.TotalDiskonProduk
	voucher := abs(entry.BiayaAdminShopee)
	admin := abs(entry.PromoGratisOngkirPenjual)
	layanan := abs(entry.PromoDiskonShopee)
	affiliateAmt := abs(affiliate)
	saldo := netSale - voucher - admin - layanan - affiliateAmt

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
