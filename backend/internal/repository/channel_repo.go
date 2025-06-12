package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// ChannelRepo handles CRUD operations for jenis_channels and stores.
type ChannelRepo struct {
	db *sqlx.DB
}

// NewChannelRepo constructs a ChannelRepo.
func NewChannelRepo(db *sqlx.DB) *ChannelRepo {
	return &ChannelRepo{db: db}
}

// CreateJenisChannel inserts a new jenis_channel row and returns the generated ID.
func (r *ChannelRepo) CreateJenisChannel(ctx context.Context, c *models.JenisChannel) (int64, error) {
	query := `INSERT INTO jenis_channels (jenis_channel) VALUES ($1) RETURNING jenis_channel_id`
	var id int64
	if err := r.db.QueryRowContext(ctx, query, c.JenisChannel).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// CreateStore inserts a new store row and returns the generated ID.
func (r *ChannelRepo) CreateStore(ctx context.Context, s *models.Store) (int64, error) {
	query := `INSERT INTO stores (jenis_channel_id, nama_toko) VALUES ($1, $2) RETURNING store_id`
	var id int64
	if err := r.db.QueryRowContext(ctx, query, s.JenisChannelID, s.NamaToko).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// ListJenisChannels returns all jenis_channels.
func (r *ChannelRepo) ListJenisChannels(ctx context.Context) ([]models.JenisChannel, error) {
	var list []models.JenisChannel
	if err := r.db.SelectContext(ctx, &list, `SELECT * FROM jenis_channels ORDER BY jenis_channel_id`); err != nil {
		return nil, err
	}
	return list, nil
}

// ListStoresByChannel returns stores belonging to a jenis_channel.
func (r *ChannelRepo) ListStoresByChannel(ctx context.Context, channelID int64) ([]models.Store, error) {
	var list []models.Store
	if err := r.db.SelectContext(ctx, &list, `SELECT * FROM stores WHERE jenis_channel_id=$1 ORDER BY store_id`, channelID); err != nil {
		return nil, err
	}
	return list, nil
}

// ListStoresByChannelName returns stores by joining with jenis_channels using the channel name.
func (r *ChannelRepo) ListStoresByChannelName(ctx context.Context, channelName string) ([]models.Store, error) {
        var list []models.Store
        query := `SELECT st.* FROM stores st
                JOIN jenis_channels jc ON st.jenis_channel_id = jc.jenis_channel_id
                WHERE jc.jenis_channel = $1
                ORDER BY st.store_id`
        if err := r.db.SelectContext(ctx, &list, query, channelName); err != nil {
                return nil, err
        }
        return list, nil
}
