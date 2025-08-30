package data

import (
	"time"

	"gorm.io/gorm"
)

type CodigoLugar struct {
	ID           string         `gorm:"primaryKey;default:uuid_generate_v4()" json:"-"`
	CreatedAt    time.Time      `gorm:"default:now()" json:"-"`
	UpdatedAt    time.Time      `gorm:"default:now()" json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	CodigoLugar  string         `gorm:"" json:"codigo_lugar"`
	Entidad      string         `json:"entidad"`
	ConvenioBono string         `json:"convenio_bono"`
}

func (db DB) CreateCodigoLugar(codigoLugar *CodigoLugar) error {
	result := db.Create(&codigoLugar)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (db DB) DeleteCodigoLugar() error {
	result := db.Exec("DELETE FROM codigo_lugars")
	return result.Error
}

func (db DB) UpdateCodigoLugar(codigo string, resp *CodigoLugar) (*CodigoLugar, error) {
	db.Model(&resp).Where("codigo_lugar = ?", codigo).Updates(resp)
	resps, _ := db.GetCodigoLugar(codigo)
	return resps, nil
}

func (db DB) GetCodigoLugar(codigo string) (*CodigoLugar, error) {
	var resp *CodigoLugar
	result := db.Where("codigo_lugar = ?", codigo).Find(&resp)
	if result.Error != nil {
		return resp, result.Error
	}
	return resp, nil
}

func (db DB) CodigoLugarExists(codigo string) bool {
	response := db.Model(&CodigoLugar{}).
		Where("codigo_lugar = ?", codigo).
		First(&CodigoLugar{})
	return response.Error == nil
}

func (db DB) GetCodigoById(dni string) (CodigoLugar []*CodigoLugar, err error) {
	result := db.Table("codigo_lugars").Select("DISTINCT codigo_lugar").Where("entidad = ?", dni).Order("codigo_lugar DESC").Find(&CodigoLugar)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) GetAllCodigoById() (CodigoLugar []*CodigoLugar, err error) {
	result := db.Find(&CodigoLugar)
	if result.Error != nil {
		err = result.Error
	}
	return
}
