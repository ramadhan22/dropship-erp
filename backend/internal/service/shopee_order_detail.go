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
	"os"
	"strconv"
	"time"
)

// ShopeeOrderDetail captures essential fields from Shopee order detail.
type ShopeeOrderDetail struct {
	OrderSN      string `json:"order_sn"`
	Status       string `json:"status"`
	CheckoutTime int64  `json:"checkout_time"`
	UpdateTime   int64  `json:"update_time"`
	// add other fields as needed...
}

// orderDetailAPIResponse maps the API JSON structure.
type orderDetailAPIResponse struct {
	Response struct {
		OrderList []ShopeeOrderDetail `json:"order_list"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// FetchShopeeOrderDetail calls Shopee's get_order_detail endpoint and returns the first result.
// accessToken should be retrieved from your config or DB. TODO: refresh it via
// /api/v2/auth/token/refresh if expired.
func FetchShopeeOrderDetail(ctx context.Context, accessToken, orderSN string) (*ShopeeOrderDetail, error) {
	partnerIDStr := os.Getenv("SHOPEE_PARTNER_ID")
	partnerKey := os.Getenv("SHOPEE_PARTNER_KEY")
	shopIDStr := os.Getenv("SHOPEE_SHOP_ID")
	baseURL := os.Getenv("SHOPEE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://partner.shopeemobile.com"
	}

	if partnerIDStr == "" || partnerKey == "" || shopIDStr == "" {
		return nil, fmt.Errorf("missing Shopee credentials")
	}

	partnerID, err := strconv.ParseInt(partnerIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid SHOPEE_PARTNER_ID: %w", err)
	}
	if _, err := strconv.ParseInt(shopIDStr, 10, 64); err != nil {
		return nil, fmt.Errorf("invalid SHOPEE_SHOP_ID: %w", err)
	}

	path := "/api/v2/order/get_order_detail"
	ts := time.Now().Unix()

	stringToSign := fmt.Sprintf("%d%s%d", partnerID, path, ts)
	mac := hmac.New(sha256.New, []byte(partnerKey))
	mac.Write([]byte(stringToSign))
	signature := hex.EncodeToString(mac.Sum(nil))

	q := url.Values{}
	q.Set("partner_id", partnerIDStr)
	q.Set("timestamp", fmt.Sprintf("%d", ts))
	q.Set("sign", signature)
	q.Set("shop_id", shopIDStr)
	q.Set("access_token", accessToken)
	q.Set("order_sn_list", orderSN)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var out orderDetailAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error != "" {
		return nil, fmt.Errorf("shopee error: %s", out.Error)
	}
	if len(out.Response.OrderList) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	return &out.Response.OrderList[0], nil
}
