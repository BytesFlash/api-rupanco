package data

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

type ConnectionAuth struct {
	Database string
	UserName string
	Password string
	Port     string
	Host     string
	SSL      string
	TimeZone string
	SSLCa    string
	SSLCert  string
	SSLKey   string
}

func (conAuth ConnectionAuth) Connect() (*DB, error) {

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s sslrootcert=%s sslcert=%s sslkey=%s",
			conAuth.Host,
			conAuth.UserName,
			conAuth.Password,
			conAuth.Database,
			conAuth.Port,
			conAuth.SSL,
			conAuth.TimeZone,
			conAuth.SSLCa,
			conAuth.SSLCert,
			conAuth.SSLKey,
		),
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func Migrate(db DB) error {
	db.SetupJoinTable(Role{}, "Modules", RoleModule{})
	db.AutoMigrate(&User{})
	db.createTableIfNotExists(Module{})
	db.createTableIfNotExists(User{})
	db.createTableIfNotExists(AutentiaPerson{})
	db.createTableIfNotExists(PersonDocument{})
	db.createTableIfNotExists(Password{})
	db.createTableIfNotExists(UserRoleInstitution{})
	db.createTableIfNotExists(Log{})
	db.createTableIfNotExists(UserDec{})
	db.createTableIfNotExists(Device{})
	db.createTableIfNotExists(RoleDec{})
	db.createTableIfNotExists(InstitutionRoleDec{})
	db.createTableIfNotExists(UserDecRoleDec{})
	db.createTableIfNotExists(UserRoleAutentia{})
	db.createTableIfNotExists(CodigoLugar{})
	db.createTableIfNotExists(TrxHtml{})
	db.createTableIfNotExists(AutentiaService{})
	db.createTableIfNotExists(AutentiaResource{})
	db.createTableIfNotExists(Brand{})
	db.createTableIfNotExists(Model{})

	return nil
}

func (db DB) createTableIfNotExists(model interface{}) {
	if !db.Migrator().HasTable(model) {
		db.Migrator().CreateTable(model)
	}
}
