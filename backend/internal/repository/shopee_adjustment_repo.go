package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ShopeeAdjustmentRepo handles CRUD for shopee_adjustments table.
type ShopeeAdjustmentRepo struct{ db DBTX }

func NewShopeeAdjustmentRepo(db DBTX) *ShopeeAdjustmentRepo { return &ShopeeAdjustmentRepo{db: db} }

func (r *ShopeeAdjustmentRepo) Insert(ctx context.Context, a *models.ShopeeAdjustment) error {
	q := `INSERT INTO shopee_adjustments
        (nama_toko, tanggal_penyesuaian, tipe_penyesuaian, alasan_penyesuaian, biaya_penyesuaian, no_pesanan, created_at)
        VALUES (:nama_toko,:tanggal_penyesuaian,:tipe_penyesuaian,:alasan_penyesuaian,:biaya_penyesuaian,:no_pesanan,:created_at)
        ON CONFLICT (no_pesanan, tanggal_penyesuaian, tipe_penyesuaian) DO NOTHING`
	_, err := r.db.NamedExecContext(ctx, q, a)
	return err
}

func (r *ShopeeAdjustmentRepo) Delete(ctx context.Context, order string, t time.Time, typ string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM shopee_adjustments WHERE no_pesanan=$1 AND tanggal_penyesuaian=$2 AND tipe_penyesuaian=$3`, order, t, typ)
	return err
}

func (r *ShopeeAdjustmentRepo) List(ctx context.Context, from, to string) ([]models.ShopeeAdjustment, error) {
	query := `SELECT * FROM shopee_adjustments`
	args := []interface{}{}
	conds := []string{}
	if from != "" {
		conds = append(conds, fmt.Sprintf("tanggal_penyesuaian >= $%d", len(args)+1))
		args = append(args, from)
	}
	if to != "" {
		conds = append(conds, fmt.Sprintf("tanggal_penyesuaian <= $%d", len(args)+1))
		args = append(args, to)
	}
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += " ORDER BY tanggal_penyesuaian DESC"
	var res []models.ShopeeAdjustment
	err := r.db.SelectContext(ctx, &res, query, args...)
	if res == nil {
		res = []models.ShopeeAdjustment{}
	}
	return res, err
}

func (r *ShopeeAdjustmentRepo) ListByOrder(ctx context.Context, order string) ([]models.ShopeeAdjustment, error) {
	var res []models.ShopeeAdjustment
	err := r.db.SelectContext(ctx, &res, `SELECT * FROM shopee_adjustments WHERE no_pesanan=$1 ORDER BY tanggal_penyesuaian`, order)
	if res == nil {
		res = []models.ShopeeAdjustment{}
	}
	return res, err
}
