package data

import (
	"time"

	"gorm.io/gorm"
)

type UserDec struct {
	ID             string            `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt      time.Time         `gorm:"default:now()" json:"-"`
	UpdatedAt      time.Time         `gorm:"default:now()" json:"-"`
	DeletedAt      gorm.DeletedAt    `gorm:"index" json:"-"`
	Rut            string            `json:"rut"`
	NumDoc         string            `json:"num_doc"`
	Name           string            `json:"name"`
	DateNac        string            `json:"date_nac"`
	Gender         string            `json:"gender"`
	Phone          string            `json:"phone"`
	Mail           string            `json:"email"`
	Status         bool              `gorm:"default:true" json:"status"`
	UserDecRoleDec []*UserDecRoleDec `gorm:"many2many:user_dec_role_decs;" json:"roles"`
}

type RoleDec struct {
	ID        string         `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time      `gorm:"default:now()" json:"-"`
	UpdatedAt time.Time      `gorm:"default:now()" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `json:"name"`
	Status    bool           `gorm:"default:true" json:"status"`
}

type InstitutionRoleDec struct {
	ID            string       `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt     time.Time    `gorm:"default:now()" json:"-"`
	UpdatedAt     time.Time    `gorm:"default:now()" json:"-"`
	RoleID        string       `json:"role_dec_id"`
	InstitutionID string       `json:"institution_id"`
	RoleDec       *RoleDec     `gorm:"foreignKey:RoleID" json:"role_dec"`
	Institution   *Institution `json:"institution"`
}

func (db DB) CreateUserDec(user *UserDec) {
	_ = db.Create(&user)
}

func (db DB) UpdateUserDec(id string, user *UserDec) (*UserDec, error) {
	db.Model(&user).Where("id = ?", id).Updates(user)
	users, _ := db.GetUserDec(id)
	return users, nil
}

func (db DB) GetAllUserDec() (user []*UserDec, err error) {
	result := db.Find(&user)
	if result.Error != nil {
		return user, result.Error
	}
	return
}

func (db DB) DeleteUserDec(id string) error {
	var userDec UserDec
	db.Where("id = ? AND deleted_at IS NULL", id).First(&userDec)
	result := db.Unscoped().Delete(&userDec)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

/*
	 func (db DB) GetUserDec(id string) (*UserDec, error) {
		var user *UserDec
		result := db.Where("id = ?", id).Find(&user)
		if result.Error != nil {
			return user, result.Error
		}
		return user, nil
	}
*/
func (db DB) GetUserDec(id string) (*UserDec, error) {
	var user *UserDec
	result := db.Where("user_decs.id = ?", id).Find(&user)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}

//role

func (db DB) IsExistRoleDec(name string) bool {
	var role RoleDec
	result := db.Where("name = ?", name).First(&role)
	return result.RowsAffected > 0
}

func (db DB) CreateRoleDec(roleDec *RoleDec) (*RoleDec, error) {
	err := db.Create(roleDec).Error
	if err != nil {
		return nil, err
	}

	return roleDec, nil
}

func (db DB) GetRoleDecByName(name string) (*RoleDec, error) {
	var role *RoleDec
	result := db.Where("name = ?", name).First(&role)
	if result.Error != nil {
		return role, result.Error
	}
	return role, nil
}
func (db DB) GetAllRoleDec() (role []*RoleDec, err error) {
	result := db.Find(&role)
	if result.Error != nil {
		return role, result.Error
	}
	return
}

//Rol Inst

func (db DB) CreateRoleInstDec(RolInstDec *InstitutionRoleDec) {
	_ = db.Create(&RolInstDec)
}

func (db DB) GetAllRoleInstDec(id string) (roleInst []*InstitutionRoleDec, err error) {
	result := db.Where("institution_id = ?", id).
		Preload("RoleDec").
		Preload("Institution").
		Find(&roleInst)

	if result.Error != nil {
		return roleInst, result.Error
	}
	return
}

func (db DB) IsExistRoleInstDec(idRole string, idInst string) bool {
	var InstDecRoleDec InstitutionRoleDec
	result := db.Where("role_id = ? AND institution_id = ?", idRole, idInst).First(&InstDecRoleDec)
	return result.RowsAffected > 0
}
