package routes

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"github.com/sirupsen/logrus"
)

type roleResponse struct {
	Data *data.Role `json:"data"`
}
type roleBollResponse struct {
	Code    int  `json:"code"`
	Details bool `json:"details"`
}

type ModuleParams struct {
	Name    string `form:"name" json:"name" binding:"required"`
	Checked bool   `form:"checked" json:"checked" binding:"required"`
}

type rolesParams struct {
	Name    string         `form:"name" binding:"required"`
	Modules []ModuleParams `form:"modules" binding:"required"`
}

type rolesUpdateParams struct {
	Modules []ModuleParams `form:"modules" json:"modules" binding:"required"`
}

type roleModulesResponse struct {
	Data rolesUpdateParams `json:"data"`
}

type roleInstitutionParams struct {
	Roles []roleParams `form:"roles" binding:"required"`
}

// @Summary roles
// @Description get roles
// @Tags role
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} []data.Role
// @failure 404 {object} MessageResponse
// @Router /roles [get]
func RolesGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//get roles
	router.GET("/roles", func(c *gin.Context) {
		modules, err := db.GetAllRole()
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Println("error al obtener el roles")
			return
		}
		c.JSON(http.StatusOK, modules)
	})
}

// @Summary role
// @Description get role
// @Tags role
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} roleModulesResponse
// @failure 404 {object} MessageResponse
// @Router /roles/modules/{role} [get]
// @param role path string true "role"
func RoleGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//get role
	router.GET("/roles/modules/:role", func(c *gin.Context) {
		roleName := c.Param("role")
		role, roleErr := db.GetRole(roleName)

		if roleErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: roleErr.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Printf("error: %s ,al obtener el rol %s", roleErr.Error(), roleName)
			return
		}
		var roleResponse rolesUpdateParams
		var rolModErr error
		var moduleResponse ModuleParams
		modules, _ := db.ListAllModules()
		for _, module := range modules {
			_, rolModErr = db.GetRoleModuleByIDs(role.ID, module.ID)
			if rolModErr == nil {
				moduleResponse.Checked = true

			} else {
				moduleResponse.Checked = false
			}
			moduleResponse.Name = module.Name
			roleResponse.Modules = append(roleResponse.Modules, moduleResponse)
		}
		c.JSON(http.StatusOK, roleModulesResponse{Data: roleResponse})
	})
}

// @Summary active role
// @Description get active role
// @Tags role
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 409 {object} MessageResponse
// @Router /roles/active/{role} [get]
// @param role path string true "role"
func RoleUserRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//get active role
	router.GET("/roles/active/:role", func(c *gin.Context) {
		roleName := c.Param("role")
		role, roleErr := db.GetRole(roleName)

		if roleErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: roleErr.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Printf("error: %s ,al obtener el rol %s", roleErr.Error(), roleName)
			return
		}
		if _, uriErr := db.UserRoleInstitutionExistByRole(role.ID); uriErr == nil {
			c.JSON(http.StatusConflict, MessageResponse{
				Details: "El rol tiene usuarios asociados",
				Code:    http.StatusConflict,
			})
			return
		}
		c.JSON(http.StatusOK, MessageResponse{
			Details: "El rol no tiene usuarios asociados",
			Code:    http.StatusOK,
		})
	})
}

// @Summary role institution
// @Description get role institution
// @Tags role
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} bool
// @failure 404 {object} MessageResponse
// @Router /roles/institution [post]
// @param institution body roleInstitutionParams true "institution"
func RoleInstitutionRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get institution role module dependence
	router.POST("/roles/institution", func(c *gin.Context) {

		var params roleInstitutionParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			logrus.Printf("Error en consulta de rol con los parámetros: %+v\n", params)
			return
		}
		logrus.Printf("Parámetros consulta de rol: %+v\n", params)

		if len(params.Roles) == 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Lista de roles vacía",
				Code:    http.StatusBadRequest,
			})
			return
		}

		institutionDependence := []string{}

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

			modules, modulesErr := db.ListAllModules()
			if modulesErr != nil {
				c.JSON(http.StatusInternalServerError, MessageResponse{
					Details: modulesErr.Error(),
					Code:    http.StatusInternalServerError,
				})
				logrus.Printf("error: %s ,al listar los módulos", modulesErr.Error())
				return
			}

			for _, module := range modules {
				if _, errModule := db.GetRoleModuleByIDs(role.ID, module.ID); errModule == nil {

					institutionDependence = append(institutionDependence, module.Name)
				}
			}
		}

		c.JSON(http.StatusOK, institutionDependence)
	})
}

