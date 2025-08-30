package routes

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"github.com/imedcl/manager-api/pkg/services"
)

type CGIRoleParams struct {
	Dni         string `form:"dni"`
	Name        string `form:"name"`
	Institution string `form:"institution"`
	Country     string `form:"country"`
}

type CGIRolesResponse struct {
	Data       []data.CGIUser `json:"data"`
	Total      int            `json:"total"`
	Offset     int            `json:"offset"`
	LastOffset int            `json:"last_offset"`
}

type CGIUserRolesResponse struct {
	Data data.CGIUser `json:"data"`
}

type AddCGIRolesResponse struct {
	Data struct {
		Status bool `json:"status"`
	} `json:"data"`
}

// @Summary get cgi roles
// @Description get cgi roles
// @Tags autentia roles
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} CGIRolesResponse
// @failure 404 {object} MessageResponse
// @Router /autentia/roles/cgi [get]
// @param country query string true "country"
// @param institution query string true "institution"
func CGIGetRolesRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.GET("/roles/cgi", func(c *gin.Context) {
		offset := 0
		limit := 10
		country := c.Query("country")
		institution := c.Query("institution")

		if c.Query("offset") != "" {
			offset, _ = strconv.Atoi(c.Query("offset"))
		}
		if c.Query("limit") != "" {
			limit, _ = strconv.Atoi(c.Query("limit"))
		}
		roles, newOffset, err := data.GetCGIRoles(country, institution, offset, limit)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		total := offset + (limit * 2)
		c.JSON(http.StatusOK, CGIRolesResponse{Data: roles, Offset: newOffset, LastOffset: offset, Total: total})
	})
}

// @Summary get cgi user roles
// @Description get cgi user roles
// @Tags autentia roles
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} CGIUserRolesResponse
// @failure 404 {object} MessageResponse
// @Router /autentia/roles/user/{dni} [get]
// @param dni path string true "dni"
// @param country query string true "country"
// @param institution query string true "institution"
func CGIGetUserRolesRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.GET("/roles/user/:dni", func(c *gin.Context) {
		country := c.Query("country")
		institution := c.Query("institution")
		dni := c.Param("dni")
		roles, err := data.GetCGIRole(dni, country, institution)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		c.JSON(http.StatusOK, CGIUserRolesResponse{Data: roles})
	})
}

// @Summary post cgi user roles
// @Description post cgi user roles
// @Tags autentia roles
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} AddCGIRolesResponse
// @failure 400 {object} MessageResponse
// @Router /autentia/roles/user [post]
// @Param roles body CGIRoleParams true "roles"
func CGIPostUserRolesRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.POST("/roles/user", func(c *gin.Context) {
		var params CGIRoleParams
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

		person := services.GetPerson(params.Country, params.Dni)

		if person.Result.Error != "0" {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "Persona no encontrada",
				Code:    http.StatusNotFound,
			})
			return
		}
		status, err := data.AddCGIRole(params.Dni, params.Name, params.Country, params.Institution)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Roles de Autentia",
			Event:        fmt.Sprintf("Se agrega el rol %s al dni %s en la institución %s - %s", params.Name, params.Dni, params.Institution, params.Country),
		}
		event.Write()
		data := AddCGIRolesResponse{}
		data.Data.Status = status
		c.JSON(http.StatusOK, data)
	})
}

// @Summary delete cgi user roles
// @Description delete cgi user roles
// @Tags autentia roles
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} AddCGIRolesResponse
// @failure 400 {object} MessageResponse
// @Router /autentia/roles/cgi [delete]
// @Param roles body CGIRoleParams true "roles"
func CGIDeleteUserRolesRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.DELETE("/roles/cgi", func(c *gin.Context) {
		var params CGIRoleParams
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
		status, err := data.RemoveCGIRole(params.Dni, params.Name, params.Country, params.Institution)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Roles de Autentia",
			Event:        fmt.Sprintf("Se elimina el rol %s al dni %s en la institución %s - %s", params.Name, params.Dni, params.Institution, params.Country),
		}
		event.Write()
		data := AddCGIRolesResponse{}
		data.Data.Status = status
		c.JSON(http.StatusOK, data)
	})
}
