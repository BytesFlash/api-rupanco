package data

import (
	"time"

	"gorm.io/gorm"
)

type TrxHtml struct {
	ID          string         `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time      `gorm:"default:now()"`
	UpdatedAt   time.Time      `gorm:"default:now()" json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	NameHtml    string         `gorm:"type:varchar(300);uniqueIndex;not null;default:uuid_generate_v4()" json:"name_html"`
	NameTrx     string         `gorm:"type:varchar(300);uniqueIndex;not null;default:uuid_generate_v4()" json:"name_trx"`
	Uri         string         `gorm:"type:varchar(300);uniqueIndex;not null;default:uuid_generate_v4()" json:"uri"`
	Description string         `json:"description"`
	Institution string         `json:"institution"`
}

func (db DB) CreateTrxHtml(trxHtml *TrxHtml) {
	_ = db.Create(&trxHtml)
}

func (db DB) GetAllTrxHtml() (trxHtml []*TrxHtml, err error) {
	result := db.Find(&trxHtml)
	if result.Error != nil {
		return trxHtml, result.Error
	}
	return
}

func (db DB) GetNameHtmlByName(name string) (trxHtml *TrxHtml, err error) {
	result := db.Where("name_html = ?", name).First(&trxHtml)
	if result.Error != nil {
		return trxHtml, result.Error
	}
	return
}

func (db DB) GetNameTrxByName(name string) (trxHtml *TrxHtml, err error) {
	result := db.Where("name_trx = ?", name).First(&trxHtml)
	if result.Error != nil {
		return trxHtml, result.Error
	}
	return
}
func (db DB) GetFileNameByName(name string) (trxHtml *TrxHtml, err error) {
	result := db.Where("uri = ?", name).First(&trxHtml)
	if result.Error != nil {
		return trxHtml, result.Error
	}
	return
}
