package models

type User struct {
	Id       uint   `gorm:"primaryKey"` // 手动指定主键
	Username string `gorm:"column:username;type:varchar(50);uniqueIndex;not null"`
	Email    string `gorm:"column:email;type:varchar(100);uniqueIndex;not null"`
	Phone    string `gorm:"column:phone;type:varchar(100);uniqueIndex;not null"`
	Password string `gorm:"column:password;type:varchar(100);not null"`
}

func (u User) TableName() string {
	return "users"
}
