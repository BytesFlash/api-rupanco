package routes

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/imedcl/manager-api/pkg/config"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"github.com/imedcl/manager-api/pkg/mail"
	"github.com/imedcl/manager-api/pkg/services"
)

type userStruct struct {
	Active    bool         `json:"active"`
	Country   data.Country `json:"country"`
	Email     string       `json:"email"`
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	NickName  string       `json:"nick_name"`
	Picture   string       `json:"picture"`
	Status    string       `json:"status_user"`
	Validated bool         `json:"validated"`
	ActiInst  bool         `json:"acti_inst"`
	DniEntity string       `form:"dni_entity" json: dni_entity"`
	Roles     []roleParams `json:"roles"`
}

type response struct {
	Data *data.User `json:"data"`
}

type feedbackResponse struct {
	Status bool `json:"status"`
}

type feedBackParams struct {
	Message string `form:"message" json:"message"`
	Url     string `form:"url" json:"url"`
	Browser string `form:"browser" json:"browser"`
	System  string `form:"system" json:"system"`
}

type userParams struct {
	NickName    string `form:"nickname" binding:"required"`
	Name        string `form:"name" binding:"required"`
	Email       string `form:"email" binding:"required"`
	Password    string `form:"password"`
	Description string `form:"description"`
	Country     string `form:"country" binding:"required"`
	Dni         string `form:"dni" binding:"required"`
	DniEntity   string `form:"dni_entity" json:"dni_entity"`
	ActiInst    bool   `form:"acti_inst" json:"acti_inst"`
}

type updateUserParams struct {
	NickName    string `form:"nickname"`
	Name        string `form:"name"`
	Email       string `form:"email"`
	Password    string `form:"password"`
	Description string `form:"description"`
	Country     string `form:"country"`
	Dni         string `form:"dni"`
	DniEntity   string `form:"dni_entity" json:"dni_entity"`
	StatusUser  string `form:"status_user" json:"status_user"`
	ActiInst    bool   `form:"acti_inst" json:"acti_inst"`
}

type roleManagerParams struct {
	Role        string `form:"role" binding:"required"`
	Institution string `form:"institution"`
	Country     string `form:"country"`
}

type userValidateParams struct {
	Token string `form:"token"`
}

type validateParams struct {
	Token    string `form:"token"`
	Password string `form:"password"`
	Recovery bool   `form:"recovery"`
}

type roleParams struct {
	Name string `form:"name"`
}

type userRoleParams struct {
	Institutions []roleParams `form:"institutions" binding:"required"`
	NickName     string       `form:"nickname" binding:"required"`
	Roles        []roleParams `form:"roles" binding:"required"`
}
type roleResp struct {
	Name    string `form:"name" json:"name"`
	Checked bool   `form:"checked" json:"checked"`
}

type userRoleResponse struct {
	Roles []roleResp `form:"roles" json:"roles"`
}

type RoleInstParams struct {
	Role         data.Role          `form:"role"`
	Institutions []data.Institution `form:"institutions"`
}

type UserRoleInstParams struct {
	User  data.User        `form:"user"`
	Roles []RoleInstParams `form:"roles"`
}

// @Summary user
// @Description user
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} UserRoleInstParams
// @failure 404 {object} MessageResponse
// @Router /users/roles/{email} [get]
// @param email path string true "email"
func UserRolesRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	const moduleName = "users"
	// Get User
	router.GET("/users/roles/:email", func(c *gin.Context) {
		email := (c.Param("email"))

		nick := email
		if email == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Parametros incorrectos",
				Code:    http.StatusBadRequest,
			})
			return
		}
		userEmail, errEmail := db.GetUserEmail(email)

		if errEmail != nil {
			nick = email
		} else {
			nick = userEmail.NickName
		}

		user, userErr := db.UserExistsName(nick)

		if userErr != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Usuario no encontrado",
				Code:    http.StatusBadRequest,
			})
			return
		}

		roles, err := db.GetDistinctUserRoles(user.ID)

		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		var userResponse UserRoleInstParams
		userResponse.User = *user

		for _, role := range roles {
			var roleResponse RoleInstParams

			institutions, _ := db.GetUserRolesInstitutionsByRole(role.Role.ID, user.ID)
			for _, inst := range institutions {
				roleResponse.Institutions = append(roleResponse.Institutions, *inst.Institution)
			}
			roleResponse.Role = *role.Role
			userResponse.Roles = append(userResponse.Roles, roleResponse)
		}

		c.JSON(http.StatusOK, userResponse)
	})
}

