package data

import (
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Country struct {
	ID        string         `gorm:"primaryKey;default:uuid_generate_v4()" json:"-"`
	CreatedAt time.Time      `gorm:"default:now()" json:"-"`
	UpdatedAt time.Time      `gorm:"default:now()" json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Name      string         `json:"name" gorm:"type:varchar(100);uniqueIndex;not null"`
	Active    bool           `gorm:"default:true;not null" json:"active"`
}

func (db DB) CreateCountry(country *Country) {
	_ = db.Create(&country)
}

func (db DB) GetCountryByName(name string) (country *Country, err error) {
	result := db.Where("name = ?", name).First(&country)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) GetCountryById(countryId string) (country *Country, err error) {
	result := db.Where("id = ?", countryId).First(&country)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) CreateDefaultCountry(name string) (*Country, error) {
	country, err := db.GetCountryByName(name)
	if err != nil {
		if err.Error() == "record not found" {
			country = &Country{
				Name: name,
			}
			db.CreateCountry(country)
			logrus.Printf("country: %s, created!", name)
			return country, nil
		} else {
			return nil, err
		}
	}

	if country.Name == name {
		logrus.Printf("country: %s, exists!", name)
	}

	return country, err
}

func (db DB) ListAllCountries() (countries []*Country, err error) {
	result := db.Where("name NOT LIKE '*.%'").Find(&countries)
	if result.Error != nil {
		err = result.Error
	}
	return
}
