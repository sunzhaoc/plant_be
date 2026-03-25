package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type PlantSku struct {
	Id         uint64          `gorm:"primaryKey;autoIncrement;column:id;comment:规格自增ID"`
	PlantId    uint64          `gorm:"column:plant_id;type:bigint unsigned;not null;comment:植物ID"`
	Size       string          `gorm:"column:size;type:varchar(64);not null;default:'';comment:规格名称"`
	Price      decimal.Decimal `gorm:"column:price;type:decimal(10,2);not null;comment:规格对应价格"`
	Stock      uint            `gorm:"column:stock;type:int unsigned;not null;default:0;comment:规格库存"`
	Sort       uint8           `gorm:"column:sort;type:tinyint unsigned;not null;default:0;comment:规格排序"`
	CreateTime time.Time       `gorm:"column:create_time;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdateTime time.Time       `gorm:"column:update_time;type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间"`
}

func (u PlantSku) TableName() string {
	return "plant_sku"
}
