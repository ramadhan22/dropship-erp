package service

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
