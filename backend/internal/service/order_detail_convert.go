package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/ramadhan22/dropship-erp/backend/internal/models"
)

// helper conversion functions
func asString(m map[string]any, key string) *string {
	if v, ok := m[key]; ok {
		switch s := v.(type) {
		case string:
			if s == "" {
				return nil
			}
			str := s
			return &str
		case float64:
			str := fmt.Sprintf("%v", s)
			return &str
		}
	}
	return nil
}

func asInt64(m map[string]any, key string) *int64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			i := int64(n)
			return &i
		case string:
			if i, err := strconv.ParseInt(n, 10, 64); err == nil {
				return &i
			}
		}
	}
	return nil
}

func asInt(m map[string]any, key string) *int {
	if v := asInt64(m, key); v != nil {
		i := int(*v)
		return &i
	}
	return nil
}

func asFloat64(m map[string]any, key string) *float64 {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			f := n
			return &f
		case string:
			if f, err := strconv.ParseFloat(n, 64); err == nil {
				return &f
			}
		}
	}
	return nil
}

func asBool(m map[string]any, key string) *bool {
	if v, ok := m[key]; ok {
		switch b := v.(type) {
		case bool:
			val := b
			return &val
		case string:
			if b == "" {
				return nil
			}
			val := b == "true"
			return &val
		}
	}
	return nil
}

func asTimeVal(m map[string]any, key string) *time.Time {
	if v, ok := m[key]; ok {
		switch t := v.(type) {
		case float64:
			tm := time.Unix(int64(t), 0)
			return &tm
		case string:
			if i, err := strconv.ParseInt(t, 10, 64); err == nil {
				tm := time.Unix(i, 0)
				return &tm
			}
			if tm, err := time.Parse(time.RFC3339, t); err == nil {
				return &tm
			}
		}
	}
	return nil
}

func asStringSlice(m map[string]any, key string) []string {
	if v, ok := m[key]; ok {
		switch arr := v.(type) {
		case []interface{}:
			res := make([]string, 0, len(arr))
			for _, a := range arr {
				if s, ok := a.(string); ok {
					res = append(res, s)
				} else {
					res = append(res, fmt.Sprint(a))
				}
			}
			if len(res) > 0 {
				return res
			}
		case []string:
			return arr
		}
	}
	return nil
}

