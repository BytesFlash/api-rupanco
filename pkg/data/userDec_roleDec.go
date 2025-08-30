package data

import (
	"time"

	"gorm.io/gorm"
)

type UserDecRoleDec struct {
	ID                   string              `gorm:"primaryKey;default:uuid_generate_v4()" json:"-"`
	CreatedAt            time.Time           `gorm:"default:now()" json:"-"`
	UpdatedAt            time.Time           `gorm:"default:now()" json:"-"`
	DeletedAt            gorm.DeletedAt      `gorm:"index" json:"-"`
	InstitutionRoleDecID string              `gorm:"foreignKey:InstitutionRoleDecID;references:ID" json:"inst_role_dec_id"`
	UserDecID            string              `gorm:"foreignKey:UserDecID;references:ID" json:"user_dec_id"`
	InstitutionRoleDec   *InstitutionRoleDec `json:"role_institution"`
	UserDec              *UserDec            `json:"-"`
}

func (db DB) CreateUserRoleDec(userDecRoleDec *UserDecRoleDec) error {
	result := db.Create(&userDecRoleDec)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (db DB) GetUserRoleDec(id string) ([]*UserDecRoleDec, error) {
	var user []*UserDecRoleDec
	result := db.Where("user_dec_id = ?", id).Preload("InstitutionRoleDec.RoleDec").Preload("InstitutionRoleDec.Institution").Find(&user)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}

func (db DB) DeleteAllUserRolDec(id string) ([]*UserDecRoleDec, error) {
	var user []*UserDecRoleDec
	db.Where("user_dec_id = ? AND deleted_at IS NULL", id).Find(&user)
	result := db.Unscoped().Delete(&user)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}