// @Summary users
// @Description users
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} []userStruct
// @failure 404 {object} MessageResponse
// @Router /users [get]
func UsersRolesRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.GET("/users", func(c *gin.Context) {

		var users []*data.User
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if _, err := db.UserHasRole("SUPER ADMIN-MANAGER", usuario.ID); err == nil {
			users = db.GetUsersByRole("ADMIN MANAGER (PAÍS)")
		}

		if user, err := db.UserHasRole("ADMIN MANAGER (PAÍS)", usuario.ID); err == nil {
			country, _ := db.GetCountryById(user.CountryID)
			users = db.GetUsersByCountry(country.ID)

		}

		var userDataList []userStruct
		for _, userInfo := range users {
			var user userStruct
			user.Active = *&userInfo.Active
			user.Country = *userInfo.Country
			user.Email = *&userInfo.Email
			user.ID = *&userInfo.ID
			user.Name = *&userInfo.Name
			user.NickName = *&userInfo.NickName
			user.Picture = *&userInfo.Picture
			user.Status = *&userInfo.StatusUser
			user.Validated = *&userInfo.Validated
			user.ActiInst = *&userInfo.ActiInst
			user.DniEntity = *&userInfo.DniEntity
			userRoles, _ := db.GetDistinctUserRoles(userInfo.ID)

			for _, roleInfo := range userRoles {
				var role roleParams
				role.Name = roleInfo.Role.Name
				user.Roles = append(user.Roles, role)
			}
			userDataList = append(userDataList, user)
		}

		c.JSON(http.StatusOK, userDataList)
	})
}

// @Summary users
// @Description users
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} []userStruct
// @failure 404 {object} MessageResponse
// @Router /users [get]
func UsersRolesEmailRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.GET("/users/email/:email", func(c *gin.Context) {
		email := (c.Param("email"))
		var users []*data.User
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if _, err := db.UserHasRole("SUPER ADMIN-MANAGER", usuario.ID); err == nil {
			users = db.GetUsersByRole("ADMIN MANAGER (PAÍS)")
		}

		if user, err := db.UserHasRole("ADMIN MANAGER (PAÍS)", usuario.ID); err == nil {
			country, _ := db.GetCountryById(user.CountryID)
			users = db.GetUsersByCountryAndEmail(country.ID, email)

		}

		var userDataList []userStruct
		for _, userInfo := range users {
			var user userStruct
			user.Active = *&userInfo.Active
			user.Country = *userInfo.Country
			user.Email = *&userInfo.Email
			user.ID = *&userInfo.ID
			user.Name = *&userInfo.Name
			user.NickName = *&userInfo.NickName
			user.Picture = *&userInfo.Picture
			user.Status = *&userInfo.StatusUser
			user.Validated = *&userInfo.Validated
			user.ActiInst = *&userInfo.ActiInst
			user.DniEntity = *&userInfo.DniEntity
			userRoles, _ := db.GetDistinctUserRoles(userInfo.ID)

			for _, roleInfo := range userRoles {
				var role roleParams
				role.Name = roleInfo.Role.Name
				user.Roles = append(user.Roles, role)
			}
			userDataList = append(userDataList, user)
		}

		c.JSON(http.StatusOK, userDataList)
	})
}

