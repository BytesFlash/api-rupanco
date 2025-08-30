package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/imedcl/manager-api/pkg/config"
)

type User struct {
	ID                  string                 `gorm:"primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatedAt           time.Time              `gorm:"default:now()" json:"-"`
	UpdatedAt           time.Time              `gorm:"default:now()" json:"-"`
	DeletedAt           gorm.DeletedAt         `gorm:"index" json:"-"`
	NickName            string                 `gorm:"type:varchar(100);uniqueIndex;not null;default:uuid_generate_v4()" json:"nick_name"`
	Name                string                 `json:"name"`
	Dni                 string                 `json:"dni"`
	Email               string                 `json:"email"`
	Password            string                 `json:"-"`
	Token               string                 `gorm:"default:uuid_generate_v4()" json:"-"`
	ExpiresAt           time.Time              `gorm:"default:now()" json:"-"`
	Validated           bool                   `json:"validated"`
	StatusUser          string                 `gorm:"default:Activo" json:"status_user"`
	ActiInst            bool                   `gorm:"default:false;not null" json:"acti_inst"`
	Active              bool                   `gorm:"default:true;not null" json:"active"`
	Picture             string                 `json:"picture"`
	Description         string                 `json:"-"`
	DniEntity           string                 `json:"dni_entity"`
	LastTokenSession    string                 `json:"-"`
	CountryID           string                 `gorm:"index:idx_user_country" json:"-"`
	Country             *Country               `json:"country"`
	UserRolInstitutions []*UserRoleInstitution `json:"-"`
	UserRolesAutentia   []*UserRoleAutentia    `gorm:"foreignKey:UserID" json:"user_roles_autentia"`
}

type Users []User

const graphAPIEndpoint = "https://graph.microsoft.com/v1.0/me"

// GraphUser contiene los datos del usuario de Microsoft Graph
type GraphUser struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"mail"`
}

type Claims struct {
	User User `json:"user"`
	jwt.StandardClaims
}

const sessionDuration = 15

func (db DB) CreateUser(user *User) {
	user.Password, _ = HashPassword(user.Password)
	_ = db.Create(&user)
}

