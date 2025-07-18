package service

import (
	"bytes"
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
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
	shopeego "github.com/teacat/shopeego"
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
	rateLimiter  *RateLimiter
	retryConfig  RetryConfig
}

// RetryConfig holds retry mechanism configuration
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
}

// NewShopeeClient constructs a ShopeeClient from configuration.
func NewShopeeClient(cfg config.ShopeeAPIConfig) *ShopeeClient {
	base := cfg.BaseURLShopee
	if base == "" {
		base = "https://partner.test-stable.shopeemobile.com"
	}

	// Create rate limiter for 1000 requests per hour (Shopee API limit)
	rateLimiter := NewRateLimiter(1000, time.Hour/1000)

	retryConfig := RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   time.Second,
	}

	return &ShopeeClient{
		BaseURL:      base,
		PartnerID:    cfg.PartnerID,
		PartnerKey:   cfg.PartnerKey,
		ShopID:       cfg.ShopID,
		AccessToken:  cfg.AccessToken,
		RefreshToken: cfg.RefreshToken,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		rateLimiter:  rateLimiter,
		retryConfig:  retryConfig,
	}
}

// NewShopeeClientWithConfig constructs a ShopeeClient with custom rate limiting and retry configuration
func NewShopeeClientWithConfig(cfg config.ShopeeAPIConfig, rateLimit int, maxAttempts int, baseDelay time.Duration) *ShopeeClient {
	client := NewShopeeClient(cfg)
	client.rateLimiter = NewRateLimiter(rateLimit, time.Hour/time.Duration(rateLimit))
	client.retryConfig = RetryConfig{
		MaxAttempts: maxAttempts,
		BaseDelay:   baseDelay,
	}
	return client
}

func (c *ShopeeClient) signWithToken(path string, ts int64, token string) string {
	msg := fmt.Sprintf("%s%s%d%s%s", c.PartnerID, path, ts, token, c.ShopID)
	h := hmac.New(sha256.New, []byte(c.PartnerKey))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))
}

// signWithTokenShop generates a signature using the provided token and shop ID.
// This matches Shopee's specification for endpoints that require an access token.
func (c *ShopeeClient) signWithTokenShop(path string, ts int64, token, shopID string) string {
	msg := fmt.Sprintf("%s%s%d%s%s", c.PartnerID, path, ts, token, shopID)
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

// ========== Optimized HTTP Request Methods ==========

// makeRequestWithRetry executes HTTP requests with rate limiting and retry logic
func (c *ShopeeClient) makeRequestWithRetry(ctx context.Context, method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	// Apply rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit timeout: %w", err)
	}

	var resp *http.Response
	var err error

	for attempt := 1; attempt <= c.retryConfig.MaxAttempts; attempt++ {
		// Create request
		req, reqErr := http.NewRequestWithContext(ctx, method, url, body)
		if reqErr != nil {
			return nil, fmt.Errorf("failed to create request: %w", reqErr)
		}

		// Set headers
		for key, value := range headers {
			req.Header.Set(key, value)
		}

		// Log request
		log.Printf("ShopeeClient request (attempt %d/%d): %s %s", attempt, c.retryConfig.MaxAttempts, method, url)

		// Execute request
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode < 500 {
			// Success or client error (4xx) - don't retry client errors
			return resp, nil
		}

		// Close response body if present before retry
		if resp != nil {
			resp.Body.Close()
		}

		// Don't sleep after the last attempt
		if attempt < c.retryConfig.MaxAttempts {
			// Exponential backoff: baseDelay * 2^(attempt-1)
			delay := c.retryConfig.BaseDelay * time.Duration(1<<(attempt-1))
			log.Printf("Request failed (attempt %d/%d), retrying in %v. Error: %v",
				attempt, c.retryConfig.MaxAttempts, delay, err)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue to next attempt
			}
		}
	}

	return resp, fmt.Errorf("request failed after %d attempts: %w", c.retryConfig.MaxAttempts, err)
}

// GetRateLimiterStats returns current rate limiter statistics
func (c *ShopeeClient) GetRateLimiterStats() (availableTokens int, maxTokens int) {
	return c.rateLimiter.GetStats()
}

// TokenValidationInterface provides methods for token validation
type TokenValidationInterface interface {
	GetStoreByName(ctx context.Context, name string) (*models.Store, error)
	UpdateStore(ctx context.Context, s *models.Store) error
}

