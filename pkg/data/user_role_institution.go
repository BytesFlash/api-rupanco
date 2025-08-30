package data

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type UserRoleInstitution struct {
	ID            string         `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt     time.Time      `gorm:"default:now()" json:"-"`
	UpdatedAt     time.Time      `gorm:"default:now()" json:"-"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	UserID        string         `gorm:"uniqueIndex:idx_user_role_institution" json:"-"`
	RoleID        string         `gorm:"uniqueIndex:idx_user_role_institution" json:"-"`
	InstitutionID string         `gorm:"uniqueIndex:idx_user_role_institution" json:"-"`
	Role          *Role          `json:"role"`
	Institution   *Institution   `json:"institution"`
	User          *User          `json:"-"`
}

func (db DB) CreateUserRolInstitution(userRolInstitution *UserRoleInstitution) error {
	result := db.Create(&userRolInstitution)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (db DB) GetRoleInstbyId(RoleID string) (*UserRoleInstitution, error) {
	var role *UserRoleInstitution
	result := db.Where("role_id = ?", RoleID).First(&role)
	if result.Error != nil {
		return role, result.Error
	}
	return role, nil
}

func (db DB) GetRoleInstbyUser(userID string) (uris []*UserRoleInstitution, err error) {
	result := db.Where("user_id = ?", userID).
		Preload("User").
		Preload("Role").
		Preload("Institution").
		Find(&uris)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) DeleteUserRolInstitution(user *User, institution *Institution, role *Role) error {
	var urRoleInst UserRoleInstitution
	db.Where("user_id = ? AND role_id = ? AND institution_id = ? AND deleted_at IS NULL",
		user.ID,
		role.ID,
		institution.ID,
	).First(&urRoleInst)
	result := db.Unscoped().Delete(&urRoleInst)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (db DB) DeleteUserRolInstitutionId(userID string, institutionID string, roleID string) error {
	var urRoleInst UserRoleInstitution

	logrus.Infof("Intentando buscar UserRoleInstitution: UserID=%s, RoleID=%s, InstitutionID=%s", userID, roleID, institutionID)

	// Intentar encontrar el registro
	err := db.Where("user_id = ? AND role_id = ? AND institution_id = ? AND deleted_at IS NULL",
		userID,
		roleID,
		institutionID,
	).First(&urRoleInst).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Warnf("No se encontró el registro para eliminar: UserID=%s, RoleID=%s, InstitutionID=%s", userID, roleID, institutionID)
			return nil // No es crítico si no se encuentra
		}
		logrus.Errorf("Error buscando el registro para eliminar: %v", err)
		return err
	}

	logrus.Infof("Registro encontrado: %+v", urRoleInst)

	// Eliminar el registro encontrado
	result := db.Unscoped().Delete(&urRoleInst)
	if result.Error != nil {
		logrus.Errorf("Error eliminando UserRoleInstitution: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logrus.Warnf("No se eliminó ningún registro: UserID=%s, RoleID=%s, InstitutionID=%s", userID, roleID, institutionID)
	} else {
		logrus.Infof("Eliminación exitosa: UserID=%s, RoleID=%s, InstitutionID=%s", userID, roleID, institutionID)
	}

	return nil
}

func (db DB) DeleteAllUserRolInstitution(user *User) error {
	var urRoleInst []*UserRoleInstitution
	db.Where("user_id = ? AND deleted_at IS NULL", user.ID).Find(&urRoleInst)
	result := db.Unscoped().Delete(&urRoleInst)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (db DB) GetDistinctUserRoles(userId string) (uris []*UserRoleInstitution, err error) {
	result := db.Where("user_id = ?", userId).
		Preload("Role").
		Preload("Role.Modules").
		Preload("Institution").
		Distinct("role_id").
		Find(&uris)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) GetUserRolesInstitutionsByRole(roleId string, userId string) (uris []*UserRoleInstitution, err error) {
	result := db.Where("role_id = ? AND user_id = ?", roleId, userId).
		Preload("Institution").
		Preload("Institution.Country").
		Preload("Institution.Owner").
		Distinct("institution_id").
		Find(&uris)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) GetUserAndInstitution(userId string, institutionId string) (uris *UserRoleInstitution, err error) {
	result := db.Where("user_id = ? AND institution_id = ?", userId, institutionId).
		First(&uris)
	if result.Error != nil {
		err = result.Error
	}
	return
}