// normalizeOrderDetail converts a ShopeeOrderDetail map into database rows.
func normalizeOrderDetail(orderSN, namaToko string, det ShopeeOrderDetail) (*models.ShopeeOrderDetailRow, []models.ShopeeOrderItemRow, []models.ShopeeOrderPackageRow) {
	m := map[string]any(det)

	row := &models.ShopeeOrderDetailRow{
		OrderSN:      orderSN,
		NamaToko:     namaToko,
		Status:       asString(m, "status"),
		OrderStatus:  asString(m, "order_status"),
		CheckoutTime: asTimeVal(m, "checkout_time"),
		UpdateTime:   asTimeVal(m, "update_time"),
		PayTime:      asTimeVal(m, "pay_time"),
		TotalAmount:  asFloat64(m, "total_amount"),
		Currency:     asString(m, "currency"),

		ActualShippingFeeConfirmed: asBool(m, "actual_shipping_fee_confirmed"),
		BuyerCancelReason:          asString(m, "buyer_cancel_reason"),
		BuyerCPFID:                 asString(m, "buyer_cpf_id"),
		BuyerUserID:                asInt64(m, "buyer_user_id"),
		BuyerUsername:              asString(m, "buyer_username"),
		CancelBy:                   asString(m, "cancel_by"),
		CancelReason:               asString(m, "cancel_reason"),
		COD:                        asBool(m, "cod"),
		CreateTime:                 asTimeVal(m, "create_time"),
		DaysToShip:                 asInt(m, "days_to_ship"),
		Dropshipper:                asString(m, "dropshipper"),
		DropshipperPhone:           asString(m, "dropshipper_phone"),
		EstimatedShippingFee:       asFloat64(m, "estimated_shipping_fee"),
		FulfillmentFlag:            asString(m, "fulfillment_flag"),
		GoodsToDeclare:             asBool(m, "goods_to_declare"),
		MessageToSeller:            asString(m, "message_to_seller"),
		Note:                       asString(m, "note"),
		NoteUpdateTime:             asTimeVal(m, "note_update_time"),
		PickupDoneTime:             asTimeVal(m, "pickup_done_time"),
		Region:                     asString(m, "region"),
		ReverseShippingFee:         asFloat64(m, "reverse_shipping_fee"),
		ShipByDate:                 asTimeVal(m, "ship_by_date"),
		ShippingCarrier:            asString(m, "shipping_carrier"),
		SplitUp:                    asBool(m, "split_up"),
		PaymentMethod:              asString(m, "payment_method"),
		CreatedAt:                  time.Now(),
	}

	if addr, ok := m["recipient_address"].(map[string]any); ok {
		row.RecipientName = asString(addr, "name")
		row.RecipientPhone = asString(addr, "phone")
		row.RecipientFullAddress = asString(addr, "full_address")
		row.RecipientCity = asString(addr, "city")
		row.RecipientDistrict = asString(addr, "district")
		row.RecipientState = asString(addr, "state")
		row.RecipientTown = asString(addr, "town")
		row.RecipientZipcode = asString(addr, "zipcode")
	}

	items := []models.ShopeeOrderItemRow{}
	if list, ok := m["item_list"].([]interface{}); ok {
		for _, it := range list {
			if im, ok := it.(map[string]any); ok {
				items = append(items, models.ShopeeOrderItemRow{
					OrderSN:                orderSN,
					OrderItemID:            asInt64(im, "order_item_id"),
					ItemName:               asString(im, "item_name"),
					ModelOriginalPrice:     asFloat64(im, "model_original_price"),
					ModelQuantityPurchased: asInt(im, "model_quantity_purchased"),
					ItemID:                 asInt64(im, "item_id"),
					ItemSKU:                asString(im, "item_sku"),
					ModelID:                asInt64(im, "model_id"),
					ModelName:              asString(im, "model_name"),
					ModelSKU:               asString(im, "model_sku"),
					ModelDiscountedPrice:   asFloat64(im, "model_discounted_price"),
					Weight:                 asFloat64(im, "weight"),
					PromotionID:            asInt64(im, "promotion_id"),
					PromotionType:          asString(im, "promotion_type"),
					PromotionGroupID:       asInt64(im, "promotion_group_id"),
					AddOnDeal:              asBool(im, "add_on_deal"),
					AddOnDealID:            asInt64(im, "add_on_deal_id"),
					MainItem:               asBool(im, "main_item"),
					IsB2COwnedItem:         asBool(im, "is_b2c_owned_item"),
					IsPrescriptionItem:     asBool(im, "is_prescription_item"),
					Wholesale:              asBool(im, "wholesale"),
					ProductLocationID:      pq.StringArray(asStringSlice(im, "product_location_id")),
					ImageURL:               asString(im, "image_url"),
				})
			}
		}
	}

	packages := []models.ShopeeOrderPackageRow{}
	if plist, ok := m["package_list"].([]interface{}); ok {
		for _, p := range plist {
			if pm, ok := p.(map[string]any); ok {
				packages = append(packages, models.ShopeeOrderPackageRow{
					OrderSN:                    orderSN,
					PackageNumber:              asString(pm, "package_number"),
					LogisticsStatus:            asString(pm, "logistics_status"),
					ShippingCarrier:            asString(pm, "shipping_carrier"),
					LogisticsChannelID:         asInt64(pm, "logistics_channel_id"),
					ParcelChargeableWeightGram: asInt(pm, "parcel_chargeable_weight_gram"),
					AllowSelfDesignAWB:         asBool(pm, "allow_self_design_awb"),
					SortingGroup:               asString(pm, "sorting_group"),
					GroupShipmentID:            asString(pm, "group_shipment_id"),
				})
			}
		}
	}

	return row, items, packages
}
