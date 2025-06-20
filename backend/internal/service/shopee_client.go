package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/config"
)

// ShopeeClient handles calls to Shopee partner API.
type ShopeeClient struct {
	BaseURL      string
	PartnerID    string
	PartnerKey   string
	ShopID       string
	AccessToken  string
	RefreshToken string
	httpClient   *http.Client
}

// NewShopeeClient constructs a ShopeeClient from configuration.
func NewShopeeClient(cfg config.ShopeeAPIConfig) *ShopeeClient {
	base := cfg.BaseURL
	if base == "" {
		base = "https://ads.shopeemobile.com"
	}
	return &ShopeeClient{
		BaseURL:      base,
		PartnerID:    cfg.PartnerID,
		PartnerKey:   cfg.PartnerKey,
		ShopID:       cfg.ShopID,
		AccessToken:  cfg.AccessToken,
		RefreshToken: cfg.RefreshToken,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *ShopeeClient) signWithToken(path string, ts int64, token string) string {
	msg := fmt.Sprintf("%s%s%d%s%s", c.PartnerID, path, ts, token, c.ShopID)
	h := hmac.New(sha256.New, []byte(c.PartnerKey))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *ShopeeClient) sign(path string, ts int64) string {
	return c.signWithToken(path, ts, c.AccessToken)
}

// orderDetailResp only includes the order_status field we care about.
type orderDetailResp struct {
	Response struct {
		OrderStatus string `json:"order_status"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// refreshResp captures the access_token response.
type refreshResp struct {
	Response struct {
		AccessToken string `json:"access_token"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// RefreshAccessToken fetches a new access token using the refresh token.
func (c *ShopeeClient) RefreshAccessToken(ctx context.Context) error {
	cfg := config.MustLoadConfig()
	path := "/api/v2/auth/access_token/get"
	ts := time.Now().Unix()
	sign := c.signWithToken(path, ts, c.RefreshToken)

	q := url.Values{}
	q.Set("partner_id", cfg.Shopee.PartnerID)
	q.Set("timestamp", fmt.Sprintf("%d", ts))
	q.Set("sign", sign)
	q.Set("shop_id", cfg.Shopee.ShopID)
	q.Set("refresh_token", cfg.Shopee.RefreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+path, strings.NewReader(q.Encode()))
	fmt.Println(cfg)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var out refreshResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	if out.Error != "" {
		return fmt.Errorf("shopee error: %s", out.Error)
	}
	if out.Response.AccessToken != "" {
		c.AccessToken = out.Response.AccessToken
	}
	return nil
}

// GetOrderDetail fetches order detail for a given order_sn and returns the status.
func (c *ShopeeClient) GetOrderDetail(ctx context.Context, orderSn string) (string, error) {
	if err := c.RefreshAccessToken(ctx); err != nil {
		return "", err
	}
	path := "/api/v2/order/get_order_detail"
	ts := time.Now().Unix()
	sign := c.sign(path, ts)

	q := url.Values{}
	q.Set("partner_id", c.PartnerID)
	q.Set("timestamp", fmt.Sprintf("%d", ts))
	q.Set("sign", sign)
	q.Set("shop_id", c.ShopID)
	q.Set("access_token", c.AccessToken)
	q.Set("order_sn_list", orderSn)

	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+path+"?"+q.Encode(), nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var out orderDetailResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Error != "" {
		return "", fmt.Errorf("shopee error: %s", out.Error)
	}
	return out.Response.OrderStatus, nil
}
