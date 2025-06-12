package service

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// expectedHeaders lists column names (excluding the leading "No." column)
// that must appear in the header row at index 5. These were derived from the
// provided sample file located under sample_data.
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
	"Pengembalian Dana ke Pembeli",
	"Pro-rata Koin yang Ditukarkan untuk Pengembalian Barang",
	"Pro-rata Voucher Shopee untuk Pengembalian Barang",
	"Pro-rated Bank Payment Channel Promotion  for return refund Items",
	"Pro-rated Shopee Payment Channel Promotion  for return refund Items",
}

// ShopeeRepoInterface defines methods used by ShopeeService.
type ShopeeRepoInterface interface {
	InsertShopeeSettled(ctx context.Context, s *models.ShopeeSettled) error
}

// ShopeeService handles import of settled Shopee orders from XLSX files.
type ShopeeService struct {
	repo ShopeeRepoInterface
}

// NewShopeeService constructs a ShopeeService.
func NewShopeeService(r ShopeeRepoInterface) *ShopeeService {
	return &ShopeeService{repo: r}
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

	// Validate header row at index 5 matches expected columns (after the
	// first \"No.\" column). If headers don't match we fail early so the
	// caller is aware the template changed.
	if len(rows) <= 5 {
		return 0, fmt.Errorf("header row missing")
	}
	header := rows[5]
	if len(header) < len(expectedHeaders)+1 { // +1 for the \"No.\" column
		return 0, fmt.Errorf("invalid header length")
	}
	for i, name := range expectedHeaders {
		if strings.TrimSpace(header[i+1]) != name {
			return 0, fmt.Errorf("unexpected header %q at column %d", header[i+1], i+2)
		}
	}

	inserted := 0
	for i := 6; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 37 {
			continue
		}
		if strings.TrimSpace(row[1]) == "" || strings.Contains(strings.ToLower(row[1]), "total") || strings.Contains(strings.ToLower(row[1]), "summary") {
			continue
		}

		entry, err := parseShopeeRow(row)
		if err != nil {
			continue
		}
		if err := s.repo.InsertShopeeSettled(ctx, entry); err != nil {
			return inserted, fmt.Errorf("insert row %d: %w", i+1, err)
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

func parseFloat(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", "")
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func parseShopeeRow(row []string) (*models.ShopeeSettled, error) {
	var err error
	res := &models.ShopeeSettled{}
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
	// skip row[28] which stores voucher code in the new template
	if res.Kompensasi, err = parseFloat(row[29]); err != nil {
		return nil, err
	}
	if res.PromoGratisOngkirDariPenjual, err = parseFloat(row[30]); err != nil {
		return nil, err
	}
	res.JasaKirim = row[31]
	res.NamaKurir = row[32]
	if res.PengembalianDanaKePembeli, err = parseFloat(row[33]); err != nil {
		return nil, err
	}
	if res.ProRataKoinYangDitukarkanUntukPengembalianBarang, err = parseFloat(row[34]); err != nil {
		return nil, err
	}
	if res.ProRataVoucherShopeeUntukPengembalianBarang, err = parseFloat(row[35]); err != nil {
		return nil, err
	}
	if res.ProRatedBankPaymentChannelPromotionForReturns, err = parseFloat(row[36]); err != nil {
		return nil, err
	}
	if res.ProRatedShopeePaymentChannelPromotionForReturns, err = parseFloat(row[37]); err != nil {
		return nil, err
	}
	return res, nil
}
