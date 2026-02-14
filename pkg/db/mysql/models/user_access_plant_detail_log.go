package models

import (
	"time"
)

type UserAccessPlantDetailLog struct {
	Id         uint      `gorm:"column:id;primaryKey;autoIncrement;type:bigint unsigned;comment:主键ID"`
	UserId     uint      `gorm:"column:user_id;not null;type:bigint;comment:用户ID"`
	PlantId    uint      `gorm:"column:plant_id;not null;type:bigint;comment:植物ID"`
	Ip         string    `gorm:"column:ip;not null;type:varchar(64);default:'';comment:用户访问IP地址"`
	Country    string    `gorm:"column:country;type:varchar(64);default:'';comment:国家"`
	Province   string    `gorm:"column:province;type:varchar(64);default:'';comment:省份"`
	City       string    `gorm:"column:city;type:varchar(64);default:'';comment:城市"`
	Area       string    `gorm:"column:area;type:varchar(64);default:'';comment:区域"`
	Isp        string    `gorm:"column:isp;type:varchar(64);default:'';comment:运营商"`
	CreateTime time.Time `gorm:"column:create_time;not null;type:datetime;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdateTime time.Time `gorm:"column:update_time;not null;type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
}

func (u UserAccessPlantDetailLog) TableName() string {
	return "user_access_plant_detail_log"
}
