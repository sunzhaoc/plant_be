package models

import (
	"time"
)

type UserLogin struct {
	Id         uint      `gorm:"column:id;primaryKey;autoIncrement;type:bigint unsigned;comment:主键ID"`
	UserId     uint      `gorm:"column:user_id;not null;type:bigint;comment:用户ID"`
	CreateTime time.Time `gorm:"column:create_time;not null;type:datetime;default:CURRENT_TIMESTAMP;comment:登录时间"`
	UpdateTime time.Time `gorm:"column:update_time;not null;type:datetime;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
}

func (u UserLogin) TableName() string {
	return "user_login"
}
