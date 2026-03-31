package models

import (
	"time"
)

// LotteryWinningRecord 抽奖活动中奖表
type LotteryWinningRecord struct {
	Id          uint64    `gorm:"primaryKey;autoIncrement;column:id;comment:自增ID"`
	UserId      uint64    `gorm:"column:user_id;not null;comment:中奖用户ID"`
	ActivityId  uint64    `gorm:"column:activity_id;not null;comment:活动ID"`
	WinCode     string    `gorm:"column:win_code;type:varchar(64);not null;comment:中奖码"` // 新增字段
	WinningTime time.Time `gorm:"column:winning_time;not null;comment:中奖时间"`
	IsReceived  bool      `gorm:"column:is_received;type:tinyint;not null;default:0;comment:是否已领取 0-未领取 1-已领取"`
	Remark      string    `gorm:"column:remark;type:varchar(512);default:'';comment:备注"`
	CreateTime  time.Time `gorm:"column:create_time;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdateTime  time.Time `gorm:"column:update_time;type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime;comment:更新时间"`
}

// TableName 指定数据库表名
func (LotteryWinningRecord) TableName() string {
	return "lottery_winning_record"
}
