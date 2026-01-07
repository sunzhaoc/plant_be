package models

import (
	"time"
)

type OrderItem struct {
	Id             uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	OrderId        uint64    `gorm:"column:order_id"`
	PlantId        uint64    `gorm:"column:plant_id"`
	SkuId          uint64    `gorm:"column:sku_id"`
	PlantName      string    `gorm:"column:plant_name"`
	PlantLatinName string    `gorm:"column:plant_latin_name"`
	SkuSize        string    `gorm:"column:sku_size"`
	MainImgUrl     string    `gorm:"column:main_img_url"`
	Price          float64   `gorm:"column:price"`
	Quantity       uint      `gorm:"column:quantity"`
	CreateTime     time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime     time.Time `gorm:"column:update_time;autoUpdateTime"`
}

func (u OrderItem) TableName() string {
	return "order_items"
}
