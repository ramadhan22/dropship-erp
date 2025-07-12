package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ramadhan22/dropship-erp/backend/internal/config"
)

func TestFetchShopeeEscrowDetails(t *testing.T) {
	partnerID := "1"
	partnerKey := "secret"
	shopID := "2"
	token := "token"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/payment/get_escrow_detail_batch" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
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
		var req struct {
			OrderSNList []string `json:"order_sn_list"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if !reflect.DeepEqual(req.OrderSNList, []string{"123", "456"}) {
			t.Errorf("unexpected body %+v", req.OrderSNList)
		}
		fmt.Fprint(w, `{"response":[{"order_sn":"123","escrow_detail":{"amount":1}},{"order_sn":"456","escrow_detail":{"amount":2}}]}`)
	}))
	defer srv.Close()

	cfg := config.ShopeeAPIConfig{
		BaseURLShopee: srv.URL,
		PartnerID:     partnerID,
		PartnerKey:    partnerKey,
	}
	client := NewShopeeClient(cfg)

	res, err := client.FetchShopeeEscrowDetails(context.Background(), token, shopID, []string{"123", "456"})
	if err != nil {
		t.Fatalf("FetchShopeeEscrowDetails error: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res))
	}
	if m := res["123"]; m["amount"] != float64(1) {
		t.Errorf("unexpected detail %v", m)
	}
}
