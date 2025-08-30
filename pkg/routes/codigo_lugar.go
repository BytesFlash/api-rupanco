package routes

import (
	"fmt"
	"net/http"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"

	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
)

type CodigoLugarParams struct {
	CodigoLugar  string `json:"codigo_lugar"`
	Entidad      string `json:"entidad"`
	ConvenioBono string `json:"convenio_bono"`
}

type CodigosLugarRequest struct {
	Token   string              `json:"token"`
	Codigos []CodigoLugarParams `json:"codigos"`
}

func ValidateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		expectedToken := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjoiNGE1Njk1MGMtZDE0OS00ZDFlLWE3ZTctMjg2MDk0OWQ1OTAzIiwibmlja19uYW1lIjoiMW5pY29sYXNhIiwibmFtZSI6Ik5pY28iLCJkbmkiOiIxOTQ4NTQzOC0yIiwiZW1haWwiOiJuaWdvbnphbGV6QGF1dGVudGlhLmNsIiwidmFsaWRhdGVkIjp0cnVlLCJzdGF0dXNfdXNlciI6IkFjdGl2byIsImFjdGlfaW5zdCI6ZmFsc2UsImFjdGl2ZSI6dHJ1ZSwicGljdHVyZSI6IiIsImNvdW50cnkiOnsibmFtZSI6IkNISUxFIiwiYWN0aXZlIjp0cnVlfSwidXNlcl9yb2xlc19hdXRlbnRpYSI6bnVsbH0sImlzcyI6IkF1dGVudGlhIEFkbWluIn0.p13Gp07o-GL53muoC47yoAs3kFURi_HcUYyxyPrdu90"

		if token != expectedToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized. Invalid token.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// @Summary create or update codigo lugar
// @Description create or update codigo lugar
// @Tags codigolugar
// @Security BearerToken
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Failure 406 {object} MessageResponse
// @Router /codigoslugares [post]
// @Param codigoslugares body CodigosLugarRequest true "Array of CodigoLugar"
func PostCodigosLugarRoute(router *gin.Engine, db *data.DB, client *unleash.Client) {
	// Create or Update CodigosLugar
	router.POST("/codigoslugar", ValidateToken(), func(c *gin.Context) {
		var request CodigosLugarRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		var failed []string

		db.DeleteCodigoLugar()
		for _, params := range request.Codigos {
			codigo := &data.CodigoLugar{
				CodigoLugar:  params.CodigoLugar,
				Entidad:      params.Entidad,
				ConvenioBono: params.ConvenioBono,
			}
			if err := db.CreateCodigoLugar(codigo); err != nil {
				failed = append(failed, params.CodigoLugar)
			}
		}

		if len(failed) > 0 {
			c.JSON(http.StatusNotAcceptable, MessageResponse{
				Details: "Failed to process some codigos_lugar",
				Code:    http.StatusNotAcceptable,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "All codigos_lugar processed successfully",
			})
		}

		event := &events.EventLog{
			UserNickname: "Sin nombre",
			Resource:     "codigo de lugar",
			Event:        fmt.Sprintf("Se han registrado nuevos codigo de lugar"),
			Params:       "Dartos subido",
		}
		event.Write()
	})
}

func PostAddCodigosLugarRoute(router *gin.Engine, db *data.DB, client *unleash.Client) {
	// Add CodigosLugar
	router.POST("/codigoslugar/add", ValidateToken(), func(c *gin.Context) {
		var request CodigosLugarRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		var failed []string

		for _, params := range request.Codigos {
			codigo := &data.CodigoLugar{
				CodigoLugar:  params.CodigoLugar,
				Entidad:      params.Entidad,
				ConvenioBono: params.ConvenioBono,
			}
			if err := db.CreateCodigoLugar(codigo); err != nil {
				failed = append(failed, params.CodigoLugar)
			}
		}

		if len(failed) > 0 {
			c.JSON(http.StatusNotAcceptable, MessageResponse{
				Details: "Failed to process some codigos_lugar",
				Code:    http.StatusNotAcceptable,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "All codigos_lugar processed successfully",
			})
		}

		event := &events.EventLog{
			UserNickname: "Sin nombre",
			Resource:     "codigo de lugar",
			Event:        fmt.Sprintf("Se han registrado nuevos codigo de lugar"),
			Params:       "Dartos subido",
		}
		event.Write()
	})
}

// @Summary codigoslugar
// @Description codigoslugar
// @Tags codigoslugar
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} []data.codigoslugar
// @failure 409 {object} MessageResponse
// @Router /codigoslugar/{dni} [get]
// @param codigoslugar path string true "codigoslugar"
func GetCodigosLugarRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.GET("/codigoslugar/:dni", func(c *gin.Context) {

		DniEntidad := c.Param("dni")
		var country interface{}
		var err error

		if DniEntidad == "0076957430-1" {
			country, err = db.GetAllCodigoById()
		} else {
			country, err = db.GetCodigoById(DniEntidad)
		}

		if err != nil {
			c.JSON(http.StatusConflict, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusConflict,
			})
			return
		}
		c.JSON(http.StatusOK, country)
	})
}

//No copiar codigo
