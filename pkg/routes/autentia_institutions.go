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
)

type responseInstitution struct {
	//Data data.Sensor `json:"data"`
	Data interface{} `json:"data"`
}

type institutionCreateParams struct {
	Name        string `form:"name"`
	Nemo        string `form:"nemo"`
	Country     string `form:"country" binding:"required"`
	Description string `form:"description"`
	Email       string `form:"email" binding:"required"`
	State       int    `form:"state"`
	Dec         int    `form:"dec"`
	Dni         string `form:"dni" binding:"required"`
	Verified    string `form:"verified"`
	Owner       string `form:"owner" binding:"required"`
}

type institutionUpdateParams struct {
	Country     string `form:"country" binding:"required"`
	Description string `form:"description"`
	Email       string `form:"email"`
	State       int    `form:"state"`
	Dec         int    `form:"dec"`
	Owner       string `form:"owner"`
}

// @Summary get autentia institution
// @Description get autentia institution
// @Tags autentia institutions
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responseInstitution
// @failure 400 {object} MessageResponse
// @Router /autentia/institutions/{name} [get]
// @param name path string true "name"
// @param country query string true "country"
func AutentiaGetInstitutionRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get Institution
	router.GET("/institutions/:name", func(c *gin.Context) {
		name := c.Param("name")
		countryName := strings.ToUpper(c.Query("country"))

		if countryName == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Parametros incorrectos",
				Code:    http.StatusBadRequest,
			})
			return
		}
		country, err := db.GetCountryByName(countryName)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		localInstitution, err := db.GetInstitution(name, country.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		c.JSON(http.StatusOK, responseInstitution{Data: localInstitution})
	})
}

// @Summary get autentia institutions
// @Description get autentia institutions
// @Tags autentia institutions
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responseInstitution
// @failure 400 {object} MessageResponse
// @Router /autentia/institutions [get]
// @param country query string true "country"
func AutentiaGetInstitutionsRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// List Institutions by country
	router.GET("/institutions", func(c *gin.Context) {
		countryName := strings.ToUpper(c.Query("country"))
		if countryName == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Parametros incorrectos",
				Code:    http.StatusBadRequest,
			})
			return
		}
		country, err := db.GetCountryByName(countryName)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		institutions, err := db.ListAllInstitutionsByCountry(country.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		c.JSON(http.StatusOK, responseInstitution{Data: institutions})
	})
}

// @Summary get all autentia institutions
// @Description get all autentia institutions
// @Tags autentia institutions
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responseInstitution
// @failure 400 {object} MessageResponse
// @Router /autentia/institutions/all [get]
func AutentiaGetAllInstitutionsRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// List Institutions
	router.GET("/institutions/all", func(c *gin.Context) {

		institutions, err := db.ListAllInstitutions()
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		c.JSON(http.StatusOK, responseInstitution{Data: institutions})
	})
}

// @Summary create institution
// @Description create institution
// @Tags autentia institutions
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responseInstitution
// @failure 406 {object} MessageResponse
// @Router /autentia/institutions [post]
// @Param institution body institutionCreateParams true "institution"
func AutentiaPostInstitutionsRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Create Institution
	router.POST("/institutions", func(c *gin.Context) {

		var params institutionCreateParams
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
		name := strings.ToUpper(strings.TrimSpace(params.Name))
		if len(name) < 3 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El nombre debe tener 3 o más carácteres",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if strings.Contains(name, " ") {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El nombre no puede contener espacios",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if params.State != 1 && params.State != 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El estado no es válido",
				Code:    http.StatusBadRequest,
			})
			return
		}
		description := strings.TrimRight(params.Description, " ")
		description = strings.TrimLeft(description, " ")
		if description == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Debe ingresar una descripción para la institución",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(description) > 500 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "La descripción no debe tener más de 500 carácteres",
				Code:    http.StatusBadRequest,
			})
			return
		}
		countryName := strings.ToUpper(params.Country)

		country, err := db.GetCountryByName(countryName)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		email := (params.Email)
		emailRegex := regexp.MustCompile(`^([a-zA-Z0-9._-]+)@[a-zA-Z0-9-]+[.]{1}[a-zA-Z]{2,4}$`)
		if !emailRegex.MatchString(email) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El email no es válido",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if params.Dec != 1 && params.Dec != 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El DEC no es válido",
				Code:    http.StatusBadRequest,
			})
			return
		}
		nemo := strings.ToUpper(strings.TrimSpace(params.Nemo))
		if nemo == "" || strings.Contains(nemo, " ") {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Debe ingresar un Nemo válido",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(nemo) != 4 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El Nemo debe tener 4 carácteres",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if _, nemoErr := db.InstitutionsExistsNemo(nemo); nemoErr == nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El Nemo ya existe",
				Code:    http.StatusBadRequest,
			})
			return
		}
		dni := strings.ToLower(params.Dni)
		dniRegex := regexp.MustCompile(`^\d{6,}[0-9-]{1}?[0-9kK]{1}?$`)
		if !dniRegex.MatchString(dni) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El dni no es válido",
				Code:    http.StatusBadRequest,
			})
			return
		}
		institution := &data.Institution{
			Name:        name,
			Country:     country,
			Nemo:        nemo,
			Description: description,
			Email:       params.Email,
			FlagDec:     params.Dec,
			State:       params.State,
			Dni:         strings.ToLower(params.Dni),
		}
		if !db.InstitutionExists(institution.Name, institution.Country.ID) {
			owner, err := db.GetOwner(params.Owner)
			if err == nil {
				institution.Owner = owner
			} else {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "El propietario no es válido",
					Code:    http.StatusBadRequest,
				})
				return
			}
			db.CreateInstitution(institution)
			PrtyParams, _ := events.PrettyParams(params)
			event := &events.EventLog{
				UserNickname: usuario.NickName,
				Resource:     "Instituciones",
				Event:        fmt.Sprintf("Se registra la institución %s en %s", params.Name, params.Country),
				Params:       PrtyParams,
			}
			event.Write()
			c.JSON(http.StatusOK, responseInstitution{Data: institution})
		} else {
			c.JSON(http.StatusNotAcceptable, MessageResponse{
				Details: "El nombre de la institución en el país indicado ya se encuentra registrado",
				Code:    http.StatusNotAcceptable,
			})
		}
	})
}

