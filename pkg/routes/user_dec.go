package routes

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"github.com/sirupsen/logrus"
)

type userDecStruct struct {
	Rut     string `json:"rut"`
	NumDoc  string `json:"num_doc"`
	Name    string `json:"name"`
	DateNac string `json:"date_nac"`
	Gender  string `json:"gender"`
	Phone   string `json:"phone"`
	Mail    string `json:"mail"`
}

type userDecParams struct {
	Rut      string `json:"rut"`
	NumDoc   string `json:"num_doc"`
	Name     string `json:"name"`
	DateNac  string `json:"date_nac"`
	Gender   string `json:"gender"`
	Phone    string `json:"phone"`
	Mail     string `json:"mail"`
	IdRolDec string `json:"id_rol_dec"`
}

type updateUserDecParams struct {
	Profile  string `json:"Profile"`
	Rut      string `json:"rut"`
	NumDoc   string `json:"num_doc"`
	Name     string `json:"name"`
	DateNac  string `json:"date_nac"`
	Gender   string `json:"gender"`
	Phone    string `json:"phone"`
	Mail     string `json:"mail"`
	IdRolDec string `json:"id_rol_dec"`
}

type roleDecParams struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	InstitutionID string `json:"institution_id"`
}

type InstRoleDecParams struct {
	Id string `form:"id"`
}
type userRoleDecParams struct {
	Id            string              `json:"id_role"`
	UserDecID     string              `json:"user_dec_id"`
	RoleInstDecID []InstRoleDecParams `json:"role_institution"`
}

func validateDate(dateStr string) error {
	// Intentar parsear la fecha
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("La fecha no tiene el formato correspondiente")
	}

	// Obtener la fecha actual
	now := time.Now()

	// Calcular la fecha máxima permitida (tres años antes de la fecha de hoy)
	maxDate := now.AddDate(-3, 0, 0)

	// Calcular la fecha mínima permitida (01 de enero de 1960)
	minDate := time.Date(1960, 1, 1, 0, 0, 0, 0, time.UTC)

	// Verificar si la fecha está dentro del rango permitido
	if date.Before(minDate) {
		return fmt.Errorf("La fecha no puede ser anterior al 01 de enero de 1960")
	} else if date.After(maxDate) {
		return fmt.Errorf("La fecha debe ser hace tres años o más antes de la fecha actual")
	}

	return nil
}

// @Summary userDec
// @Description create userDec
// @Tags userDec
// @security BarerToken
// @Accept json
// @Produce json
// @Param user body useDecrParams true "user"
// @Success 201 "Usuario Creado"
// @failure 400 {object} MessageResponse
// @Router /users [post]
func UserDecPostRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.POST("/usersDec", func(c *gin.Context) {

		var params userDecParams
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		rut := (params.Rut)
		num_doc := (params.NumDoc)
		name := (params.Name)
		userDateNac := (params.DateNac)
		gender := strings.ToLower(params.Gender)
		phone := (params.Phone)
		mail := (params.Mail)

		mailRegex := regexp.MustCompile(`^([a-zA-Z0-9._-]+)@[a-zA-Z0-9-]+[.]{1}[a-zA-Z]{2,4}$`)
		if len(rut) == 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Rut no es válido",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if len(num_doc) <= 8 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El número de documento debe ser correcto",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(name) <= 3 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El Nombre debe tener más de 3 carácteres",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := validateDate(userDateNac); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(gender) <= 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El formato del genero no correspondiente",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(phone) <= 7 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El formato del telefono no correspondiente",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if !mailRegex.MatchString(mail) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El email no es válido",
				Code:    http.StatusBadRequest,
			})
			return
		}
		user := &data.UserDec{
			Rut:     rut,
			NumDoc:  num_doc,
			Name:    name,
			DateNac: userDateNac,
			Gender:  gender,
			Phone:   phone,
			Mail:    mail,
		}
		db.CreateUserDec(user)
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Usuarios DEC",
			Event:        fmt.Sprintf("Se ha creado el usuario %s", name),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, MessageResponse{
			Details: ("Usuario DEC5 creado correctamente"),
			Code:    http.StatusOK,
		})
	})

}

