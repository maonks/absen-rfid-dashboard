package models

import "time"

type Siswa struct {
	ID           uint   `gorm:"primaryKey"`
	NIS          string `gorm:"unique;not null"`
	Nama         string `gorm:"size:100;not null"`
	JenisKelamin string `gorm:"size:1"` // L / P
	TempatLahir  string
	TanggalLahir time.Time
	Alamat       string
	NamaWali     string
	NoHP         string
	Status       string `gorm:"default:aktif"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	//RELASI
	KelasID uint
	Kelas   Kelas `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	JurusanID uint
	Jurusan   Jurusan `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	// RELASI
	Kartu *Kartu `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type Kelas struct {
	ID   uint   `gorm:"primaryKey"`
	Nama string `gorm:"type:varchar(50);unique;not null"`
}

type Jurusan struct {
	ID   uint   `gorm:"primaryKey"`
	Nama string `gorm:"type:varchar(50);unique;not null"`
}
