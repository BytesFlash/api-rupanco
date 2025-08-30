package data

import (
	"time"

	"gorm.io/gorm"
)

type AutentiaService struct {
	ID        string              `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time           `gorm:"default:now()" json:"-"`
	UpdatedAt time.Time           `gorm:"default:now()" json:"-"`
	DeletedAt gorm.DeletedAt      `gorm:"index" json:"-"`
	Name      string              `json:"name" gorm:"uniqueIndex;not null"`
	Resources []*AutentiaResource `gorm:"foreignKey:ServiceID" json:"resources"`
}

type AutentiaResource struct {
	ID        string         `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time      `gorm:"default:now()" json:"-"`
	UpdatedAt time.Time      `gorm:"default:now()" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	ServiceID string         `json:"servicio_id" gorm:"not null;uniqueIndex:idx_service_resource"`
	Name      string         `json:"name" gorm:"not null;uniqueIndex:idx_service_resource"`
}

//Resource Autentia

func (db DB) CreateAutentiaService(resource *AutentiaService) (*AutentiaService, error) {
	result := db.Create(resource)
	if result.Error != nil {
		return nil, result.Error
	}
	return resource, nil
}

func (db DB) GetAutentiaServiceByName(name string) (service *AutentiaService, err error) {
	result := db.Where("name = ?", name).First(&service)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db *DB) CreateAutentiaResource(resource *AutentiaResource) error {
	return db.Create(resource).Error
}

func (db DB) GetAutentiaResourceByNameAndID(name string, servideId string) (resource *AutentiaResource, err error) {
	result := db.Where("name = ? and service_id = ?", name, servideId).First(&resource)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) GetAllService() (service []*AutentiaService, err error) {
	result := db.Find(&service)
	if result.Error != nil {
		return service, result.Error
	}
	return
}

func (db DB) GetAutentiaResourceByName(servideId string) (resource []*AutentiaResource, err error) {
	result := db.Where("service_id = ?", servideId).Find(&resource)
	if result.Error != nil {
		return resource, result.Error
	}
	return
}
