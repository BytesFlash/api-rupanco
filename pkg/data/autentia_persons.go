package data

import (
	"time"

	"gorm.io/gorm"
)

type AutentiaPerson struct {
	ID                     string                    `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt              time.Time                 `gorm:"default:now()" json:"create"`
	UpdatedAt              time.Time                 `gorm:"default:now()" json:"-"`
	DeletedAt              gorm.DeletedAt            `gorm:"index" json:"-"`
	Dni                    string                    `gorm:"type:varchar(12);not null;default:uuid_generate_v4()" json:"dni"`
	Country                string                    `json:"country"`
	Name                   string                    `json:"name" `
	Names                  string                    `json:"names"`
	MiddleName             string                    `json:"middle_name"`
	LastName               string                    `json:"last_name"`
	Institution            string                    `json:"institution"`
	Gender                 string                    `json:"gender"`
	Birthdate              string                    `json:"birthdate"`
	Version                int                       `json:"version"`
	VersionChange          string                    `json:"version_change"`
	NroAudit               string                    `json:"nro_audit"`
	Description            string                    `json:"descriptions" `
	InstitutionBloq        string                    `json:"institution_bloq"`
	TypeBloq               string                    `json:"type_bloq"`
	UserID                 string                    `gorm:"index:idx_users" json:"user"`
	User                   *User                     `json:"users"`
	DocumentRegisterPerson []*DocumentRegisterPerson `json:"document"`
}

type PersonDocument struct {
	ID               string          `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt        time.Time       `gorm:"default:now()" json:"-"`
	UpdatedAt        time.Time       `gorm:"default:now()" json:"-"`
	DeletedAt        gorm.DeletedAt  `gorm:"index" json:"-"`
	Name             string          `gorm:"type:varchar(300);uniqueIndex;not null;default:uuid_generate_v4()" json:"name"`
	Uri              string          `gorm:"type:varchar(300);uniqueIndex;not null;default:uuid_generate_v4()" json:"uri"`
	AutentiaPersonId string          `gorm:"index:idx_autentia_person_id" json:"person_id"`
	AutentiaPerson   *AutentiaPerson `json:"person"`
}

// borrar DocumentRegisterPerson
type DocumentRegisterPerson struct {
	ID                       string                  `gorm:"primaryKey;default:uuid_generate_v4()" json:"-"`
	CreatedAt                time.Time               `gorm:"default:now()" json:"-"`
	UpdatedAt                time.Time               `gorm:"default:now()" json:"-"`
	DeletedAt                gorm.DeletedAt          `gorm:"index" json:"-"`
	AutentiaPersonId         string                  `gorm:"index:idx_autentia_person_id" json:"people"`
	DocumentRegisterPersonId string                  `gorm:"uniqueIndex:idx_document_register_person_id" json:"document_people"`
	DocumentRegisterPerson   *DocumentRegisterPerson `json:"document"`
	AutentiaPerson           *AutentiaPerson         `json:"person"`
}

//People

func (db DB) CreatePersonManager(person *AutentiaPerson) {
	_ = db.Create(&person)
}

func (db DB) CountPeopleVersion(dni string) (person *AutentiaPerson, err error) {
	result := db.Debug().Where("dni = ?", dni).Order("version DESC").First(&person)
	if result.Error != nil {
		return person, result.Error
	}
	return person, nil
}

func (db DB) ExistsPeopleManagerbyDni(dni string) (person *AutentiaPerson, err error) {
	result := db.Where("dni = ? ", dni).First(&person)
	if result.Error != nil {
		return person, result.Error
	}
	return person, nil
}

func (db DB) SearchPeopleManager(dni string, version string) (person []*AutentiaPerson, err error) {
	result := db.Where("dni = ? AND version = ? ", dni, version).Find(&person)
	if result.Error != nil {
		return person, result.Error
	}
	return
}

func (db DB) ExistsPeopleManagerbyDniVersion(dni string) (person []*AutentiaPerson, err error) {
	result := db.Debug().
		Where("dni = ? AND version_change = ?", dni, "Actual").Order("created_at DESC").
		Preload("User").
		Find(&person)
	if result.Error != nil {
		return person, result.Error
	}
	return
}

//Document

func (db DB) CreateDocumentRegisterPerson(document *PersonDocument) {
	_ = db.Create(&document)
}

func (db DB) GetDocumentsByPerson(dni string) (idPeople []*PersonDocument, err error) {
	result := db.Debug().Table("person_documents").
		Select("person_documents.*").
		Preload("AutentiaPerson").
		Preload("AutentiaPerson.User").
		Joins("JOIN autentia_people ON person_documents.autentia_person_id = autentia_people.id").
		Where("autentia_people.dni = ? ", dni).
		Find(&idPeople)
	if result.Error != nil {
		err = result.Error
	}
	return
}

// Get all document by person get
func (db DB) GetDocumentByPersonVersion(dni string, version string) (idPeople []*PersonDocument, err error) {
	result := db.Table("person_documents").
		Select("person_documents.*").
		Preload("AutentiaPerson").
		Preload("AutentiaPerson.User").
		Joins("JOIN autentia_people ON person_documents.autentia_person_id = autentia_people.id").
		Where("autentia_people.dni = ? AND autentia_people.version = ?", dni, version).
		Find(&idPeople)
	if result.Error != nil {
		err = result.Error
	}
	return
}

// Get all document
func (db DB) GetDocuments() (documents []*PersonDocument, err error) {
	result := db.Table("person_documents").Find(&documents)
	if result.Error != nil {
		err = result.Error
	}
	return
}

// Delete document
func (db DB) DeleteDocument(document *PersonDocument) error {
	result := db.Delete(&document)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