// @Summary user
// @Description create user
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Param user body userParams true "user"
// @Success 201 "Usuario Creado"
// @failure 400 {object} MessageResponse
// @Router /users [post]
func UserPostRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.POST("/users", func(c *gin.Context) {

		var params userParams
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
		name := strings.Title(params.Name)
		email := (params.Email)
		dni := strings.ToLower(params.Dni)
		emailRegex := regexp.MustCompile(`^([a-zA-Z0-9._-]+)@[a-zA-Z0-9-]+[.]{1}[a-zA-Z]{2,4}$`)
		nickName := strings.ToLower(params.NickName)
		if db.UserAllDBExists(nickName) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Usuario ya existe en la base de datos",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if !emailRegex.MatchString(email) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El email no es válido",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(params.Name) <= 3 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El Nombre debe tener más de 3 carácteres",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(dni) == 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Dni no es válido",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(params.Description) > 500 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "La descripción no debe tener más de 500 carácteres",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if len(nickName) <= 3 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El usuario debe tener más de 3 carácteres",
				Code:    http.StatusBadRequest,
			})
			return
		}
		nickNameRegex := regexp.MustCompile(`^[{a-zA-Z-}]+$`)
		if !nickNameRegex.MatchString(nickName) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Usuario incorrecto, solo puede tener letras y guiones",
				Code:    http.StatusBadRequest,
			})
			return
		}

		country, countryErr := db.GetCountryByName(strings.ToUpper(params.Country))
		if countryErr != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "País no encontrado",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if db.UserExistsEmail(email) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El email ya se encuentra registrado",
				Code:    http.StatusBadRequest,
			})
			return
		}
		user := &data.User{
			Name:        name,
			Email:       email,
			Password:    params.Password,
			Description: params.Description,
			Country:     country,
			NickName:    nickName,
			Dni:         dni,
			DniEntity:   params.DniEntity,
			ActiInst:    params.ActiInst,
		}
		db.CreateUser(user)
		mail.SendUserRegister(user)
		PrtyParams, _ := events.PrettyParams(user)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Usuarios",
			Event:        fmt.Sprintf("Se ha creado el usuario %s", nickName),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusCreated, "Usuario Creado")
	})
}

// @Summary user instrospection
// @Description instrospection
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} response
// @Router /users/introspection [get]
func UserInstrospectionRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {

	// Get User By Token
	router.GET("/users/introspection", func(c *gin.Context) {
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response{Data: usuario})
	})
}

// @Summary user get
// @Description user get
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} response
// @failure 404 {object} MessageResponse
// @Router /users/{userIdentifier} [get]
// @param userIdentifier path string true "userIdentifier"
func UserGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get User
	router.GET("/users/:userIdentifier", func(c *gin.Context) {

		userIdentifier := strings.ToLower(c.Param("userIdentifier"))
		_, err := uuid.Parse(userIdentifier)
		var user *data.User
		if err != nil {
			user, err = db.GetUser(userIdentifier)
		} else {
			user, err = db.GetUserByID(userIdentifier)
		}
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		c.JSON(http.StatusOK, response{Data: user})
	})
}

// @Summary user get
// @Description user get
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} response
// @failure 404 {object} MessageResponse
// @Router /users/{nickName} [get]
// @param nickName path string true "nickName"
func UserGetRouteNickname(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get User
	router.GET("/user/:nickName", func(c *gin.Context) {

		userNickName := strings.ToLower(c.Param("nickName"))
		_, err := uuid.Parse(userNickName)
		var name *string
		if err != nil {
			name, err = db.GetUserNickName(userNickName)
		}

		if err != nil || name == nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "User not found",
				Code:    http.StatusNotFound,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"name": *name})
	})
}

