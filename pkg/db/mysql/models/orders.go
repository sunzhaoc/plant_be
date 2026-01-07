package models

import (
	"time"
)

type Orders struct {
	Id              uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	OrderSn         string    `gorm:"column:order_sn;unique"`
	UserId          uint64    `gorm:"column:user_id"`
	TotalAmount     float64   `gorm:"column:total_amount"`
	PayAmount       float64   `gorm:"column:pay_amount"`
	OrderStatus     int       `gorm:"column:order_status;default:0"`
	ReceiverName    string    `gorm:"column:receiver_name"`
	ReceiverPhone   string    `gorm:"column:receiver_phone"`
	ReceiverAddress string    `gorm:"column:receiver_address"`
	CreateTime      time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime      time.Time `gorm:"column:update_time;autoUpdateTime"`
}

func (o Orders) TableName() string {
	return "orders"
}
