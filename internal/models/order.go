package models

import "order/domain"

// Order данная модель нужна для маппинга данных из базы в домен (domain.Order)
type Order struct {
	OrderUID string       `db:"order_uid"`
	Data     domain.Order `db:"data"`
}