// @Summary update user
// @Description update user
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 "El usuario se ha actualizado con éxito"
// @failure 400 {object} MessageResponse
// @Router /users/{nickName} [put]
// @Param user body updateUserParams true "user"
// @param nickName path string true "userIdentifier"
func UserPutRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Update User
	router.PUT("/users/:nickName", func(c *gin.Context) {
		nickName := strings.ToLower(c.Param("nickName"))
		usuario, erru := GetUserFromToken(c)
		if erru != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": erru.Error()})
			return
		}
		if !db.UserExists(nickName) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Usuario no encontrado",
				Code:    http.StatusBadRequest,
			})
			return
		}
		var params updateUserParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		newNickName := strings.ToLower(params.NickName)
		if newNickName != "" {
			if len(newNickName) <= 3 {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "El usuario debe tener más de 3 carácteres",
					Code:    http.StatusBadRequest,
				})
				return
			}
			nickNameRegex := regexp.MustCompile(`^[{a-zA-Z-}]+$`)
			if !nickNameRegex.MatchString(newNickName) {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "Usuario incorrecto, solo puede tener letras y guiones",
					Code:    http.StatusBadRequest,
				})
				return
			}
		}
		country, countryErr := db.GetCountryByName(strings.ToUpper(params.Country))
		if params.Country != "" {
			if countryErr != nil {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "País no encontrado",
					Code:    http.StatusBadRequest,
				})
				return
			}
		}

		name := strings.Title(params.Name)
		if params.Name != "" {
			if len(name) <= 3 {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "El Nombre debe tener más de 3 carácteres",
					Code:    http.StatusBadRequest,
				})
				return
			}
		}
		email := (params.Email)
		emailRegex := regexp.MustCompile(`^([a-zA-Z0-9._-]+)@[a-zA-Z0-9-]+[.]{1}[a-zA-Z]{2,4}$`)
		if params.Email != "" {
			if !emailRegex.MatchString(email) {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "El email no es válido",
					Code:    http.StatusBadRequest,
				})
				return
			}
		}

		userId, _ := db.UserExistsName(nickName)
		userEmail, _ := db.UserExistEmail(email)
		if userEmail.ID != userId.ID {
			if db.UserExistsEmail(email) {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "El email ya se encuentra registrado",
					Code:    http.StatusBadRequest,
				})
				return
			}

		}

		if len(params.Description) > 500 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "La descripción no debe tener más de 500 carácteres",
				Code:    http.StatusBadRequest,
			})
			return
		}
		dni := strings.ToLower(params.Dni)
		user := &data.User{
			NickName:    newNickName,
			Name:        name,
			Email:       email,
			StatusUser:  params.StatusUser,
			Description: params.Description,
			Country:     country,
			Dni:         dni,
			DniEntity:   params.DniEntity,
			ActiInst:    params.ActiInst,
		}
		if params.Password != "" {
			passValidate, passError := config.Password(params.Password)
			if !passValidate {
				c.JSON(http.StatusBadRequest, PasswordMessageResponse{
					Details:      "La contraseña no es válida",
					Code:         http.StatusBadRequest,
					Requirements: passError,
				})
				return
			}
			result := db.PasswordExists(params.Password, user.ID)
			if result == nil {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "No puede registrar una contraseña anteriormente utilizada",
					Code:    http.StatusBadRequest,
				})
				return
			} else {
				db.CreatePassword(user.Password, user)
			}
		}
		if params.StatusUser != "" {
			listArray := contains(config.ALL_STATUS, user.StatusUser)
			if !listArray {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "Código de estado incorrecto",
					Code:    http.StatusBadRequest,
				})
				return
			}
		}
		_, err := db.UpdateUser(nickName, user, params.Password)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Usuarios",
			Event:        fmt.Sprintf("Se ha actualizado el usuario %s", nickName),
			Params:       PrtyParams,
		}
		event.Write()

		c.JSON(http.StatusOK, "El usuario se ha actualizado con éxito")

	})
}

