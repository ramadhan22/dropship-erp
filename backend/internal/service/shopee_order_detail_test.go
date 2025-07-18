package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/config"
)

func TestFetchShopeeOrderDetail(t *testing.T) {
	partnerID := "1"
	partnerKey := "secret"
	shopID := "2"
	token := "token"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/order/get_order_detail" {
			http.NotFound(w, r)
			return
		}
		ts := r.URL.Query().Get("timestamp")
		pid := r.URL.Query().Get("partner_id")
		sign := r.URL.Query().Get("sign")
		stringToSign := fmt.Sprintf("%s%s%s%s%s", pid, r.URL.Path, ts, token, shopID)
		h := hmac.New(sha256.New, []byte(partnerKey))
		h.Write([]byte(stringToSign))
		expSign := hex.EncodeToString(h.Sum(nil))
		if sign != expSign {
			t.Errorf("invalid signature: %s != %s", sign, expSign)
		}
		if r.URL.Query().Get("order_sn_list") != "123" {
			t.Errorf("expected order_sn_list=123, got %s", r.URL.Query().Get("order_sn_list"))
		}
		if r.URL.Query().Get("shop_id") != shopID {
			t.Errorf("expected shop_id=%s, got %s", shopID, r.URL.Query().Get("shop_id"))
		}
		fmt.Fprint(w, `{"response":{"order_list":[{"order_sn":"123","status":"NEW","checkout_time":1,"update_time":2}]}}`)
	}))
	defer srv.Close()

	cfg := config.ShopeeAPIConfig{
		BaseURLShopee: srv.URL,
		PartnerID:     partnerID,
		PartnerKey:    partnerKey,
		ShopID:        shopID,
	}
	client := NewShopeeClient(cfg)

	detail, err := client.FetchShopeeOrderDetail(context.Background(), token, shopID, "123")
	if err != nil {
		t.Fatalf("FetchShopeeOrderDetail error: %v", err)
	}
	if (*detail)["order_sn"] != "123" || (*detail)["status"] != "NEW" {
		t.Errorf("unexpected detail %+v", detail)
	}
}

func TestFetchShopeeOrderDetails(t *testing.T) {
	partnerID := "1"
	partnerKey := "secret"
	shopID := "2"
	token := "token"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/order/get_order_detail" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("order_sn_list") != "123,456" {
			t.Errorf("expected order_sn_list=123,456, got %s", r.URL.Query().Get("order_sn_list"))
		}
		fmt.Fprint(w, `{"response":{"order_list":[{"order_sn":"123"},{"order_sn":"456"}]}}`)
	}))
	defer srv.Close()

	cfg := config.ShopeeAPIConfig{
		BaseURLShopee: srv.URL,
		PartnerID:     partnerID,
		PartnerKey:    partnerKey,
		ShopID:        shopID,
	}
	client := NewShopeeClient(cfg)

	list, err := client.FetchShopeeOrderDetails(context.Background(), token, shopID, []string{"123", "456"})
	if err != nil {
		t.Fatalf("FetchShopeeOrderDetails error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 results, got %d", len(list))
	}
}
