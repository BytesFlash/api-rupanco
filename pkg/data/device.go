package data

import (
	"time"

	"gorm.io/gorm"
)

type Device struct {
	CreatedAt time.Time              `gorm:"primaryKey"`
	UpdatedAt time.Time              `gorm:"default:now()"`
	DeletedAt gorm.DeletedAt         `gorm:"index"`
	Name      string                 `json:"name"`
	Data      map[string]interface{} `gorm:"serializer:json"`
}

type Devices []Device

func (db DB) CreateDevice(device *Device) {
	_ = db.Create(&device)
}

func (db DB) GetAllDevice() (device []*Device, err error) {
	result := db.Find(&device)
	if result.Error != nil {
		return device, result.Error
	}
	return
}