// @Summary user roles
// @Description user roles
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Param roles body userRoleParams true "roles"
// @Success 200 "Se han asignado los roles correctamente"
// @failure 400 {object} MessageResponse
// @Router /users/roles [post]
func UserRolesPostRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Add role to user
	router.POST("/users/roles", func(c *gin.Context) {

		var params userRoleParams
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
			logrus.Printf("Error en registro de rol con los parámetros: %+v\n", params)
			return
		}
		logrus.Printf("Parámetros registro de rol: %+v\n", params)
		nickName := strings.ToLower(params.NickName)
		user, userErr := db.GetUser(nickName)

		if userErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: userErr.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Printf("error: %s ,al obtener el usuario %s", userErr.Error(), nickName)
			return
		}

		if len(params.Roles) == 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Lista de roles vacía",
				Code:    http.StatusBadRequest,
			})
			return
		}

		for _, roleJson := range params.Roles {
			role, roleErr := db.GetRole(roleJson.Name)

			if roleErr != nil {
				c.JSON(http.StatusNotFound, MessageResponse{
					Details: roleErr.Error(),
					Code:    http.StatusNotFound,
				})
				logrus.Printf("error: %s ,al obtener el rol %s", roleErr.Error(), roleJson.Name)
				return
			}

			var institutionDependence = false

			modules, _ := db.ListAllModules()

			for _, module := range modules {
				if module.Name == "Usuarios Sistema Autentia" || module.Name == "Lectores" || module.Name == "LME" || module.Name == "Usuario DEC" || module.Name == "Vigencia de Cédula" || module.Name == "Reporte Operaciones" {
					if _, errModule := db.GetRoleModuleByIDs(role.ID, module.ID); errModule == nil {
						institutionDependence = true
					}
				}
			}

			if institutionDependence {
				for _, instJson := range params.Institutions {
					specialInst, instErr := db.GetInstitution(instJson.Name, user.Country.ID)

					if instErr != nil {
						c.JSON(http.StatusNotFound, MessageResponse{
							Details: instErr.Error(),
							Code:    http.StatusNotFound,
						})
						logrus.Printf("error: %s ,al obtener la institución %s", instErr.Error(), instJson.Name)
						return
					}

					for _, module := range modules {
						_, err := db.GetRoleModuleByIDs(role.ID, module.ID)
						if err == nil {
							valueModule, _ := db.GetModuleByID(module.ID)
							if valueModule.Name == "Usuarios Sistema Autentia" {
								person := services.GetPerson(user.Country.Name, user.Dni)

								if person.Result.Error != "0" {
									c.JSON(http.StatusNotFound, MessageResponse{
										Details: "Persona no enrolada, no se le puede asignar el rol Usuarios Sistema Autentia",
										Code:    http.StatusNotFound,
									})
									return
								}

								_, err := data.AddCGIRole(user.Dni, "ADMIN", user.Country.Name, instJson.Name)
								if err != nil {
									c.JSON(http.StatusNotFound, MessageResponse{
										Details: err.Error(),
										Code:    http.StatusNotFound,
									})
									return
								}
							}
						}

					}

					userRolInstitution := &data.UserRoleInstitution{
						UserID:        user.ID,
						RoleID:        role.ID,
						InstitutionID: specialInst.ID,
					}

					createRolErr := db.CreateUserRolInstitution(userRolInstitution)

					if createRolErr != nil {
						c.JSON(http.StatusConflict, MessageResponse{
							Details: createRolErr.Error(),
							Code:    http.StatusConflict,
						})
						logrus.Printf("error: %s ,al registrar el rol %+v\n", createRolErr.Error(), userRolInstitution)
						return
					}
					PrtyParams, _ := events.PrettyParams(params)
					event := &events.EventLog{
						UserNickname: usuario.NickName,
						Resource:     "Roles Manager",
						Event:        fmt.Sprintf("Se registra rol %s para el usuario %s", roleJson.Name, nickName),
						Params:       PrtyParams,
					}
					event.Write()

					logrus.Printf("Se registra con éxito el rol %+v\n", userRolInstitution)
				}
			} else {

				instName := fmt.Sprintf("*.%s*", strings.ToLower(user.Country.Name))

				institution, _ := db.GetInstitution(instName, user.Country.ID)

				userRolInstitution := &data.UserRoleInstitution{
					UserID:        user.ID,
					RoleID:        role.ID,
					InstitutionID: institution.ID,
				}

				createRolErr := db.CreateUserRolInstitution(userRolInstitution)

				if createRolErr != nil {
					c.JSON(http.StatusConflict, MessageResponse{
						Details: createRolErr.Error(),
						Code:    http.StatusConflict,
					})
					logrus.Printf("error: %s ,al registrar el rol %+v\n", createRolErr.Error(), userRolInstitution)
					return
				}
				PrtyParams, _ := events.PrettyParams(params)
				event := &events.EventLog{
					UserNickname: usuario.NickName,
					Resource:     "Roles Manager",
					Event:        fmt.Sprintf("Se registra rol %s para el usuario %s", roleJson.Name, nickName),
					Params:       PrtyParams,
				}
				event.Write()

				logrus.Printf("Se registra con éxito el rol %+v\n", userRolInstitution)
			}

		}

		c.JSON(http.StatusOK, "Se han asignado los roles correctamente")
	})
}

