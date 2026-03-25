package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Plants struct {
	Id         uint64          `gorm:"primaryKey;autoIncrement;column:id;comment:植物自增ID"`
	Name       string          `gorm:"column:name;type:varchar(255);default:null;comment:中文名"`
	LatinName  string          `gorm:"column:latin_name;type:varchar(255);default:null;comment:拉丁学名"`
	Genus      string          `gorm:"column:genus;type:varchar(64);default:null;index:idx_genus;comment:属名（默认根据拉丁学名自动生成，可手动覆盖）"`
	MainImgUrl string          `gorm:"column:main_img_url;type:varchar(255);default:null;comment:主图/列表缩略图"`
	MinPrice   decimal.Decimal `gorm:"column:min_price;type:decimal(10,2);default:0.00;comment:起始价格(缓存用于列表展示)"`
	Stock      uint            `gorm:"column:stock;type:int unsigned;not null;default:0;comment:总库存（所有规格库存累加）"`
	IsOnSale   bool            `gorm:"column:is_on_sale;type:tinyint(1);not null;default:1;comment:是否上架"`
	Tag        string          `gorm:"column:tag;type:varchar(64);default:null;comment:标签"`
	Detail     string          `gorm:"column:detail;type:text;default:null;comment:商品详情"`
	CreateTime time.Time       `gorm:"column:create_time;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdateTime time.Time       `gorm:"column:update_time;type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间"`
}

func (u Plants) TableName() string {
	return "plants"
}