// @Summary userDec
// @Description delete userDec
// @Tags userDec
// @security BarerToken
// @Accept json
// @Produce json
// @Param user body useDecrParams true "user"
// @Success 201 "Usuario Eliminado"
// @failure 400 {object} MessageResponse
// @Router /UserDec/delete/:id" [delete]
func UserDecDeleteRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// DELETE User Dec
	router.DELETE("/UserDec/delete/:id", func(c *gin.Context) {
		idUser := c.Param("id")
		var params userDecStruct
		usuario, erru := GetUserFromToken(c)
		if erru != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": erru.Error()})
			return
		}
		if err := c.ShouldBind(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		user, errU := db.GetUserDec(idUser)
		if errU != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Ususario DEC no existe",
				Code:    http.StatusBadRequest,
			})
			return
		}
		db.DeleteAllUserRolDec(idUser)
		err := db.DeleteUserDec(idUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al eliminar el usuario",
				Code:    http.StatusBadRequest,
			})
			return
		}
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Usuarios DEC",
			Event:        fmt.Sprintf("Se ha eliminado el usuario DEC %s", user.Name),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, MessageResponse{
			Details: ("Usuario DEC5 eliminado correctamente"),
			Code:    http.StatusOK,
		})
	})

}

// @Summary userDec
// @Description get userDec
// @Tags userDec
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} []data.userDec
// @failure 404 {object} MessageResponse
// @Router /userDec [get]
func UserDecGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//get User Dec
	router.GET("/userDec", func(c *gin.Context) {
		users, err := db.GetAllUserDec()
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Println("Error al obtener los usuarios DEC5")
			return
		}
		c.JSON(http.StatusOK, users)
	})
}

// @Summary userDec get
// @Description userDec get
// @Tags userDec
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} response
// @failure 404 {object} MessageResponse
// @Router /userDec/{id} [get]
// @param id path string true "id"
func GetUserIdDec(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get User Dec by Id
	router.GET("/userDec/:id", func(c *gin.Context) {
		idUser := c.Param("id")
		userDec, userDecErr := db.GetUserDec(idUser)

		if userDecErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: userDecErr.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		c.JSON(http.StatusOK, userDec)
	})
}

// @Summary put UserDec
// @Description put UserDec
// @Tags UserDec
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 "Usuario Actualizado correctamente"
// @failure 400 {object} MessageResponse
// @Router /UserDec/update/{id} [put]
// @Param UserDec body updateUserDecParams true "UserDec"
func UserDecPutRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Update User Dec
	router.PUT("/UserDec/update/:id", func(c *gin.Context) {
		idUser := c.Param("id")
		var params updateUserDecParams
		if err := c.ShouldBind(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		var userData, _ = db.GetUserByID(params.Profile)

		_, userDecErr := db.GetUserDec(idUser)

		if userDecErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: ("Usuario DEC5 no existe"),
				Code:    http.StatusNotFound,
			})
			return
		}

		rut := (params.Rut)
		num_doc := (params.NumDoc)
		name := (params.Name)
		userDateNac := (params.DateNac)
		gender := strings.ToLower(params.Gender)
		phone := (params.Phone)
		mail := (params.Mail)

		mailRegex := regexp.MustCompile(`^([a-zA-Z0-9._-]+)@[a-zA-Z0-9-]+[.]{1}[a-zA-Z]{2,4}$`)

		if len(rut) == 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Rut no es válido",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if len(num_doc) <= 8 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El número de documento debe ser correcto",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(name) <= 3 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El Nombre debe tener más de 3 carácteres",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := validateDate(userDateNac); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if len(gender) <= 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El formato del genero no correspondiente",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(phone) <= 7 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El formato del telefono no correspondiente",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if !mailRegex.MatchString(mail) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El email no es válido",
				Code:    http.StatusBadRequest,
			})
			return
		}

		//create sensor
		user := &data.UserDec{
			Rut:     rut,
			NumDoc:  num_doc,
			Name:    name,
			DateNac: userDateNac,
			Gender:  gender,
			Phone:   phone,
			Mail:    mail,
		}
		db.UpdateUserDec(idUser, user)

		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: userData.NickName,
			Resource:     "Sensor",
			Event:        fmt.Sprintf("El usuario DEC5 %s fue actualizado correctamente", user.Name),
			Params:       PrtyParams,
		}
		event.Write()

		var response responseSensorAddBatch
		response.Data.Status = true
		c.JSON(http.StatusOK, MessageResponse{
			Details: ("Usuario DEC5 actualizado correctamente"),
			Code:    http.StatusOK,
		})
	})

}