// @Summary update user roles
// @Description update user roles
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 "Se han asignado los roles correctamente"
// @failure 400 {object} MessageResponse
// @Router /users/institutions/roles [put]
// @Param user body userRoleParams true "user"
func UserRolesPutRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Add role to user
	router.PUT("/users/institutions/roles", func(c *gin.Context) {

		var params userRoleParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			logrus.Printf("Error en registro de rol con los parámetros: %+v\n", params)
			return
		}
		logrus.Printf("Parámetros registro de rol: %+v\n", params)
		nickName := strings.ToLower(params.NickName)
		user, userErr := db.GetUser(nickName)

		if userErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: userErr.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Printf("error: %s ,al obtener el usuario %s", userErr.Error(), nickName)
			return
		}

		// Obtener roles actuales
		existingRoles, err := db.GetRoleInstbyUser(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, MessageResponse{
				Details: "Error al obtener roles actuales del usuario",
				Code:    http.StatusInternalServerError,
			})
			logrus.Errorf("Error al obtener roles actuales del usuario %s: %v", nickName, err)
			return
		}

		existingMap := make(map[string]*data.UserRoleInstitution)
		for _, r := range existingRoles {
			key := fmt.Sprintf("%s-%s", r.RoleID, r.InstitutionID)
			existingMap[key] = r
		}

		newMap := make(map[string]bool)
		modules, _ := db.ListAllModules()

		for _, roleJson := range params.Roles {
			role, roleErr := db.GetRole(roleJson.Name)

			if roleErr != nil {
				c.JSON(http.StatusNotFound, MessageResponse{
					Details: roleErr.Error(),
					Code:    http.StatusNotFound,
				})
				logrus.Printf("error: %s ,al obtener el rol %s", roleErr.Error(), roleJson.Name)
				return
			}

			institutionDependence := false
			for _, module := range modules {
				if module.Name == "Usuarios Sistema Autentia" || module.Name == "Lectores" || module.Name == "LME" || module.Name == "Usuario DEC" || module.Name == "Vigencia de Cédula" || module.Name == "Reporte Operaciones" {
					if _, errModule := db.GetRoleModuleByIDs(role.ID, module.ID); errModule == nil {
						institutionDependence = true
						break
					}
				}
			}

			if institutionDependence {
				for _, instJson := range params.Institutions {
					specialInst, instErr := db.GetInstitution(instJson.Name, user.Country.ID)

					if instErr != nil {
						c.JSON(http.StatusNotFound, MessageResponse{
							Details: instErr.Error(),
							Code:    http.StatusNotFound,
						})
						logrus.Printf("error: %s ,al obtener la institución %s", instErr.Error(), instJson.Name)
						return
					}
					for _, module := range modules {
						if _, err := db.GetRoleModuleByIDs(role.ID, module.ID); err == nil {
							valueModule, _ := db.GetModuleByID(module.ID)
							if valueModule.Name == "Usuarios Sistema Autentia" {
								person := services.GetPerson(user.Country.Name, user.Dni)

								if person.Result.Error != "0" {
									c.JSON(http.StatusNotFound, MessageResponse{
										Details: "Persona no enrolada, no se le puede asignar el rol Usuarios Sistema Autentia",
										Code:    http.StatusNotFound,
									})
									logrus.Printf("Persona no enrolada: DNI %s, país %s", user.Dni, user.Country.Name)
									return
								}

								if _, err := data.AddCGIRole(user.Dni, "ADMIN", user.Country.Name, instJson.Name); err != nil {
									c.JSON(http.StatusNotFound, MessageResponse{
										Details: err.Error(),
										Code:    http.StatusNotFound,
									})
									logrus.Printf("Error añadiendo CGI role: %v", err)
									return
								}
							}
						}
					}
					key := fmt.Sprintf("%s-%s", role.ID, specialInst.ID)
					newMap[key] = true

					if _, exists := existingMap[key]; !exists {
						err := createAndLogUserRole(db, user, role, specialInst, params, nickName)
						if err != nil {
							c.JSON(http.StatusConflict, MessageResponse{
								Details: err.Error(),
								Code:    http.StatusConflict,
							})
							return
						}
					}
				}
			} else {
				instName := fmt.Sprintf("*.%s*", strings.ToLower(user.Country.Name))
				institution, _ := db.GetInstitution(instName, user.Country.ID)

				key := fmt.Sprintf("%s-%s", role.ID, institution.ID)
				newMap[key] = true

				if _, exists := existingMap[key]; !exists {
					err := createAndLogUserRole(db, user, role, institution, params, nickName)
					if err != nil {
						c.JSON(http.StatusConflict, MessageResponse{
							Details: err.Error(),
							Code:    http.StatusConflict,
						})
						return
					}
				}
			}
		}

		// Eliminar roles que ya no están en la nueva lista
		for key, existingRole := range existingMap {
			if !newMap[key] {
				if err := db.DeleteUserRolInstitutionId(existingRole.UserID, existingRole.InstitutionID, existingRole.RoleID); err != nil {
					c.JSON(http.StatusInternalServerError, MessageResponse{
						Details: fmt.Sprintf("Error eliminando el rol %s para la institución %s", existingRole.RoleID, existingRole.InstitutionID),
						Code:    http.StatusInternalServerError,
					})
					logrus.Errorf("Error eliminando rol: %v", err)
					return
				}
				logrus.Printf("Rol eliminado: %v\n", existingRole)
			}
		}

		c.JSON(http.StatusOK, "Se han asignado los roles correctamente")
	})
}

