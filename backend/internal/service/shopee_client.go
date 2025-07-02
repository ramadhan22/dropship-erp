package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ramadhan22/dropship-erp/backend/internal/config"
	"github.com/ramadhan22/dropship-erp/backend/internal/logutil"
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
	base := cfg.BaseURLShopee
	if base == "" {
		base = "https://partner.test-stable.shopeemobile.com"
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

func (c *ShopeeClient) signSimple(path string, ts int64) string {
	msg := fmt.Sprintf("%s%s%d", c.PartnerID, path, ts)
	h := hmac.New(sha256.New, []byte(c.PartnerKey))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}

// orderDetailResp only includes the order_status field we care about.
type orderDetailResp struct {
	Response struct {
		OrderStatus string `json:"order_status"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// orderDetailExtResp captures additional fields needed for pending balance.
type orderDetailExtResp struct {
	Response struct {
		OrderList []struct {
			OrderSN          string  `json:"order_sn"`
			OrderStatus      string  `json:"order_status"`
			DeliveryTime     int64   `json:"delivery_time"`
			ActualIncome     float64 `json:"actual_income"`
			BuyerTotalAmount float64 `json:"buyer_total_amount"`
		} `json:"order_list"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// orderListResp is the response for get_order_list.
type orderListResp struct {
	Response struct {
		OrderList []struct {
			OrderSN string `json:"order_sn"`
		} `json:"order_list"`
		More bool `json:"more"`
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

// tokenResp captures the token/get response.
type tokenResp struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	ExpireIn     int    `json:"expire_in"`
	RequestID    string `json:"request_id"`
	Error        string `json:"error"`
	Message      string `json:"message"`
}

// RefreshAccessToken fetches a new access token using the refresh token.
func (c *ShopeeClient) RefreshAccessToken(ctx context.Context) error {
	path := "/api/v2/auth/access_token/get"
	ts := time.Now().Unix()
	sign := c.signWithToken(path, ts, c.RefreshToken)

	q := url.Values{}
	q.Set("partner_id", c.PartnerID)
	q.Set("timestamp", fmt.Sprintf("%d", ts))
	q.Set("sign", sign)
	q.Set("shop_id", c.ShopID)
	q.Set("refresh_token", c.RefreshToken)

	body := q.Encode()
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+path, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	log.Printf("ShopeeClient request: POST %s body=%s", c.BaseURL+path, body)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logutil.Errorf("RefreshAccessToken request error: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logutil.Errorf("RefreshAccessToken unexpected status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var out refreshResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	if out.Error != "" {
		logutil.Errorf("RefreshAccessToken API error: %s", out.Error)
		return fmt.Errorf("shopee error: %s", out.Error)
	}
	if out.Response.AccessToken != "" {
		c.AccessToken = out.Response.AccessToken
	}
	return nil
}

// GetAccessToken retrieves a new access token using the authorization code and shop_id.
func (c *ShopeeClient) GetAccessToken(ctx context.Context, code, shopID string) (string, error) {
	path := "/api/v2/auth/token/get"
	ts := time.Now().Unix()
	sign := c.signSimple(path, ts)

	q := url.Values{}
	q.Set("partner_id", c.PartnerID)
	q.Set("timestamp", fmt.Sprintf("%d", ts))
	q.Set("sign", sign)
	q.Set("code", code)
	q.Set("shop_id", shopID)

	body := q.Encode()
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+path, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	log.Printf("ShopeeClient request: POST %s body=%s", c.BaseURL+path, body)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logutil.Errorf("GetAccessToken request error: %v", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logutil.Errorf("GetAccessToken unexpected status %d: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var out tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Error != "" {
		logutil.Errorf("GetAccessToken API error: %s", out.Error)
		return "", fmt.Errorf("shopee error: %s", out.Error)
	}
	return out.AccessToken, nil
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

	urlStr := c.BaseURL + path + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return "", err
	}
	log.Printf("ShopeeClient request: GET %s", urlStr)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logutil.Errorf("GetOrderDetail request error: %v", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logutil.Errorf("GetOrderDetail unexpected status %d: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var out orderDetailResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Error != "" {
		logutil.Errorf("GetOrderDetail API error: %s", out.Error)
		return "", fmt.Errorf("shopee error: %s", out.Error)
	}
	return out.Response.OrderStatus, nil
}

// getOrderDetailExt fetches additional order fields for pending balance.
func (c *ShopeeClient) getOrderDetailExt(ctx context.Context, orderSn string) (*orderDetailExtResp, error) {
	if err := c.RefreshAccessToken(ctx); err != nil {
		return nil, err
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

	urlStr := c.BaseURL + path + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	log.Printf("ShopeeClient request: GET %s", urlStr)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logutil.Errorf("getOrderDetailExt request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logutil.Errorf("getOrderDetailExt unexpected status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var out orderDetailExtResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error != "" {
		logutil.Errorf("getOrderDetailExt API error: %s", out.Error)
		return nil, fmt.Errorf("shopee error: %s", out.Error)
	}
	if len(out.Response.OrderList) == 0 {
		return nil, fmt.Errorf("empty response")
	}
	return &out, nil
}

// GetPendingBalance sums actual income for pending orders.
func (c *ShopeeClient) GetPendingBalance(ctx context.Context, store string) (float64, error) {
	const settlementDelay = 5 * 24 * time.Hour
	const pageSize = 50
	offset := 0
	var total float64

	for {
		if err := c.RefreshAccessToken(ctx); err != nil {
			return 0, err
		}
		path := "/api/v2/order/get_order_list"
		ts := time.Now().Unix()
		sign := c.sign(path, ts)

		q := url.Values{}
		q.Set("partner_id", c.PartnerID)
		q.Set("timestamp", fmt.Sprintf("%d", ts))
		q.Set("sign", sign)
		q.Set("shopid", c.ShopID)
		q.Set("access_token", c.AccessToken)
		q.Set("order_statuses", "[\"READY_TO_SHIP\",\"SHIPPED\",\"COMPLETED\"]")
		q.Set("pagination_offset", strconv.Itoa(offset))
		q.Set("pagination_entries_per_page", strconv.Itoa(pageSize))

		urlStr := c.BaseURL + path + "?" + q.Encode()
		req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			return 0, err
		}
		log.Printf("ShopeeClient request: GET %s", urlStr)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			logutil.Errorf("GetOrderList request error: %v", err)
			return 0, err
		}
		var more bool
		func() {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				logutil.Errorf("GetOrderList unexpected status %d: %s", resp.StatusCode, string(body))
				err = fmt.Errorf("unexpected status %d", resp.StatusCode)
				return
			}
			var out orderListResp
			if e := json.NewDecoder(resp.Body).Decode(&out); e != nil {
				err = e
				return
			}
			if out.Error != "" {
				err = fmt.Errorf("shopee error: %s", out.Error)
				return
			}
			for _, o := range out.Response.OrderList {
				det, e := c.getOrderDetailExt(ctx, o.OrderSN)
				if e != nil {
					err = e
					return
				}
				d := det.Response.OrderList[0]
				inc := d.ActualIncome
				if inc == 0 {
					inc = d.BuyerTotalAmount
				}
				pending := d.OrderStatus == "READY_TO_SHIP" || d.OrderStatus == "SHIPPED" ||
					(d.OrderStatus == "COMPLETED" && time.Unix(d.DeliveryTime, 0).Add(settlementDelay).After(time.Now()))
				if pending {
					total += inc
				}
			}
			more = out.Response.More
		}()
		if err != nil {
			return 0, err
		}
		if !more {
			break
		}
		offset += pageSize
	}
	return total, nil
}
