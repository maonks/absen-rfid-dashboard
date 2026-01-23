package models

import "time"

type Absen struct {
	ID       uint `gorm:"primaryKey"`
	UID      string
	DeviceId string
	Waktu    time.Time `gorm:"type:timestamp"`
}

// Untuk Tap Realtime
type RealTime struct {
	UID      string  `json:"uid"`
	Nama     *string `json:"nama"`
	DeviceId string  `json:"device_id"`
	Waktu    string  `json:"waktu"`
}

// Untuk Absensi Bulanan dan Report
type HariCell struct {
	Masuk  string
	Pulang string
	Status string // OK | LATE
}

type AbsensiRow struct {
	ID   uint // ← SiswaID
	Nama string
	Hari map[int]*HariCell
}

type AbsensiBulananVM struct {
	Days []int
	Rows []AbsensiRow
}

type AbsensiStatus struct {
	ID      uint      `gorm:"primaryKey"`
	SiswaID uint      `gorm:"index;not null"`
	Tanggal time.Time `gorm:"type:date;index"`
	Status  string    `gorm:"type:varchar(10);not null"`
	// OK | LATE | IJIN | SAKIT | ALPA

	Keterangan string
	UpdatedBy  *uint
	UpdatedAt  time.Time
	CreatedAt  time.Time

	Siswa Siswa `gorm:"constraint:OnDelete:CASCADE"`
}

type AbsensiHari struct {
	ID        uint      `gorm:"primaryKey"`
	Tanggal   time.Time `gorm:"type:date;uniqueIndex"`
	Status    string    `gorm:"type:varchar(10)"` // LIBUR
	CreatedAt time.Time
	UpdatedAt time.Time
}