func createAndLogUserRole(db *data.DB, user *data.User, role *data.Role, institution *data.Institution, params userRoleParams, nickName string) error {
	userRoleInstitution := &data.UserRoleInstitution{
		UserID:        user.ID,
		RoleID:        role.ID,
		InstitutionID: institution.ID,
	}
	if err := db.CreateUserRolInstitution(userRoleInstitution); err != nil {
		logrus.Errorf("Error creando rol %+v: %v", userRoleInstitution, err)
		return err
	}

	PrtyParams, _ := events.PrettyParams(params)
	event := &events.EventLog{
		UserNickname: nickName,
		Resource:     "Roles Manager",
		Event:        fmt.Sprintf("Se registra rol %s para el usuario %s", role.Name, nickName),
		Params:       PrtyParams,
	}
	event.Write()
	logrus.Printf("Se registra con éxito el rol %+v\n", userRoleInstitution)
	return nil
}

// @Summary get user roles
// @Description get user roles
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} userRoleResponse
// @failure 400 {object} MessageResponse
// @Router /users/institutions/roles/{nickName} [get]
// @Param nickName path string true "nickName"
// @Param institution query string true "institution"
func UserRolesGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get institution user roles
	router.GET("/users/institutions/roles/:nickName", func(c *gin.Context) {

		nickName := strings.ToLower(c.Param("nickName"))
		institutionName := c.Query("institution")

		user, userErr := db.GetUser(nickName)

		if userErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: userErr.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Printf("error: %s ,al obtener el usuario %s", userErr.Error(), nickName)
			return
		}

		institution, instErr := db.GetInstitution(institutionName, user.Country.ID)

		if instErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: instErr.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Printf("error: %s ,al obtener la institución %s", instErr.Error(), institutionName)
			return
		}
		var rolesResponse userRoleResponse
		var roleResponse roleResp
		roles, _ := db.GetAllRole()
		for _, role := range roles {

			if _, err := db.GetUserRoleInstitution(user.ID, role.ID, institution.ID); err != nil {
				roleResponse.Checked = false
			} else {
				roleResponse.Checked = true

			}
			roleResponse.Name = role.Name
			rolesResponse.Roles = append(rolesResponse.Roles, roleResponse)
		}

		c.JSON(http.StatusOK, rolesResponse)
	})
}

