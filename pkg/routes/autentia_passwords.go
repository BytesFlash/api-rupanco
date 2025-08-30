package routes

import (
	"fmt"
	"net/http"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"

	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"github.com/imedcl/manager-api/pkg/services"
)

type responsePasswords struct {
	Data struct {
		Status bool `json:"status"`
	} `json:"data"`
}

type passwordParams struct {
	Password    string `form:"password"`
	Institution string `form:"institution"`
	Country     string `form:"country"`
	System      string `form:"system"`
}

// @Summary put autentia password
// @Description put autentia password
// @Tags autentia passwords
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responsePasswords
// @failure 400 {object} MessageResponse
// @Router /autentia/passwords/{dni} [put]
// @param dni path string true "dni"
// @Param password body passwordParams true "password"
func AutentiaPasswordsRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get Sensor
	router.PUT("/passwords/:dni", func(c *gin.Context) {
		dni := c.Param("dni")
		var params passwordParams
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if dni == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Parametros incorrectos",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		var response responsePasswords
		response.Data.Status = services.ChangePassword(dni, params.Password, params.Country, params.Institution, params.System)
		PrtyParams, _ := events.PrettyParams(c.Params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Autentia Passwords",
			Event:        fmt.Sprintf("Se actualiza contraseña"),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, response)
	})
}