//dec

// @Summary RoleDec
// @Description create RoleDec
// @Tags RoleDec
// @security BarerToken
// @Accept json
// @Produce json
// @Param user body roleDecParams true "roleDec"
// @Success 201 "Rol Dec Creado"
// @failure 400 {object} MessageResponse
// @Router /RoleDec [post]
func RoleDecPostRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.POST("/RoleDec", func(c *gin.Context) {

		var params roleDecParams
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		name := (params.Name)
		institution := (params.InstitutionID)
		instName, _ := db.InstitutionsExistsId(institution)

		if len(institution) == 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Debe contener una institución",
				Code:    http.StatusBadRequest,
			})
			return
		}
		var IdRol string
		if db.IsExistRoleDec(name) {
			nameRole, _ := db.GetRoleDecByName(name)
			IdRol = nameRole.ID
			if db.IsExistRoleInstDec(IdRol, institution) {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: fmt.Sprintf("El rol %s ya se encuentra en la institución %s", name, instName.Name),
					Code:    http.StatusBadRequest,
				})
				return
			}

		} else {
			rol := &data.RoleDec{
				Name: name,
			}
			DataRol, _ := db.CreateRoleDec(rol)
			IdRol = DataRol.ID
		}

		roleInst := &data.InstitutionRoleDec{
			RoleID:        IdRol,
			InstitutionID: institution,
		}
		db.CreateRoleInstDec(roleInst)

		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Rol DEC",
			Event:        fmt.Sprintf("Se ha creado el rol %s, en la institución %s", name, instName.Name),
			Params:       PrtyParams,
		}
		event.Write()

		c.JSON(http.StatusOK, MessageResponse{
			Details: "Rol DEC5 creado con éxito",
			Code:    http.StatusOK,
		})
	})

}

// @Summary RoleDec
// @Description get RoleDec
// @Tags RoleDec
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} []data.RoleDec
// @failure 404 {object} MessageResponse
// @Router /roleDec [get]
func RoleDecGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//get User Dec
	router.GET("/roleDec/:id", func(c *gin.Context) {
		idInst := c.Param("id")
		role, err := db.GetAllRoleInstDec(idInst)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Println("error al obtener los roles")
			return
		}
		c.JSON(http.StatusOK, role)
	})
}

// @Summary UserRoleDec
// @Description create UserRoleDec
// @Tags UserRoleDec
// @security BarerToken
// @Accept json
// @Produce json
// @Param user body UserRoleDec true "UserRoleDec"
// @Success 201 "Rol asociado al usuario Dec Creado"
// @failure 400 {object} MessageResponse
// @Router /UserRoleDec [post]
func UserRoleDecPostRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.POST("/UserRoleDec", func(c *gin.Context) {

		var params userRoleDecParams
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		userId := (params.UserDecID)
		instRoleId := (params.RoleInstDecID)

		if len(userId) == 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Debe contener un usuario",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(instRoleId) == 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Debe contener un rol",
				Code:    http.StatusBadRequest,
			})
			return
		}
		db.DeleteAllUserRolDec(userId)
		for _, roleJson := range instRoleId {

			userRole := &data.UserDecRoleDec{
				UserDecID:            userId,
				InstitutionRoleDecID: roleJson.Id,
			}

			db.CreateUserRoleDec(userRole)

			PrtyParams, _ := events.PrettyParams(params)
			event := &events.EventLog{
				UserNickname: usuario.NickName,
				Resource:     "Usuario DEC5",
				Event:        fmt.Sprintf("Se ha creado el usuario %s, en el rol %s", userId, instRoleId),
				Params:       PrtyParams,
			}
			event.Write()
		}

		c.JSON(http.StatusOK, MessageResponse{
			Details: "Usuario DEC5 asociado con éxito",
			Code:    http.StatusOK,
		})
	})

}

// @Summary roleUserDec get
// @Description roleUserDec get
// @Tags roleUserDec
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} response
// @failure 404 {object} MessageResponse
// @Router /roleUserDec/{id} [get]
// @param id path string true "id"
func GetRoleIdUserDec(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get User Dec by Id
	router.GET("/roleUserDec/:id", func(c *gin.Context) {
		idUser := c.Param("id")
		userDec, userDecErr := db.GetUserRoleDec(idUser)

		if userDecErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: userDecErr.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		c.JSON(http.StatusOK, userDec)
	})
}
