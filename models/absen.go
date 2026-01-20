package models

import "time"

type Absen struct {
	ID       uint `gorm:"primaryKey"`
	UID      string
	DeviceId string
	Waktu    time.Time `gorm:"type:timestamp"`
}

type HariCell struct {
	Masuk  string
	Pulang string
	Status string // OK | LATE
}

type AbsensiRow struct {
	Nama string
	Hari map[int]*HariCell // key = tanggal (1..31)
}

type AbsensiBulananVM struct {
	Days []int
	Rows []AbsensiRow
}
