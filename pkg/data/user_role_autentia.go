package data

import (
	"time"

	"gorm.io/gorm"
)

type UserRoleAutentia struct {
	ID        string         `gorm:"primaryKey;default:uuid_generate_v4()" json:"-"`
	CreatedAt time.Time      `gorm:"default:now()" json:"-"`
	UpdatedAt time.Time      `gorm:"default:now()" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    string         `gorm:"" json:"-"`
	RoleName  string         `json:"role_name"`
	RoleId    string         `json:"role_id"`
	User      *User          `json:"-"`
}

func (db DB) CreateUserRolAutentia(userRoleAutentia *UserRoleAutentia) error {
	result := db.Create(&userRoleAutentia)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (db DB) GetUserRoleAutentia(id string) ([]*UserRoleAutentia, error) {
	var user []*UserRoleAutentia
	result := db.Where("user_id = ?", id).Find(&user)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}

func (db DB) GetRoleAutentiabyId(UserId string) (*UserRoleAutentia, error) {
	var role *UserRoleAutentia
	result := db.Where("user_id = ?", UserId).First(&role)
	if result.Error != nil {
		return role, result.Error
	}
	return role, nil
}

func (db DB) DeleteAllUserRolAutentia(UserId string) ([]*UserRoleAutentia, error) {
	var user []*UserRoleAutentia
	db.Where("user_id = ? AND deleted_at IS NULL", UserId).Find(&user)
	result := db.Unscoped().Delete(&user)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}
