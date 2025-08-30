package data

import (
	"time"

	"github.com/imedcl/manager-api/pkg/config"
	"github.com/imedcl/manager-api/pkg/services"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Institution struct {
	ID                   string         `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt            time.Time      `gorm:"default:now()" json:"-"`
	UpdatedAt            time.Time      `gorm:"default:now()" json:"-"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
	Description          string         `json:"description"`
	CountryID            string         `gorm:"index:idx_institution_country" json:"-"`
	Active               bool           `gorm:"default:true;not null" json:"active"`
	Dni                  string         `gorm:"index" json:"dni"`
	Name                 string         `gorm:"varchar(100);uniqueIndex:idx_institution_country;not null" json:"name"`
	Email                string         `json:"email"`
	FlagDec              int            `json:"flag_dec"`
	Nemo                 string         `json:"nemo"`
	State                int            `json:"state"`
	FormatPdf            string         `json:"format_pdf"`
	NameInst             string         `json:"name_inst"`
	OwnerID              string         `json:"-"`
	Owner                *Owner         `json:"owner"`
	Country              *Country
	UserRolInstitutiones []UserRoleInstitution `json:"-"`
}

func (db DB) CreateInstitution(institution *Institution) {
	services.CreateInstitution(
		institution.Country.Name,
		institution.Name,
		institution.Nemo,
		institution.Email,
		institution.State,
		institution.Description,
		institution.FlagDec,
		institution.Dni,
	)
	db.Create(&institution)
	db.AddConfigOwner(institution)
}

func (db DB) CreateLocalInstitution(institution *Institution) {
	db.Create(&institution)
	db.AddConfigOwner(institution)
}

func (db DB) AddConfigOwner(institution *Institution) {
	owner := config.DEFAULT_OWNER
	if institution.Owner.Name == owner {
		if config.Contains(config.ACEPTA, institution.Name) {
			owner = "ACEPTA"
		}
		if config.Contains(config.IMED, institution.Name) {
			if owner == config.DEFAULT_OWNER {
				owner = "I-MED"
			} else {
				owner = "TODOS"
			}
		}
		ownerObject, _ := db.GetOwner(owner)
		institution.Owner = ownerObject
		db.Model(&institution).Where("ID = ?", institution.ID).Updates(institution)
	}
}

func (db DB) UpdateInstitution(institution *Institution) {
	if db.InstitutionExists(institution.Name, institution.Country.ID) {
		instit, _ := db.GetInstitution(institution.Name, institution.Country.ID)
		email := institution.Email
		if email == "" {
			email = instit.Email
			institution.Email = email
		}
		services.UpdateInstitution(
			institution.Country.Name,
			institution.Name,
			instit.Nemo,
			email,
			institution.State,
			institution.Description,
			institution.FlagDec,
			instit.Dni,
		)
		if institution.State == 0 {
			db.Model(&institution).
				Where("Name = ? AND country_id = ?", institution.Name, institution.Country.ID).
				Update("state", institution.State)
		}
		if institution.FlagDec == 0 {
			db.Model(&institution).
				Where("Name = ? AND country_id = ?", institution.Name, institution.Country.ID).
				Update("flag_dec", institution.FlagDec)
		}
		db.Model(&institution).
			Where("Name = ? AND country_id = ?", institution.Name, institution.Country.ID).
			Updates(institution)
	}
}

func (db DB) InstitutionExists(name string, country_id string) bool {
	response := db.Model(&Institution{}).
		Where("Name = ? AND country_id = ?", name, country_id).
		First(&Institution{})
	return response.Error == nil
}

func (db DB) GetInstitution(name string, countryId string) (*Institution, error) {
	var institution *Institution
	result := db.Where("Name = ? AND country_id = ?", name, countryId).
		Preload("Country").
		Preload("Owner").
		First(&institution)
	return institution, result.Error
}

func (db DB) InstitutionsExistsId(institutionid string) (*Institution, error) {
	var institution *Institution
	result := db.Where("id = ?", institutionid).First(&institution)
	if result.Error != nil {
		return institution, result.Error
	}
	return institution, nil
}

func (db DB) InstitutionsExistsNemo(nemo string) (*Institution, error) {
	var institution *Institution
	result := db.Where("nemo = ?", nemo).First(&institution)
	if result.Error != nil {
		return institution, result.Error
	}
	return institution, nil
}

func (db DB) GetInstitutionByOwner(name string, country string, ownerId string) (*Institution, error) {
	var institution *Institution
	result := db.Where("Name = ? AND Country = ? AND owner_id = ?", name, country, ownerId).First(&institution)
	return institution, result.Error
}

func (db DB) ListAllInstitutions() (institutions []*Institution, err error) {
	result := db.
		Where("name NOT LIKE '*.%'").
		Preload("Country").
		Preload("Owner").
		Find(&institutions)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) GetAllInstitutions() (institutions []*Institution, err error) {
	result := db.
		Preload("Country").
		Preload("Owner").
		Find(&institutions)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) ListAllInstitutionsByCountry(id string) (institutions []*Institution, err error) {
	result := db.Where("country_id = ?", id).
		Where("name NOT LIKE '*.%'").
		Preload("Country").
		Preload("Owner").
		Find(&institutions)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) ListInstitution(user *User, country string) ([]*Institution, error) {
	var institutions []*Institution
	var result *gorm.DB

	result = db.
		Where("user_id = ?", user.ID).
		Where("name NOT LIKE '*.%'").
		Joins("RIGHT JOIN roles ON roles.institution_id = institutions.id").
		Where("institutions.country = ?", country).
		Where("roles.deleted_at IS NULL").
		Preload("Owner").
		Find(&institutions)
	if result.Error != nil {
		return institutions, result.Error
	}
	return institutions, nil
}

func (db DB) ListUserInstitution(userID string, country string) (*[]Role, error) {
	var roles *[]Role
	result := db.Where("user_id = ?", userID).
		Where("name NOT LIKE '*.%'").
		Joins("RIGHT JOIN institutions ON institutions.id = roles.institution_id").
		Where("institutions.country = ?", country).
		Preload("Institution").
		Find(&roles)
	if result.Error != nil {
		return roles, result.Error
	}
	return roles, nil
}

func (db DB) GetCountries() ([]string, error) {
	var institution = Institution{}
	var countries []string
	result := db.Model(institution).Distinct().Pluck("Country", &countries)
	if result.Error != nil {
		return []string{}, result.Error
	}
	return countries, nil
}

func (db DB) CreateInstitutionDefault(institution *Institution) {
	_ = db.Create(&institution)
}

func (db DB) GetInstitutionByName(name string) (institution *Institution, err error) {
	result := db.Where("name = ?", name).First(&institution)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) CreateDefaultInstitution(idCountry string, name string) (*Institution, error) {
	owner, _ := db.GetOwner("TODOS")
	var institution *Institution
	institution, err := db.GetInstitutionByName(name)
	if err != nil {
		if err.Error() == "record not found" {
			institution = &Institution{
				Description: name,
				Name:        name,
				CountryID:   idCountry,
				OwnerID:     owner.ID,
			}
			db.CreateInstitutionDefault(institution)
			logrus.Printf("institution: %s, created!", name)
			return institution, nil
		} else {
			return nil, err
		}
	}

	if institution.Name == name && institution.CountryID == idCountry {
		logrus.Printf("institution: %s, exists!", name)
	}

	return institution, err
}
func (db DB) DeleteInstitution(institution *Institution) error {
	result := db.Delete(&institution)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
