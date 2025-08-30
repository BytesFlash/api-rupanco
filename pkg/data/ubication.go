package data

import (
	"time"

	"github.com/imedcl/manager-api/pkg/config"
)

type Ubication struct {
	ID            string       `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	Code          string       `gorm:"uniqueIndex:idx_code_institution" json:"code"`
	Name          string       `json:"name"`
	Description   string       `gorm:"type:varchar(200);uniqueIndex;not null" json:"description"`
	Entitity      string       `json:"entitity"`
	Address       string       `json:"address"`
	State         string       `json:"state"`
	InstitutionID string       `gorm:"uniqueIndex:idx_code_institution" json:"-"`
	Institution   *Institution `json:"institution"`
	CreatedAt     time.Time    `gorm:"default:now()"`
	UpdatedAt     time.Time    `gorm:"default:now()"`
}

func (db DB) CreateUbication(ubication *Ubication) bool {
	result := db.Create(ubication)
	return result.Error == nil
}

func (db DB) UpdateUbication(code string, ubication *Ubication, institution *Institution) bool {
	result := db.Model(&ubication).Where("Code = ? AND institution_id = ?", code, institution.ID).Updates(ubication)
	return result.Error == nil
}

func (db DB) RemoveUbication(code string, institution *Institution) (bool, error) {
	var ubication Ubication
	db.Where("code = ? AND institution_id = ?",
		code,
		institution.ID,
	).First(&ubication)
	db.Delete(&ubication)
	return true, nil
}

func (db DB) ExistsUbication(code string, institution *Institution) bool {
	var ubication Ubication
	result := db.Where("code = ? AND institution_id = ?",
		code,
		institution.ID,
	).First(&ubication)
	return result.Error == nil
}

func (db DB) GetUbications(institution *Institution) *[]Ubication {
	var ubications []Ubication
	var ubication Ubication
	code := config.DEFAULT_UBICATION
	if !db.ExistsUbication(code, institution) {
		ubication.Code = code
		ubication.Institution = institution
		ubication.Name = code
		db.CreateUbication(&ubication)
	}
	_ = db.Where("institution_id = ?", institution.ID).Preload("Institution").Find(&ubications)

	return &ubications
}

func (db DB) GetUbication(code string, institution *Institution) (ubication *Ubication, err error) {
	result := db.Where("code = ? AND institution_id = ?", code, institution.ID).Preload("Institution").First(&ubication)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) GetDefaultUbication(institution *Institution) *Ubication {
	var ubication Ubication
	code := config.DEFAULT_UBICATION
	if !db.ExistsUbication(code, institution) {
		ubication.Code = code
		ubication.Institution = institution
		ubication.Name = code
		db.CreateUbication(&ubication)
	}
	_ = db.Where("code = ? AND institution_id = ?", code, institution.ID).Preload("Institution").First(&ubication)

	return &ubication
}

func (db DB) CreateNewUbication(location string, institution *Institution) *Ubication {
	var ubication Ubication
	code := location
	if !db.ExistsUbication(code, institution) {
		ubication.Code = code
		ubication.Institution = institution
		ubication.Name = code
		db.CreateUbication(&ubication)
	}
	_ = db.Where("code = ? AND institution_id = ?", code, institution.ID).Preload("Institution").First(&ubication)

	return &ubication
}
