package data

import (
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AutentiaRole struct {
	ID        string         `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time      `gorm:"default:now()" json:"-"`
	UpdatedAt time.Time      `gorm:"default:now()" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `json:"name" gorm:"uniqueIndex;not null"`
}

type AutentiaRoleInstitution struct {
	AutentiaRoleID string
	InstitutionID  string
	CreatedAt      time.Time     `gorm:"default:now()"`
	UpdatedAt      time.Time     `gorm:"default:now()"`
	AutentiaRole   *AutentiaRole `json:"role"`
	Institution    *Institution  `json:"institution"`
}

//Role Autentia

func (db DB) CreateAutentiaRole(role *AutentiaRole) {
	_ = db.Create(&role)
}

func (db DB) DeleteAutentiaRole(name string) error {
	var role AutentiaRole
	db.Where("name = ? AND deleted_at IS NULL", name).First(&role)
	result := db.Unscoped().Delete(&role)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (db DB) GetAllAutentiaRole() (role []*AutentiaRole, err error) {
	result := db.Find(&role)
	if result.Error != nil {
		return role, result.Error
	}
	return
}

func (db DB) GetAutentiaRoleByName(name string) (role *AutentiaRole, err error) {
	result := db.Where("name = ?", name).First(&role)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) GetAutentiaRoleByID(role string) (autentiaRole *AutentiaRole, err error) {
	result := db.Where("id = ?", role).First(&autentiaRole)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) CreateDefaultAutentiaRole(name string) (*AutentiaRole, error) {
	var role *AutentiaRole
	role, err := db.GetAutentiaRoleByName(name)
	if err != nil {
		if err.Error() == "record not found" {
			role = &AutentiaRole{
				Name: name,
			}
			db.CreateAutentiaRole(role)
			logrus.Printf("role: %s, created!", name)
			return role, nil
		} else {
			return nil, err
		}
	}

	if role.Name == name {
		logrus.Printf("role: %s, exists!", name)
	}
	return role, err
}