func (db DB) DeleteUser(user *User) error {
	result := db.Unscoped().Delete(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (db DB) Login(user *User, password string) (string, error) {
	cfg := config.New()
	if !user.Validated || user.StatusUser != config.STATUS_ACTIVE {
		return "", errors.New("Usuario inactivo")
	}

	mySigningKey := []byte(cfg.SignKey())
	if checkPasswordHash(password, user.Password) {
		//roles, _ := db.GetRoles(user.ID)
		//user.Roles = roles
		claims := &Claims{
			User: *user,
			StandardClaims: jwt.StandardClaims{
				Issuer: "Autentia Admin",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		var signed string
		signed, err := token.SignedString(mySigningKey)
		if err != nil {
			return "", err
		}
		user.ExpiresAt = time.Now().Add(time.Minute * sessionDuration)
		user.LastTokenSession = signed
		_ = db.Model(&user).Where("nick_name = ?", user.NickName).Updates(user)
		return signed, nil
	}
	return "", errors.New("Usuario - Contraseña no es una combinación válida")
}

/* Start Microsoft */
func (db DB) LoginMicrosoft(user *User) (string, error) {
	cfg := config.New()

	if !user.Validated || user.StatusUser != config.STATUS_ACTIVE {
		return "", errors.New("Usuario inactivo")
	}

	mySigningKey := []byte(cfg.SignKey())
	//roles, _ := db.GetRoles(user.ID)
	//user.Roles = roles
	claims := &Claims{
		User: *user,
		StandardClaims: jwt.StandardClaims{
			Issuer: "Autentia Admin",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	var signed string
	signed, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", err
	}
	user.ExpiresAt = time.Now().Add(time.Minute * sessionDuration)
	user.LastTokenSession = signed
	_ = db.Model(&user).Where("nick_name = ?", user.NickName).Updates(user)
	return signed, nil
	/*
		return "", errors.New("Usuario - Contraseña no es una combinación válida") */
}
func FetchGraphUserData(Barer string) (*GraphUser, error) {
	req, err := http.NewRequest("GET", graphAPIEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creando la solicitud: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", Barer))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error haciendo la solicitud: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("solicitud fallida: %s", resp.Status)
	}

	var user GraphUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("error decodificando la respuesta: %v", err)
	}

	return &user, nil
}

/* Finish Microsoft */

func (db DB) ExtendToken(user *User) {
	user.ExpiresAt = time.Now().Add(time.Minute * sessionDuration)
	_ = db.Model(&user).Where("nick_name = ?", user.NickName).Updates(user)
}

func (db DB) GetRoleByUser(nickName string) (userRoleInstitution *User, err error) {
	result := db.Debug().Where("users.nick_name = ?", nickName).
		Preload("UserRolInstitutions.Institution").
		Preload("UserRolInstitutions.Role.Modules").
		Preload("Country").
		Find(&userRoleInstitution)
	if result.Error != nil {
		err = result.Error
	}

	return
}

func (db DB) UserExists(nickName string) bool {
	var user User
	result := db.Where("nick_name = ?", nickName).First(&user)
	return result.RowsAffected > 0
}

func (db DB) UserExistsEmail(email string) bool {
	var user User
	result := db.Where("email = ?", email).First(&user)
	return result.RowsAffected > 0
}
func (db DB) UserExistsName(nickName string) (user *User, err error) {
	result := db.Where("nick_name = ?", nickName).Preload("Country").First(&user)
	if result.Error != nil {
		return user, result.Error
	}

	return
}
func (db DB) UserExistEmail(email string) (user *User, err error) {
	result := db.Where("email = ?", email).Preload("Country").First(&user)
	if result.Error != nil {
		return user, result.Error
	}

	return
}

func (db DB) ExpireToken(user *User) {
	user.ExpiresAt = time.Now().Add(time.Minute * -1)
	_ = db.Model(&user).Where("nick_name = ?", user.NickName).Updates(user)
}

func (db DB) UserAllDBExists(nickName string) bool {
	var user User
	result := db.Unscoped().Where("nick_name = ?", nickName).First(&user)
	return result.RowsAffected > 0
}

func (db DB) GetUser(nickName string) (*User, error) {
	var user *User
	result := db.Where("nick_name = ?", nickName).Preload("Country").First(&user)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}
func (db DB) GetUserEmail(email string) (*User, error) {
	var user *User
	result := db.Where("email = ?", email).Preload("Country").First(&user)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}

func (db DB) GetUserNickName(nickName string) (*string, error) {
	var name string
	result := db.Model(&User{}).Where("nick_name = ?", nickName).Select("name").Scan(&name)
	if result.Error != nil {
		return nil, result.Error
	}
	return &name, nil
}

func (db DB) GetUserReccovery(email string) (*User, error) {
	var user *User
	result := db.Where("email = ? ", email).Preload("Country").First(&user)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}

func (db DB) GetUserByID(id string) (*User, error) {
	var user *User
	result := db.Where("ID = ?", id).First(&user)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}

func (db DB) IsActive(id string) bool {
	var user User
	result := db.Where("ID = ?", id).First(&user)
	if result.Error != nil {
		return false
	}
	return user.StatusUser == config.STATUS_ACTIVE && user.ExpiresAt.Unix() >= time.Now().Unix()
}

func (db DB) GetUserByToken(token string) (*User, error) {
	var user *User
	result := db.Where("Token = ?", token).First(&user)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}

func (db DB) UpdateUser(nickName string, user *User, password string) (*User, error) {
	if nickName != "" {
		if password != "" {
			user.Password, _ = HashPassword(password)
		}
		if user.Description == "" {
			_ = db.Model(&user).Where("nick_name = ?", nickName).Update("description", user.Description)
			_ = db.Model(&user).Where("nick_name = ?", nickName).Update("acti_inst", user.ActiInst)
			_ = db.Model(&user).Where("nick_name = ?", nickName).Update("dni_entity", user.DniEntity)

		}
		_ = db.Model(&user).Where("nick_name = ?", nickName).Updates(user)
		if user.NickName != "" {
			return db.GetUser(user.NickName)
		} else {
			return db.GetUser(nickName)
		}

	} else {
		return &User{}, errors.New("user not found")
	}
}

func (db DB) GetUsers() *Users {
	var users *Users
	_ = db.Find(&users)

	return users
}

func (db DB) GetUsersByCountry(countryId string) []*User {
	var users []*User
	_ = db.Where("country_id = ?", countryId).
		Preload("Country").
		Find(&users)

	return users
}

func (db DB) GetUsersByCountryAndEmail(countryId string, email string) []*User {
	emailParts := strings.Split(email, "@")
	if len(emailParts) < 2 {
		return nil
	}
	domain := emailParts[1]

	var users []*User
	result := db.Where("country_id = ? AND email LIKE ?", countryId, "%@"+domain).
		Preload("Country").
		Find(&users)

	if result.Error != nil {
		fmt.Printf("Error en la consulta: %v\n", result.Error)
	} else {
		fmt.Printf("Usuarios encontrados: %d\n", result.RowsAffected)
	}

	return users
}

func (db DB) GetUsersByRole(role string) []*User {
	admin, _ := db.GetRoleByName(role)
	var users []*User

	err := db.
		Table("users").
		Joins("JOIN user_role_institutions ON user_role_institutions.user_id = users.id").
		Joins("JOIN roles ON user_role_institutions.role_id = roles.id").
		Where("roles.id = ?", admin.ID).
		Preload("Country").
		Find(&users).Error

	if err != nil {
		log.Println("Error al obtener usuarios por rol:", err)
	}

	return users
}

func (db DB) UserHasRole(role string, userId string) (user *User, err error) {
	admin, _ := db.GetRoleByName(role)
	result := db.Joins("JOIN user_role_institutions ON user_role_institutions.user_id = users.id").
		Where("user_role_institutions.role_id = ?", admin.ID).
		Where("users.id = ?", userId).
		Preload("Country").
		First(&user)

	if result.Error != nil {
		err = result.Error
	}
	return
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (db DB) CreateDefaultUser(idCountry string, nick string, name string, email string, pass string) (*User, error) {
	var user *User
	user, err := db.GetUser(nick)
	if err != nil {
		if err.Error() == "record not found" {
			user = &User{
				NickName:  nick,
				Name:      name,
				CountryID: idCountry,
				Email:     email,
				Password:  pass,
				Validated: true,
			}
			db.CreateUser(user)
			logrus.Printf("user: %s, created!", nick)
			return user, nil
		} else {
			return nil, err
		}
	}

	if user.NickName == nick {
		logrus.Printf("user: %s, exists!", nick)
	}

	return user, err
}

func (db DB) CreateUserRoleInstitution(uri *UserRoleInstitution) {
	_ = db.Create(&uri)
}

func (db DB) GetUserRoleInstitution(idUser string, idRole string, idInstitution string) (uri *UserRoleInstitution, err error) {
	result := db.Where("user_id = ? AND role_id = ? AND institution_id = ?", idUser, idRole, idInstitution).First(&uri)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) UserRoleInstitutionExist(idUser string) (uri *UserRoleInstitution, err error) {
	result := db.Where("user_id = ?", idUser).First(&uri)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) UserRoleInstitutionExistByRole(idRole string) (uri *UserRoleInstitution, err error) {
	result := db.Where("role_id = ?", idRole).First(&uri)
	if result.Error != nil {
		err = result.Error
	}
	return
}

func (db DB) CreateDefaultUserRoleInstitution(idUser string, idRole string, idInstitution string) (*UserRoleInstitution, error) {
	var uri *UserRoleInstitution
	uri, err := db.GetUserRoleInstitution(idUser, idRole, idInstitution)
	if err != nil {
		if err.Error() == "record not found" {
			uri = &UserRoleInstitution{
				UserID:        idUser,
				RoleID:        idRole,
				InstitutionID: idInstitution,
			}
			db.CreateUserRoleInstitution(uri)
			logrus.Print("uri: created!")
			return uri, nil
		} else {
			return nil, err
		}
	}

	if uri.UserID == idUser && uri.RoleID == idRole && uri.InstitutionID == idInstitution {
		logrus.Print("uri: exists!")
	}

	return uri, err
}
