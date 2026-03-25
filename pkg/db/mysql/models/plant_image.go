package models

import (
	"time"
)

// PlantImage 植物商品图片表模型
type PlantImage struct {
	Id         uint64    `gorm:"primaryKey;autoIncrement;column:id;comment:图片自增ID"`
	PlantId    uint64    `gorm:"column:plant_id;type:bigint unsigned;not null;comment:植物ID"`
	ImgUrl     string    `gorm:"column:img_url;type:varchar(255);not null;comment:图片URL"`
	Sort       int8      `gorm:"column:sort;type:tinyint;not null;default:0;comment:图片排序"`
	CreateTime time.Time `gorm:"column:create_time;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdateTime time.Time `gorm:"column:update_time;type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间"`
}

// TableName 指定数据库表名
func (p PlantImage) TableName() string {
	return "plant_image"
}
