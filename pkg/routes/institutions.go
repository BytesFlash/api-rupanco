package routes

import (
	"net/http"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/sirupsen/logrus"
)

type institutionParams struct {
	Dni     string `form:"dni" binding:"required"`
	Name    string `form:"name" binding:"required"`
	Country string `form:"country" binding:"required"`
	Nemo    string `form:"nemo"`
	Owner   string `form:"owner"`
}

type CountriesResponse struct {
	Data []string `json:"data"`
}

type InstitutionResponse struct {
	Data *data.Institution `json:"data"`
}

type InstitutionsResponse struct {
	Data []*data.Institution `json:"data"`
}

func InstitutionsRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {

	router.GET("/institutions", func(c *gin.Context) {
		// country, status := c.GetQuery("country")
		// if !status {
		// 	c.JSON(http.StatusBadRequest, MessageResponse{
		// 		Details: "Bad Request",
		// 		Code:    http.StatusBadRequest,
		// 	})
		// 	return
		// }
		// institutions, err := db.ListInstitution(currentUser, country)
		// if err != nil {
		// 	c.JSON(http.StatusNotFound, MessageResponse{
		// 		Details: err.Error(),
		// 		Code:    http.StatusNotFound,
		// 	})
		// 	return
		// }
		// c.JSON(http.StatusOK, InstitutionsResponse{Data: institutions})
	})
}

func InstitutionsPostRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.POST("/institutions", func(c *gin.Context) {

		var params institutionParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		// institution := &data.Institution{
		// 	Name:    params.Name,
		// 	Country: params.Country,
		// 	Nemo:    params.Nemo,
		// 	Dni:     strings.ToLower(params.Dni),
		// }
		// owner, err := db.GetOwner(params.Owner)
		// if err == nil {
		// 	institution.Owner = owner
		// }
		// db.CreateInstitution(institution)
		// event := &events.EventLog{
		// 	UserID:   currentUser.Dni,
		// 	Resource: "Instituciones",
		// 	Event:    fmt.Sprintf("Se registra la institución %s en %s", params.Name, params.Country),
		// }
		// event.Write()
		// c.JSON(http.StatusOK, InstitutionResponse{Data: institution})
	})
}

func CountriesRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.GET("/countries", func(c *gin.Context) {
		// countries, err := db.GetCountries()
		// if err != nil {
		// 	c.JSON(http.StatusNotFound, MessageResponse{
		// 		Details: err.Error(),
		// 		Code:    http.StatusNotFound,
		// 	})
		// 	return
		// }
		// c.JSON(http.StatusOK, CountriesResponse{Data: countries})
	})
}

// @Summary institutions
// @Description institutions
// @Tags institution
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} []data.Institution
// @failure 409 {object} MessageResponse
// @Router /institutions/sync/{country} [get]
// @param country path string true "country"
func InstitutionCountriesRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.GET("/institutions/sync/:country", func(c *gin.Context) {

		countryName := c.Param("country")
		country, err := db.GetCountryByName(countryName)
		if err != nil {
			c.JSON(http.StatusConflict, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusConflict,
			})
			logrus.Printf("Error %s en la sincronización", err.Error())
			return
		}
		institutionList := syncInstitutions(country)
		c.JSON(http.StatusOK, institutionList)
	})
}