// @Summary role modules
// @Description post role modules
// @Tags role
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 400 {object} MessageResponse
// @Router /roles/modules [post]
// @param roles body rolesParams true "roles"
func RoleModulesRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Post role modules
	router.POST("/roles/modules", func(c *gin.Context) {
		var params rolesParams
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
		logrus.Printf("Parámetros registro de rol: %+v\n", params.Modules)

		roleName := strings.ToUpper(params.Name)

		_, rolExistErr := db.GetRoleByName(roleName)
		if rolExistErr == nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El rol ya existe",
				Code:    http.StatusBadRequest,
			})
			return
		}

		nameRegex := regexp.MustCompile(`^[{a-zA-Z- }]+$`)
		if !nameRegex.MatchString(roleName) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Nombre incorrecto, solo puede tener letras, guiones y espacios",
				Code:    http.StatusBadRequest,
			})
			return
		}
		atLeastOneModule := false
		for _, module := range params.Modules {
			if module.Checked {
				atLeastOneModule = true
			}
		}

		if !atLeastOneModule {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Debe seleccionar al menos un módulo",
				Code:    http.StatusBadRequest,
			})
			return
		}

		justSensor := true

		for _, module := range params.Modules {
			if module.Checked {
				if module.Name != "Lectores" {
					justSensor = false
				}
			}
		}
		roles, _ := db.GetAllRole()
		if !justSensor {
			roleExist := false
			var mod *data.Module
			var rolModErr error
			for _, role := range roles {
				for _, module := range params.Modules {
					mod, _ = db.GetModuleByName(module.Name)
					if module.Checked {
						_, rolModErr = db.GetRoleModuleByIDs(role.ID, mod.ID)
						if rolModErr == nil {
							roleExist = true
						} else {
							roleExist = false
							break
						}
					} else {
						_, rolModErr = db.GetRoleModuleByIDs(role.ID, mod.ID)
						if rolModErr != nil {
							roleExist = true
						} else {
							roleExist = false
							break
						}
					}
				}
				if roleExist {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "La combinación de módulos ya existe",
						Code:    http.StatusBadRequest,
					})
					return
				}
			}
		} else {
			roleExist := false
			var mod *data.Module
			var rolModErr error
			var sensorCount = 0
			for _, role := range roles {
				for _, module := range params.Modules {
					mod, _ = db.GetModuleByName(module.Name)
					if module.Checked {
						_, rolModErr = db.GetRoleModuleByIDs(role.ID, mod.ID)
						if rolModErr == nil {
							roleExist = true
						} else {
							roleExist = false
							break
						}
					} else {
						_, rolModErr = db.GetRoleModuleByIDs(role.ID, mod.ID)
						if rolModErr != nil {
							roleExist = true
						} else {
							roleExist = false
							break
						}
					}
				}
				if roleExist {
					sensorCount++
				}
			}
			if sensorCount >= 2 {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "La combinación de módulos ya existe",
					Code:    http.StatusBadRequest,
				})
				return
			}
		}

		role := &data.Role{
			Name: roleName,
		}
		db.CreateRole(role)
		newRole, _ := db.GetRole(role.Name)
		for _, module := range params.Modules {
			if module.Checked {
				moduleForRole, _ := db.GetModuleByName(module.Name)
				db.CreateDefaultRoleModule(newRole.ID, moduleForRole.ID)
			}
		}
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Roles",
			Event:        fmt.Sprintf("Se registra el rol %s", role.Name),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, MessageResponse{
			Details: "Rol creado con éxito",
			Code:    http.StatusOK,
		})

	})
}

// @Summary role edit modules
// @Description put role modules
// @Tags role
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 400 {object} MessageResponse
// @Router /roles/modules/{role} [put]
// @param role path string true "role"
// @param roles body rolesUpdateParams true "roles"
func RoleEditModulesRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.PUT("/roles/modules/:role", func(c *gin.Context) {
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		roleName := strings.ToUpper(c.Param("role"))
		role, roleErr := db.GetRole(roleName)

		if roleErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: roleErr.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Printf("error: %s ,al obtener el rol %s", roleErr.Error(), roleName)
			return
		}
		var params rolesUpdateParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		logrus.Printf("Parámetros registro de rol: %+v\n", params.Modules)

		justSensor := true

		for _, module := range params.Modules {
			if module.Checked {
				if module.Name != "Lectores" {
					justSensor = false
				}
			}
		}
		roles, _ := db.GetAllRole()
		if !justSensor {
			roleExist := false
			var mod *data.Module
			var rolModErr error
			for _, role := range roles {
				for _, module := range params.Modules {
					mod, _ = db.GetModuleByName(module.Name)
					if module.Checked {
						_, rolModErr = db.GetRoleModuleByIDs(role.ID, mod.ID)
						if rolModErr == nil {
							roleExist = true
						} else {
							roleExist = false
							break
						}
					} else {
						_, rolModErr = db.GetRoleModuleByIDs(role.ID, mod.ID)
						if rolModErr != nil {
							roleExist = true
						} else {
							roleExist = false
							break
						}
					}
				}
				if roleExist {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "La combinación de módulos ya existe",
						Code:    http.StatusBadRequest,
					})
					return
				}
			}
		} else {
			roleExist := false
			var mod *data.Module
			var rolModErr error
			var sensorCount = 0
			for _, role := range roles {
				for _, module := range params.Modules {
					mod, _ = db.GetModuleByName(module.Name)
					if module.Checked {
						_, rolModErr = db.GetRoleModuleByIDs(role.ID, mod.ID)
						if rolModErr == nil {
							roleExist = true
						} else {
							roleExist = false
							break
						}
					} else {
						_, rolModErr = db.GetRoleModuleByIDs(role.ID, mod.ID)
						if rolModErr != nil {
							roleExist = true
						} else {
							roleExist = false
							break
						}
					}
				}
				if roleExist {
					sensorCount++
				}
			}
			if sensorCount >= 2 {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "La combinación de módulos ya existe",
					Code:    http.StatusBadRequest,
				})
				return
			}
		}

		for _, module := range params.Modules {
			if module.Checked {
				moduleForRole, _ := db.GetModuleByName(module.Name)
				db.CreateDefaultRoleModule(role.ID, moduleForRole.ID)

			} else {
				moduleForRole, _ := db.GetModuleByName(module.Name)
				db.DeleteRoleModule(role.ID, moduleForRole.ID)
			}
		}
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Roles",
			Event:        fmt.Sprintf("Se actualiza el rol %s", role.Name),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, MessageResponse{
			Details: "Rol actualizado con éxito",
			Code:    http.StatusOK,
		})
	})
}

