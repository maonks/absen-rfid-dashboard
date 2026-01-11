package models

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Nama     string `gorm:"size:100"`
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Jabatan  string `gorm:"size:100"`
	Role     string `gorm:"size:50"` // admin, operator, viewer
}