// @Summary update institution
// @Description update institution
// @Tags autentia institutions
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responseInstitution
// @failure 400 {object} MessageResponse
// @Router /autentia/institutions/{name} [put]
// @param name path string true "name"
// @Param institution body institutionUpdateParams true "institution"
func AutentiaPutInstitutionsRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Update Institution
	router.PUT("/institutions/:name", func(c *gin.Context) {

		name := c.Param("name")
		var params institutionUpdateParams
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
		countryName := strings.ToUpper(params.Country)

		country, err := db.GetCountryByName(countryName)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if db.InstitutionExists(name, country.ID) {
			description := strings.TrimRight(params.Description, " ")
			description = strings.TrimLeft(description, " ")
			if description == "" {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "Debe ingresar una descripción para la institución",
					Code:    http.StatusBadRequest,
				})
				return
			}
			if len(description) > 500 {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "La descripción no debe tener más de 500 carácteres",
					Code:    http.StatusBadRequest,
				})
				return
			}
			if params.State != 1 && params.State != 0 {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "El estado no es válido",
					Code:    http.StatusBadRequest,
				})
				return
			}
			email := (params.Email)
			if email != "" {
				emailRegex := regexp.MustCompile(`^([a-zA-Z0-9._-]+)@[a-zA-Z0-9-]+[.]{1}[a-zA-Z]{2,4}$`)
				if !emailRegex.MatchString(email) {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "El email no es válido",
						Code:    http.StatusBadRequest,
					})
					return
				}
			}
			if params.Dec != 1 && params.Dec != 0 {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "El DEC no es válido",
					Code:    http.StatusBadRequest,
				})
				return
			}

			idOwner, _ := db.GetOwner(params.Owner)
			institution := &data.Institution{
				Name:        name,
				Country:     country,
				Description: description,
				Email:       email,
				FlagDec:     params.Dec,
				State:       params.State,
				OwnerID:     idOwner.ID,
			}
			if params.Owner != "" {
				owner, err := db.GetOwner(params.Owner)
				if err == nil {
					institution.Owner = owner
				} else {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "El propietario no es válido",
						Code:    http.StatusBadRequest,
					})
					return
				}
			}
			db.UpdateInstitution(institution)
			PrtyParams, _ := events.PrettyParams(params)
			event := &events.EventLog{
				UserNickname: usuario.NickName,
				Resource:     "Instituciones",
				Event:        fmt.Sprintf("Se actualiza la institución %s en %s", name, country.Name),
				Params:       PrtyParams,
			}
			event.Write()
			updatedInstitution, _ := db.GetInstitution(name, country.ID)
			c.JSON(http.StatusOK, responseInstitution{Data: updatedInstitution})
		} else {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El nombre de la institución en el país indicado no está registrado",
				Code:    http.StatusBadRequest,
			})
			return
		}
	})
}

// @Summary update institution
// @Description update institution
// @Tags autentia institutions
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responseInstitution
// @failure 400 {object} MessageResponse
// @Router /autentia/institutions/{name} [delete]
// @param name path string true "name"
// @param country query string true "country"
func AutentiaDeleteInstitutionsRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Delete Institution
	router.DELETE("/institutions/:name", func(c *gin.Context) {

		name := c.Param("name")
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		countryName := strings.ToUpper(c.Query("country"))

		if countryName == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Parametros incorrectos",
				Code:    http.StatusBadRequest,
			})
			return
		}
		country, err := db.GetCountryByName(countryName)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		localInstitution, err := db.GetInstitution(name, country.ID)

		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if db.DeleteInstitution(localInstitution) != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al eliminar institución, intente nuevamente",
				Code:    http.StatusBadRequest,
			})
			return
		}
		PrtyParams, _ := events.PrettyParams(c.Params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Instituciones",
			Event:        fmt.Sprintf("Se elimina la institución %s", name),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, MessageResponse{Details: "Institución Eliminada", Code: http.StatusOK})
	})
}
