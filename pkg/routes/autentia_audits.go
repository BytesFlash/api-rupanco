package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/services"
)

type responseAudit struct {
	Data interface{} `json:"data"`
}

// @Summary get autentia audits
// @Description get autentia audits
// @Tags autentia audits
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responseAudit
// @failure 400 {object} MessageResponse
// @Router /autentia/audits/{code} [get]
// @param code path string true "code"
// @param country query string true "country"
func AutentiaAuditsRoute(router *gin.RouterGroup, db *data.DB) {
	// Get Audit
	router.GET("/audits/:code", func(c *gin.Context) {
		code := c.Param("code")
		country := c.Query("country")
		if country == "" || code == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Parametros incorrectos",
				Code:    http.StatusBadRequest,
			})
			return
		}
		audit := services.GetAudit(country, code)
		c.JSON(http.StatusOK, responseAudit{Data: audit})
	})
}
