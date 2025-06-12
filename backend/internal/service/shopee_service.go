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

	inserted := 0
	for i := 5; i < len(rows); i++ {
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
	if res.Kompensasi, err = parseFloat(row[28]); err != nil {
		return nil, err
	}
	if res.PromoGratisOngkirDariPenjual, err = parseFloat(row[29]); err != nil {
		return nil, err
	}
	res.JasaKirim = row[30]
	res.NamaKurir = row[31]
	if res.PengembalianDanaKePembeli, err = parseFloat(row[32]); err != nil {
		return nil, err
	}
	if res.ProRataKoinYangDitukarkanUntukPengembalianBarang, err = parseFloat(row[33]); err != nil {
		return nil, err
	}
	if res.ProRataVoucherShopeeUntukPengembalianBarang, err = parseFloat(row[34]); err != nil {
		return nil, err
	}
	if res.ProRatedBankPaymentChannelPromotionForReturns, err = parseFloat(row[35]); err != nil {
		return nil, err
	}
	if res.ProRatedShopeePaymentChannelPromotionForReturns, err = parseFloat(row[36]); err != nil {
		return nil, err
	}
	return res, nil
}
