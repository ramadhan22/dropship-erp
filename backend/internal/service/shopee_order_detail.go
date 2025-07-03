package service

// ShopeeOrderDetail represents the raw Shopee order detail object. We keep it
// flexible so new fields from Shopee are automatically preserved.
type ShopeeOrderDetail map[string]any

// orderDetailAPIResponse maps the API JSON structure.
type orderDetailAPIResponse struct {
	Response struct {
		OrderList []ShopeeOrderDetail `json:"order_list"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}
