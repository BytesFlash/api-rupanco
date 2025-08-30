package data

import (
	"time"

	"github.com/imedcl/manager-api/pkg/config"
	"gorm.io/gorm"
)

type Owner struct {
	ID        string         `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time      `gorm:"default:now()" json:"-"`
	UpdatedAt time.Time      `gorm:"default:now()" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `gorm:"type:varchar(200);uniqueIndex;not null" json:"name"`
}

func (db DB) GetOwner(name string) (*Owner, error) {
	var owner *Owner
	result := db.Where("name = ?", name).First(&owner)
	return owner, result.Error
}

func (db DB) GetOwnerByID(id string) (*Owner, error) {
	var owner *Owner
	result := db.Where("id = ?", id).First(&owner)
	return owner, result.Error
}

func (db DB) CreateDefaultOwners() {
	for _, owner := range config.OWNERS {
		var ownerModel = &Owner{}
		result := db.Model(&Owner{}).Where("name = ?", owner).First(ownerModel)
		if result.RowsAffected == 0 {
			ownerModel = &Owner{
				Name: owner,
			}
			db.Create(ownerModel)
		}

	}
}
