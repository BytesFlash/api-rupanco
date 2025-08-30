package data

import (
	"time"

	"github.com/imedcl/manager-api/pkg/config"
)

type Sensor struct {
	ID           string `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	Code         string `gorm:"uniqueIndex:idx_sensor_code_institution" json:"code"`
	Country      string `json:"country"`
	Location     string `json:"location"`
	Institution  string `gorm:"uniqueIndex:idx_sensor_code_institution" json:"institution"`
	ExternalCode string `json:"external_code"`
	LogonType    int    `json:"logon_type"`
	Technology   string `json:"technology"`
	Brand        string `gorm:"default:''" json:"brand"`
	Model        string `gorm:"default:''" json:"model"`
	State        int    `json:"state"`
}

type EventSensor struct {
	ID          string    `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	Code        string    `gorm:"uniqueIndex:idx_sensor_code_institution" json:"code"`
	Institution string    `gorm:"uniqueIndex:idx_sensor_code_institution" json:"institution"`
	User        string    `json:"user"`
	Action      string    `json:"action"`
	Glosa       string    `json:"glosa"`
	Date        time.Time `gorm:"default:now()" json:"-"`
}

func (db DB) GetSensors(institution string, country string, limit int, offset int) []*Sensor {
	var sensors []*Sensor
	_ = db.Where("institution = ? AND country=?", institution, country).Limit(limit).
		Offset(offset).
		Preload("Ubication").
		Preload("Owner").
		Find(&sensors)
	return sensors
}

func (db DB) GetSensorsByOwner(country string, ownerName string) []*Sensor {
	var sensors []*Sensor
	bothOwner, _ := db.GetOwner(config.ALL_OWNERS)
	owner, _ := db.GetOwner(ownerName)
	_ = db.Where("sensors.country = ?", country).
		Joins("JOIN institutions ON institutions.name = sensors.institution").
		Where("sensors.owner_id = ? OR sensors.owner_id = ?", owner.ID, bothOwner.ID).
		Preload("Ubication").
		Preload("Owner").
		Find(&sensors)
	return sensors
}

func (db DB) GetSensorNumber(institution string, country string) int64 {
	var total_sensors int64
	_ = db.Model(Sensor{}).Where("institution = ? AND country=?", institution, country).Count(&total_sensors)
	return total_sensors
}

func (db DB) CreateSensor(sensor *Sensor) {
	_ = db.Create(&sensor)
}

func (db DB) CreateEventSensor(eventSensor *EventSensor) {
	_ = db.Create(&eventSensor)
}

func (db DB) UpdateSensor(code string, sensor *Sensor) bool {

	result := db.Model(&sensor).Where("Code = ?", code).Updates(sensor)
	return result.Error == nil
}

func (db DB) ExistsSensor(code string) bool {
	var sensor Sensor
	result := db.Where("code = ?", code).First(&sensor)
	return result.Error == nil
}

func (db DB) GetSensor(code string) *Sensor {
	var sensor *Sensor
	_ = db.Where("code = ?", code).
		Find(&sensor)
	return sensor
}