// ensureTokenValidForStore checks token expiration for a store and refreshes if needed
func (c *ShopeeClient) ensureTokenValidForStore(ctx context.Context, store *models.Store, repo TokenValidationInterface) error {
	if repo == nil {
		return fmt.Errorf("missing store repository")
	}
	log.Printf("ensureTokenValidForStore for store %s", store.NamaToko)

	// Parse timezone for proper token expiration calculation
	loc, _ := time.LoadLocation("Asia/Jakarta")
	reinterpreted := time.Date(
		store.LastUpdated.Year(), store.LastUpdated.Month(), store.LastUpdated.Day(),
		store.LastUpdated.Hour(), store.LastUpdated.Minute(), store.LastUpdated.Second(), store.LastUpdated.Nanosecond(),
		loc,
	)
	exp := reinterpreted.Add(time.Duration(*store.ExpireIn) * time.Second)

	// Check if required fields are available
	if store.RefreshToken == nil {
		return fmt.Errorf("missing refresh token for store %s", store.NamaToko)
	}
	if store.ShopID == nil || *store.ShopID == "" {
		return fmt.Errorf("missing shop id for store %s", store.NamaToko)
	}

	// Check if token is still valid (not expired)
	if store.ExpireIn != nil && store.LastUpdated != nil {
		if time.Now().Before(exp.Local()) {
			log.Printf("Token for store %s is still valid until %v", store.NamaToko, exp)
			return nil
		}
	}

	// Token is expired, refresh it
	log.Printf("Token for store %s is expired, refreshing", store.NamaToko)
	oldShopID := c.ShopID
	oldRefreshToken := c.RefreshToken

	// Temporarily set client credentials for refresh
	c.ShopID = *store.ShopID
	c.RefreshToken = *store.RefreshToken

	resp, err := c.RefreshAccessToken(ctx)
	if err != nil {
		// Restore old credentials on error
		c.ShopID = oldShopID
		c.RefreshToken = oldRefreshToken
		return fmt.Errorf("failed to refresh token for store %s: %w", store.NamaToko, err)
	}

	// Update store with new token information
	store.AccessToken = &resp.Response.AccessToken
	if resp.Response.RefreshToken != "" {
		store.RefreshToken = &resp.Response.RefreshToken
	}
	store.ExpireIn = &resp.Response.ExpireIn
	store.RequestID = &resp.Response.RequestID
	now := time.Now()
	store.LastUpdated = &now

	// Save updated store
	if err := repo.UpdateStore(ctx, store); err != nil {
		log.Printf("Warning: failed to update store token in database: %v", err)
		// Don't fail the operation, just log the warning
	}

	// Restore original client credentials
	c.ShopID = oldShopID
	c.RefreshToken = oldRefreshToken

	log.Printf("Successfully refreshed token for store %s", store.NamaToko)
	return nil
}

// ========== End Optimized Methods ==========

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
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpireIn     int    `json:"expire_in"`
		RequestID    string `json:"request_id"`
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
func (c *ShopeeClient) RefreshAccessToken(ctx context.Context) (*refreshResp, error) {
	log.Printf("Refreshing access token for shop %s", c.ShopID)
	if c.ShopID == "" {
		return nil, fmt.Errorf("shop_id is empty")
	}

	partnerID, err := strconv.ParseInt(c.PartnerID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid partner id: %w", err)
	}
	shopID, err := strconv.ParseInt(c.ShopID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid shop id: %w", err)
	}

	reqData := &shopeego.RefreshAccessTokenRequest{
		RefreshToken: c.RefreshToken,
		ShopID:       shopID,
		PartnerID:    partnerID,
		Timestamp:    int(time.Now().Unix()),
	}
	body, _ := json.Marshal(reqData)
	urlStr := c.BaseURL + "/api/v2/auth/access_token/get"
	log.Printf("ShopeeClient request: POST %s body=%s", urlStr, string(body))

	opts := &shopeego.ClientOptions{
		Secret:    c.PartnerKey,
		IsSandbox: strings.Contains(c.BaseURL, "uat") || strings.Contains(c.BaseURL, "test"),
		Version:   shopeego.ClientVersionV2,
	}
	sc := shopeego.NewClient(opts)
	resp, err := sc.RefreshAccessToken(reqData)
	if err != nil {
		logutil.Errorf("RefreshAccessToken request error: %v", err)
		return nil, err
	}
	out := refreshResp{}
	out.Response.AccessToken = resp.AccessToken
	out.Response.RefreshToken = resp.RefreshToken
	out.Response.ExpireIn = resp.ExpireIn
	out.Response.RequestID = resp.RequestID
	out.Error = resp.Error

	if resp.AccessToken != "" {
		c.AccessToken = resp.AccessToken
	}
	if resp.RefreshToken != "" {
		c.RefreshToken = resp.RefreshToken
	}
	return &out, nil
}

