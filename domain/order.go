package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

var ErrConvertJSONB = errors.New("cannot convert to JSONB")

// Order основная сущность в приложение, т.е. сам заказ
type Order struct {
	OrderUID    string `json:"order_uid"`
	TrackNumber string `json:"track_number"`
	Entry       string `json:"entry"`
	Delivery    struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Zip     string `json:"zip"`
		City    string `json:"city"`
		Address string `json:"address"`
		Region  string `json:"region"`
		Email   string `json:"email"`
	} `json:"delivery"`
	Payment struct {
		Transaction  string `json:"transaction"`
		RequestID    string `json:"request_id"`
		Currency     string `json:"currency"`
		Provider     string `json:"provider"`
		Amount       int    `json:"amount"`
		PaymentDt    int    `json:"payment_dt"`
		Bank         string `json:"bank"`
		DeliveryCost int    `json:"delivery_cost"`
		GoodsTotal   int    `json:"goods_total"`
		CustomFee    int    `json:"custom_fee"`
	} `json:"payment"`
	Items []struct {
		ChrtID      int    `json:"chrt_id"`
		TrackNumber string `json:"track_number"`
		Price       int    `json:"price"`
		Rid         string `json:"rid"`
		Name        string `json:"name"`
		Sale        int    `json:"sale"`
		Size        string `json:"size"`
		TotalPrice  int    `json:"total_price"`
		NmID        int    `json:"nm_id"`
		Brand       string `json:"brand"`
		Status      int    `json:"status"`
	} `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

func (o *Order) Scan(dst interface{}) error {
	switch src := dst.(type) {
	case string:
		return json.Unmarshal([]byte(src), o)
	case []byte:
		return json.Unmarshal(src, o)
	case nil:
		return nil
	}
	return ErrConvertJSONB
}

func (o Order) Value() (driver.Value, error) {
	j, err := json.Marshal(o)
	if err != nil {
		return `{}`, err
	}
	return string(j), nil
}
