package models

import "time"

type Login struct {
	Id        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string `gorm:"type:varchar(20);uniqueIndex;not null" json:"username"`
	Password  string `gorm:"type:text;not null" json:"password"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