// @Summary delete user roles
// @Description delete user roles
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 "Se ha eliminado el rol correctamente"
// @failure 409 {object} MessageResponse
// @Router /users/roles/{role} [delete]
// @Param role path string true "role"
// @Param nickname query string true "nickname"
// @Param institution query string true "institution"
func UserRolesDeleteRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Delete user role
	router.DELETE("/users/roles/:role", func(c *gin.Context) {
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		roleName := c.Param("role")
		nickName := strings.ToLower(c.Query("nickname"))
		institutionName := c.Query("institution")

		user, userErr := db.GetUser(nickName)

		if userErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: userErr.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Printf("error: %s ,al obtener el usuario %s", userErr.Error(), nickName)
			return
		}

		institution, instErr := db.GetInstitution(institutionName, user.Country.ID)

		if instErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: instErr.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Printf("error: %s ,al obtener la institución %s", instErr.Error(), institutionName)
			return
		}

		role, roleErr := db.GetRole(roleName)

		if roleErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: roleErr.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Printf("error: %s ,al obtener el rol %s", roleErr.Error(), roleName)
			return
		}

		deleteRolErr := db.DeleteUserRolInstitution(user, institution, role)

		if deleteRolErr != nil {
			c.JSON(http.StatusConflict, MessageResponse{
				Details: deleteRolErr.Error(),
				Code:    http.StatusConflict,
			})
			logrus.Printf("error: %s ,al eliminar el rol %+v\n", deleteRolErr.Error(), roleName)
			return
		}
		PrtyParams, _ := events.PrettyParams(c.Params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Roles",
			Event:        fmt.Sprintf("Se elimina rol %s para el usuario %s", roleName, nickName),
			Params:       PrtyParams,
		}
		event.Write()

		logrus.Printf("Se elimina con éxito el rol %+v\n", roleName)
		c.JSON(http.StatusOK, "Se ha eliminado el rol correctamente")
	})
}

// @Summary user feedback
// @Description user feedback
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 201 {object} feedbackResponse
// @failure 400 {object} MessageResponse
// @Router /users/feedback [post]
// @Param feedback body feedBackParams true "feedback"
func UserFeedbackRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Send Feedback Mail
	router.POST("/users/feedback", func(c *gin.Context) {
		var params feedBackParams
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
		mail.SendFeedback(usuario, params.Url, params.Browser, params.System, params.Message)
		c.JSON(http.StatusCreated, feedbackResponse{Status: true})
	})
}

// @Summary user logout
// @Description user logout
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 400 "Usuario no activo"
// @Router /logout [delete]
func LogoutRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Logout User
	router.DELETE("/logout", func(c *gin.Context) {
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if !db.IsActive(usuario.ID) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Usuario no activo",
				Code:    http.StatusBadRequest,
			})
			return
		} else {
			db.ExpireToken(usuario)
			c.JSON(http.StatusOK, MessageResponse{Details: "Sesión expirada", Code: http.StatusOK})
			event := &events.EventLog{
				UserNickname: usuario.NickName,
				Resource:     "Logout",
				Event:        "Successful Logout",
			}
			event.Write()
		}

	})
}

// @Summary user delete
// @Description user delete
// @Tags user
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 400 {object} MessageResponse
// @Router /users/{nickName} [delete]
// @param nickName path string true "nickName"
func DeleteUserRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Delete User
	router.DELETE("/users/:nickName", func(c *gin.Context) {
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		nickName := strings.ToLower(c.Param("nickName"))

		user, err := db.GetUser(nickName)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Usuario no Registrado",
				Code:    http.StatusBadRequest,
			})
			return
		}

		user.StatusUser = config.STATUS_LOCKED

		_, updateErr := db.UpdateUser(nickName, user, "")

		if updateErr != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al bloquear usuario, intente nuevamente",
				Code:    http.StatusBadRequest,
			})
			return
		}
		PrtyParams, _ := events.PrettyParams(c.Params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Usuarios",
			Event:        fmt.Sprintf("Se ha bloqueado al usuario %s", nickName),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, MessageResponse{Details: "Usuario bloqueado exitosamente", Code: http.StatusOK})
	})

}
