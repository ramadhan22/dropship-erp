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
	if list == nil {
		list = []models.JenisChannel{}
	}
	return list, nil
}

// ListStoresByChannel returns stores belonging to a jenis_channel.
func (r *ChannelRepo) ListStoresByChannel(ctx context.Context, channelID int64) ([]models.Store, error) {
	var list []models.Store
	if err := r.db.SelectContext(ctx, &list, `SELECT * FROM stores WHERE jenis_channel_id=$1 ORDER BY store_id`, channelID); err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.Store{}
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
	if list == nil {
		list = []models.Store{}
	}
	return list, nil
}

// GetStoreByID fetches a store row by ID.
func (r *ChannelRepo) GetStoreByID(ctx context.Context, id int64) (*models.Store, error) {
	var st models.Store
	if err := r.db.GetContext(ctx, &st, `SELECT * FROM stores WHERE store_id=$1`, id); err != nil {
		return nil, err
	}
	return &st, nil
}

// GetStoreByName fetches a store row by nama_toko.
func (r *ChannelRepo) GetStoreByName(ctx context.Context, name string) (*models.Store, error) {
	var st models.Store
	if err := r.db.GetContext(ctx, &st, `SELECT * FROM stores WHERE nama_toko=$1`, name); err != nil {
		return nil, err
	}
	return &st, nil
}

// ListAllStores returns all stores joined with their channel names.
func (r *ChannelRepo) ListAllStores(ctx context.Context) ([]models.StoreWithChannel, error) {
	var list []models.StoreWithChannel
	query := `SELECT st.*, jc.jenis_channel FROM stores st
                JOIN jenis_channels jc ON st.jenis_channel_id = jc.jenis_channel_id
                ORDER BY st.store_id`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.StoreWithChannel{}
	}
	return list, nil
}

// UpdateStore modifies an existing store row.
func (r *ChannelRepo) UpdateStore(ctx context.Context, s *models.Store) error {
	_, err := r.db.ExecContext(ctx, `UPDATE stores SET nama_toko=$1, jenis_channel_id=$2, code_id=$3, shop_id=$4, access_token=$5, refresh_token=$6, expire_in=$7, request_id=$8, last_updated=$9 WHERE store_id=$10`,
		s.NamaToko, s.JenisChannelID, s.CodeID, s.ShopID, s.AccessToken, s.RefreshToken, s.ExpireIn, s.RequestID, s.LastUpdated, s.StoreID)
	return err
}

// DeleteStore removes a store by ID.
func (r *ChannelRepo) DeleteStore(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM stores WHERE store_id=$1`, id)
	return err
}

// GetStoresWithTokens returns all stores that have access tokens
func (r *ChannelRepo) GetStoresWithTokens(ctx context.Context) ([]models.Store, error) {
	var list []models.Store
	query := `SELECT * FROM stores WHERE access_token IS NOT NULL AND shop_id IS NOT NULL ORDER BY store_id`
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, err
	}
	if list == nil {
		list = []models.Store{}
	}
	return list, nil
}