// GetAccessToken retrieves a new access token using the authorization code and shop_id.
func (c *ShopeeClient) GetAccessToken(ctx context.Context, code, shopID string) (*tokenResp, error) {
	path := "/api/v2/auth/token/get"
	ts := time.Now().Unix()
	sign := c.signSimple(path, ts)

	q := url.Values{}
	q.Set("partner_id", c.PartnerID)
	q.Set("timestamp", fmt.Sprintf("%d", ts))
	q.Set("sign", sign)

	urlStr := c.BaseURL + path + "?" + q.Encode()

	payload := map[string]string{
		"shop_id": shopID,
		"code":    code,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", urlStr, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	log.Printf("ShopeeClient request: POST %s body=%s", urlStr, string(body))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logutil.Errorf("GetAccessToken request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logutil.Errorf("GetAccessToken unexpected status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var out tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error != "" {
		logutil.Errorf("GetAccessToken API error: %s", out.Error)
		return nil, fmt.Errorf("shopee error: %s", out.Error)
	}
	return &out, nil
}

// FetchShopeeOrderDetail fetches detailed order info using the provided access
// token and shop id. It mirrors the standalone FetchShopeeOrderDetail function
// but uses credentials from the ShopeeClient, similar to GetAccessToken.
func (c *ShopeeClient) FetchShopeeOrderDetail(ctx context.Context, accessToken, shopID, orderSN string) (*ShopeeOrderDetail, error) {
	path := "/api/v2/order/get_order_detail"
	ts := time.Now().Unix()
	sign := c.signWithTokenShop(path, ts, accessToken, shopID)

	q := url.Values{}
	q.Set("partner_id", c.PartnerID)
	q.Set("timestamp", fmt.Sprintf("%d", ts))
	q.Set("sign", sign)
	q.Set("shop_id", shopID)
	q.Set("access_token", accessToken)
	q.Set("order_sn_list", orderSN)
	q.Set("response_optional_fields", "buyer_user_id,buyer_username,estimated_shipping_fee,recipient_address,actual_shipping_fee ,goods_to_declare,note,note_update_time,item_list,pay_time,dropshipper, dropshipper_phone,split_up,buyer_cancel_reason,cancel_by,cancel_reason,actual_shipping_fee_confirmed,buyer_cpf_id,fulfillment_flag,pickup_done_time,package_list,shipping_carrier,payment_method,total_amount,buyer_username,invoice_data,order_chargeable_weight_gram,return_request_due_date,edt")

	urlStr := c.BaseURL + path + "?" + q.Encode()
	log.Printf("ShopeeClient request: GET %s", urlStr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logutil.Errorf("FetchShopeeOrderDetail request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logutil.Errorf("FetchShopeeOrderDetail unexpected status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var out orderDetailAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error != "" {
		logutil.Errorf("FetchShopeeOrderDetail API error: %s", out.Error)
		return nil, fmt.Errorf("shopee error: %s", out.Error)
	}
	if len(out.Response.OrderList) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	return &out.Response.OrderList[0], nil
}

// FetchShopeeOrderDetails fetches details for multiple order_sn values in a single request.
// The API accepts up to 50 comma separated order numbers.
func (c *ShopeeClient) FetchShopeeOrderDetails(ctx context.Context, accessToken, shopID string, orderSNs []string) ([]ShopeeOrderDetail, error) {
	if len(orderSNs) == 0 {
		return nil, nil
	}
	path := "/api/v2/order/get_order_detail"
	ts := time.Now().Unix()
	sign := c.signWithTokenShop(path, ts, accessToken, shopID)

	q := url.Values{}
	q.Set("partner_id", c.PartnerID)
	q.Set("timestamp", fmt.Sprintf("%d", ts))
	q.Set("sign", sign)
	q.Set("shop_id", shopID)
	q.Set("access_token", accessToken)
	q.Set("order_sn_list", strings.Join(orderSNs, ","))
	q.Set("response_optional_fields", "buyer_user_id,buyer_username,estimated_shipping_fee,recipient_address,actual_shipping_fee ,goods_to_declare,note,note_update_time,item_list,pay_time,dropshipper, dropshipper_phone,split_up,buyer_cancel_reason,cancel_by,cancel_reason,actual_shipping_fee_confirmed,buyer_cpf_id,fulfillment_flag,pickup_done_time,package_list,shipping_carrier,payment_method,total_amount,buyer_username,invoice_data,order_chargeable_weight_gram,return_request_due_date,edt")

	urlStr := c.BaseURL + path + "?" + q.Encode()
	log.Printf("ShopeeClient request: GET %s", urlStr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logutil.Errorf("FetchShopeeOrderDetails request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logutil.Errorf("FetchShopeeOrderDetails unexpected status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var out orderDetailAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error != "" {
		logutil.Errorf("FetchShopeeOrderDetails API error: %s", out.Error)
		return nil, fmt.Errorf("shopee error: %s", out.Error)
	}
	if len(out.Response.OrderList) == 0 {
		return nil, fmt.Errorf("empty response")
	}
	return out.Response.OrderList, nil
}

// GetEscrowDetail retrieves escrow information for an order using a direct HTTP
// call. The accessToken and shopID parameters should belong to the store that
// owns the order.
func (c *ShopeeClient) GetEscrowDetail(ctx context.Context, accessToken, shopID, orderSN string) (*ShopeeEscrowDetail, error) {
	log.Printf("ShopeeClient GetEscrowDetail order=%s shop=%s", orderSN, shopID)

	path := "/api/v2/payment/get_escrow_detail"
	ts := time.Now().Unix()
	sign := c.signWithTokenShop(path, ts, accessToken, shopID)

	q := url.Values{}
	q.Set("partner_id", c.PartnerID)
	q.Set("timestamp", fmt.Sprintf("%d", ts))
	q.Set("sign", sign)
	q.Set("shop_id", shopID)
	q.Set("access_token", accessToken)
	q.Set("order_sn", orderSN)

	urlStr := c.BaseURL + path + "?" + q.Encode()
	log.Printf("ShopeeClient request: GET %s", urlStr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logutil.Errorf("GetEscrowDetail request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("GetEscrowDetail response status: %d", resp.StatusCode)
	log.Printf("GetEscrowDetail response headers: %v", resp.Header)
	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("GetEscrowDetail response body: %s", string(bodyBytes))
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset body for decoder

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logutil.Errorf("GetEscrowDetail unexpected status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var out struct {
		Response ShopeeEscrowDetail `json:"response"`
		Error    string             `json:"error"`
		Message  string             `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error != "" {
		logutil.Errorf("GetEscrowDetail API error: %s", out.Error)
		return nil, fmt.Errorf("shopee error: %s", out.Error)
	}
	log.Printf("GetEscrowDetail response: %+v", out.Response)
	return &out.Response, nil
}

// FetchShopeeEscrowDetails retrieves escrow information for multiple orders in a single request.
// The API accepts up to 50 order numbers in the request body.
func (c *ShopeeClient) FetchShopeeEscrowDetails(ctx context.Context, accessToken, shopID string, orderSNs []string) (map[string]ShopeeEscrowDetail, error) {
	if len(orderSNs) == 0 {
		return map[string]ShopeeEscrowDetail{}, nil
	}

	path := "/api/v2/payment/get_escrow_detail_batch"
	ts := time.Now().Unix()
	sign := c.signWithTokenShop(path, ts, accessToken, shopID)

	q := url.Values{}
	q.Set("partner_id", c.PartnerID)
	q.Set("timestamp", fmt.Sprintf("%d", ts))
	q.Set("sign", sign)
	q.Set("shop_id", shopID)
	q.Set("access_token", accessToken)

	body := map[string]any{"order_sn_list": orderSNs}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, err
	}

	urlStr := c.BaseURL + path + "?" + q.Encode()
	log.Printf("ShopeeClient request: POST %s", urlStr)
	log.Printf("ShopeeClient request body: %s", buf.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logutil.Errorf("FetchShopeeEscrowDetails request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logutil.Errorf("FetchShopeeEscrowDetails unexpected status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var out struct {
		Response []struct {
			EscrowDetail ShopeeEscrowDetail `json:"escrow_detail"`
			OrderSN      string             `json:"order_sn"`
		} `json:"response"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error != "" {
		logutil.Errorf("FetchShopeeEscrowDetails API error: %s", out.Error)
		return nil, fmt.Errorf("shopee error: %s", out.Error)
	}

	log.Printf("FetchShopeeEscrowDetails response: %+v", out)
	log.Printf("FetchShopeeEscrowDetails response count: %d", len(out.Response))
	log.Printf("FetchShopeeEscrowDetails response order_sns: %v", orderSNs)

	res := make(map[string]ShopeeEscrowDetail, len(out.Response))
	for _, item := range out.Response {
		sn := strings.TrimSpace(item.OrderSN)
		// If OrderSN is empty, try fallback from EscrowDetail.OrderSN
		if sn == "" {
			// Attempt fallback if EscrowDetail is a struct and has OrderSN field
			if item.EscrowDetail["order_sn"].(string) != "" {
				sn = strings.TrimSpace(item.EscrowDetail["order_sn"].(string))
			}
			// If still blank, try fallback from EscrowDetail as map[string]interface{} (in case of map type)
			if sn == "" {
				if m, ok := any(item.EscrowDetail).(map[string]interface{}); ok {
					if orderSNVal, ok2 := m["order_sn"].(string); ok2 && orderSNVal != "" {
						sn = strings.TrimSpace(orderSNVal)
					}
				}
			}
		}
		if sn == "" {
			log.Printf("WARNING: EscrowDetail with empty order_sn: %+v", item)
			continue
		}
		res[sn] = item.EscrowDetail
	}
	log.Printf("FetchShopeeEscrowDetails mapped response: %v", res)
	return res, nil
}

// GetOrderDetail fetches order detail for a given order_sn and returns the status.
func (c *ShopeeClient) GetOrderDetail(ctx context.Context, orderSn string) (string, error) {
	if _, err := c.RefreshAccessToken(ctx); err != nil {
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
	if _, err := c.RefreshAccessToken(ctx); err != nil {
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
		if _, err := c.RefreshAccessToken(ctx); err != nil {
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

// GetWalletTransactionList calls Shopee get_wallet_transaction_list API.
func (c *ShopeeClient) GetWalletTransactionList(ctx context.Context, accessToken, shopID string, p WalletTransactionParams) (*WalletTransactionList, error) {
	path := "/api/v2/payment/get_wallet_transaction_list"
	ts := time.Now().Unix()
	sign := c.signWithTokenShop(path, ts, accessToken, shopID)

	q := url.Values{}
	q.Set("partner_id", c.PartnerID)
	q.Set("timestamp", fmt.Sprintf("%d", ts))
	q.Set("sign", sign)
	q.Set("shop_id", shopID)
	q.Set("access_token", accessToken)
	q.Set("page_no", strconv.Itoa(p.PageNo))
	q.Set("page_size", strconv.Itoa(p.PageSize))
	if p.CreateTimeFrom != nil {
		q.Set("create_time_from", strconv.FormatInt(*p.CreateTimeFrom, 10))
	}
	if p.CreateTimeTo != nil {
		q.Set("create_time_to", strconv.FormatInt(*p.CreateTimeTo, 10))
	}
	if p.WalletType != "" {
		q.Set("wallet_type", p.WalletType)
	}
	if p.TransactionType != "" {
		q.Set("transaction_type", p.TransactionType)
	}
	if p.MoneyFlow != "" {
		q.Set("money_flow", p.MoneyFlow)
	}
	if p.TransactionTabType != "" {
		q.Set("transaction_tab_type", p.TransactionTabType)
	}

	urlStr := c.BaseURL + path + "?" + q.Encode()
	log.Printf("ShopeeClient request: GET %s", urlStr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logutil.Errorf("GetWalletTransactionList request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logutil.Errorf("GetWalletTransactionList unexpected status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var out struct {
		Response struct {
			TransactionList []WalletTransaction `json:"transaction_list"`
			More            bool                `json:"more"`
		} `json:"response"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.Error != "" {
		logutil.Errorf("GetWalletTransactionList API error: %s", out.Error)
		return nil, fmt.Errorf("shopee error: %s", out.Error)
	}
	return &WalletTransactionList{Transactions: out.Response.TransactionList, more: out.Response.More}, nil
}

// GetReturnList fetches returns from Shopee's get_return_list API
func (c *ShopeeClient) GetReturnList(ctx context.Context, accessToken, shopID string, params map[string]string) (*models.ShopeeReturnResponse, error) {
	path := "/api/v2/returns/get_return_list"
	ts := time.Now().Unix()
	sign := c.signWithTokenShop(path, ts, accessToken, shopID)

	q := url.Values{}
	q.Set("partner_id", c.PartnerID)
	q.Set("timestamp", strconv.FormatInt(ts, 10))
	q.Set("sign", sign)
	q.Set("shop_id", shopID)
	q.Set("access_token", accessToken)

	// Add filter parameters
	for key, value := range params {
		if value != "" {
			q.Set(key, value)
		}
	}

	urlStr := c.BaseURL + path + "?" + q.Encode()
	log.Printf("ShopeeClient request: GET %s", urlStr)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logutil.Errorf("GetReturnList request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logutil.Errorf("GetReturnList unexpected status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result models.ShopeeReturnResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		logutil.Errorf("GetReturnList API error: %s", result.Error)
		return nil, fmt.Errorf("shopee error: %s", result.Error)
	}

	log.Printf("GetReturnList response: found %d returns", len(result.Response.Return))
	return &result, nil
}
