package data

import "github.com/sirupsen/logrus"

type Model struct {
	ID      string `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	Name    string `json:"name"`
	BrandID string `json:"brand_id"`
	Brand   Brand  `gorm:"foreignKey:BrandID"`
}

// Definimos la estructura Marca, una Marca tiene muchos modelos
type Brand struct {
	ID    string  `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	Name  string  `json:"name"`
	Model []Model `gorm:"foreignKey:BrandID"`
}

func (db DB) CreateBrand(brand *Brand) {
	_ = db.Create(&brand)
}

func (db DB) ListAllBrand() ([]*Brand, error) {
	var brands []*Brand
	result := db.Find(&brands)
	if result.Error != nil {
		return nil, result.Error
	}
	return brands, nil
}

func (db DB) UpdateBrand(brand *Brand) error {
	result := db.Model(&Brand{}).Where("id = ?", brand.ID).Updates(brand)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		logrus.Warn("No se actualizó ninguna marca")
	}
	return nil

}

func (db DB) ListModelsByBrand(brandId string) ([]*Model, error) {
	var models []*Model
	result := db.Where("brand_id = ?", brandId).Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}
	return models, nil
}

func (db DB) GetBrandByName(name string) (brand *Brand, err error) {
	result := db.Where("name = ?", name).First(&brand)
	if result.Error != nil {
		err = result.Error
	}
	return
}
func (db DB) GetBrandById(id string) (brand *Brand, err error) {
	result := db.Where("id = ?", id).First(&brand)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) CreateModel(model *Model) {
	_ = db.Create(&model)
}

func (db DB) GetModelByName(name string) (model *Model, err error) {
	result := db.Where("name = ?", name).First(&model)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) GetModelByNameBrandByID(name string, brandId string) (model *Model, err error) {
	result := db.Where("name = ? AND brand_id = ?", name, brandId).First(&model)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) CreateDefaultBrandsAndModels() {
	// Definir las marcas y modelos por defecto
	brands := []struct {
		Name  string
		Model []string
	}{
		{"Digital Persona", []string{"Uareu 4500"}},
		{"HID", []string{"Eikontc710"}},
		{"HID", []string{"Uareu 4500"}},
	}

	for _, brand := range brands {
		// Verificar si la marca ya existe
		var existingBrand Brand
		if err := db.Where("name = ?", brand.Name).First(&existingBrand).Error; err != nil {
			// Si la marca no existe, crearla
			newBrand := Brand{Name: brand.Name}
			if err := db.Create(&newBrand).Error; err != nil {
				logrus.Printf("Error al crear la marca %s: %v", brand.Name, err)
				continue
			}
			// Crear los modelos asociados a la nueva marca
			for _, modelName := range brand.Model {
				newModel := Model{Name: modelName, BrandID: newBrand.ID}
				if err := db.Create(&newModel).Error; err != nil {
					logrus.Printf("Error al crear el modelo %s para la marca %s: %v", modelName, brand.Name, err)
				}
			}
		} else {
			// Si la marca ya existe, solo crear los modelos asociados
			for _, modelName := range brand.Model {
				var existingModel Model
				// Verificar si el modelo ya existe para esta marca
				if err := db.Where("name = ? AND brand_id = ?", modelName, existingBrand.ID).First(&existingModel).Error; err != nil {
					// Si el modelo no existe, crearlo
					newModel := Model{Name: modelName, BrandID: existingBrand.ID}
					if err := db.Create(&newModel).Error; err != nil {
						logrus.Printf("Error al crear el modelo %s para la marca %s: %v", modelName, existingBrand.Name, err)
					}
				}
			}
		}
	}
}