// @Summary delete role modules
// @Description delete role modules
// @Tags role
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} roleBollResponse
// @failure 409 {object} roleBollResponse
// @Router /roles/modules/{role} [delete]
// @param role path string true "role"
// @param delete query bool true "delete"
func DeleteRoleModulesRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//delete role modules
	router.DELETE("/roles/modules/:role", func(c *gin.Context) {
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		roleName := c.Param("role")
		confirmation := c.Query("delete")

		role, roleErr := db.GetRole(roleName)

		if roleErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: roleErr.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Printf("error: %s ,al obtener el rol %s", roleErr.Error(), roleName)
			return
		}

		if confirmation == "true" {
			if roleName == "SUPER ADMIN-MANAGER" || roleName == "ADMIN MANAGER (PAÍS)" {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "El rol no se puede eliminar",
					Code:    http.StatusBadRequest,
				})
				return
			}

			_, roleErrIns := db.GetRoleInstbyId(role.ID)
			if roleErrIns == nil {
				c.JSON(http.StatusConflict, MessageResponse{
					Details: "No se puede eliminar un rol asignado",
					Code:    http.StatusConflict,
				})
				return
			}
			var rolModErr error
			modules, _ := db.ListAllModules()
			for _, module := range modules {
				_, rolModErr = db.GetRoleModuleByIDs(role.ID, module.ID)
				if rolModErr == nil {
					db.DeleteRoleModule(role.ID, module.ID)
				}
			}

			err := db.DeleteRole(role.ID)

			if err != nil {
				logrus.Printf("error: %s ,al eliminar el rol %s", err.Error(), roleName)
				c.JSON(http.StatusNotModified, "No se ha podido eliminar el Rol")
			} else {
				PrtyParams, _ := events.PrettyParams(c.Params)
				event := &events.EventLog{
					UserNickname: usuario.NickName,
					Resource:     "Roles",
					Event:        fmt.Sprintf("Se elimina el rol %s", role.Name),
					Params:       PrtyParams,
				}
				event.Write()
				c.JSON(http.StatusOK, "El rol ha sido eliminado")
			}
		} else {
			if roleName == "SUPER ADMIN-MANAGER" || roleName == "ADMIN MANAGER (PAÍS)" {
				c.JSON(http.StatusConflict, roleBollResponse{
					Details: false,
					Code:    http.StatusConflict,
				})
				return
			} else {

				_, roleErrIns := db.GetRoleInstbyId(role.ID)
				if roleErrIns == nil {
					c.JSON(http.StatusConflict, roleBollResponse{
						Details: false,
						Code:    http.StatusConflict,
					})
					return
				}
				c.JSON(http.StatusOK, roleBollResponse{
					Details: true,
					Code:    http.StatusOK,
				})
				return
			}
		}

	})
}

// @Summary edit role
// @Description edit role
// @Tags role
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 "Se ha actualizado el rol correctamente"
// @failure 409 {object} MessageResponse
// @Router /roles/{role} [put]
// @param role path string true "role"
// @param name query string true "name"
func RoleEditRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//edit role name
	router.PUT("/roles/:role", func(c *gin.Context) {
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		roleName := c.Param("role")
		newName := c.Query("name")

		role, roleErr := db.GetRole(roleName)

		if roleErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: roleErr.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Printf("error: %s ,al obtener el rol %s", roleErr.Error(), roleName)
			return
		}

		_, updateRolErr := db.UpdateRole(role, newName)

		if updateRolErr != nil {
			c.JSON(http.StatusConflict, MessageResponse{
				Details: updateRolErr.Error(),
				Code:    http.StatusConflict,
			})
			logrus.Printf("error: %s ,al actualizar el rol %s\n", updateRolErr.Error(), roleName)
			return
		}

		PrtyParams, _ := events.PrettyParams(c.Params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Roles",
			Event:        fmt.Sprintf("Se actualiza rol %s por %s", roleName, newName),
			Params:       PrtyParams,
		}
		event.Write()

		logrus.Printf("Se actualiza con éxito el rol %+v\n", role)
		c.JSON(http.StatusOK, "Se ha actualizado el rol correctamente")
	})

}
