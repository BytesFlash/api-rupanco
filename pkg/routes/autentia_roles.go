package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
)

type responseRole struct {
	Data []*data.AutentiaRole `json:"data"`
}

type paramsAutentiaRole struct {
	Name string `form:"name" json:"name"`
}

type paramsAutentiaRoleInstitution struct {
	Role        string `form:"role" json:"role"`
	Institution string `form:"institution" json:"institution"`
}

// @Summary post autentia role
// @Description post autentia role
// @Tags autentia roles
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 400 {object} MessageResponse
// @Router /autentia/roles [post]
// @Param role body paramsAutentiaRole true "role"
func AutentiaPostRoleRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//POST Role
	router.POST("/roles", func(c *gin.Context) {
		var params paramsAutentiaRole
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

		roleName := strings.ToUpper(params.Name)

		_, rolExistErr := db.GetAutentiaRoleByName(roleName)

		if roleName == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El nombre del rol no puede ser vacio",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if rolExistErr == nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El rol ya existe",
				Code:    http.StatusBadRequest,
			})
			return
		}

		nameAutenti := &data.AutentiaRole{
			Name: roleName,
		}
		db.CreateAutentiaRole(nameAutenti)
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Autentia Roles",
			Event:        fmt.Sprintf("Se crea Rol %s", roleName),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, MessageResponse{
			Details: "Rol creado con éxito",
			Code:    http.StatusOK,
		})

	})
}

// @Summary delete autentia role
// @Description delete autentia role
// @Tags autentia roles
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 400 {object} MessageResponse
// @Router /autentia/roles [delete]
// @Param role body paramsAutentiaRole true "role"
func AutentiaDeleteRoleRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//DELETE Role
	router.DELETE("/roles", func(c *gin.Context) {
		var params paramsAutentiaRole
		usuario, erru := GetUserFromToken(c)
		if erru != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": erru.Error()})
			return
		}

		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		_, roleErr := db.GetAutentiaRoleByName(params.Name)

		if roleErr != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Rol no disponible en el sistema Autentia",
				Code:    http.StatusBadRequest,
			})
			return
		}

		err := db.DeleteAutentiaRole(params.Name)

		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Autentia Roles",
			Event:        fmt.Sprintf("Se elimina Rol %s", params.Name),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, MessageResponse{
			Details: "Rol eliminado con éxito",
			Code:    http.StatusOK,
		})

	})
}

// @Summary delete autentia role
// @Description delete autentia role
// @Tags autentia roles
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} []data.AutentiaRole
// @failure 400 {object} MessageResponse
// @Router /autentia/roles [get]
func AutentiaGetRoleRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//Get all Roles
	router.GET("/roles", func(c *gin.Context) {
		role, err := db.GetAllAutentiaRole()
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Println("error al obtener roles")
			return
		}
		c.JSON(http.StatusOK, role)
	})

}
