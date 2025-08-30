package routes

import (
	"fmt"
	"net/http"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"github.com/imedcl/manager-api/pkg/mail"
	"github.com/imedcl/manager-api/pkg/services"
	"github.com/sirupsen/logrus"
)

type resourceUserCreateParams struct {
	Country       string `form:"Country"`
	Usuario       string `form:"Usuario"`
	Recurso       string `form:"Recurso"`
	Instituciones string `form:"Instituciones"`
	Clave         string `form:"Clave"`
	Opcion1       string `form:"Opcion1"`
	Opcion2       string `form:"Opcion2"`
	Opcion3       string `form:"Opcion3"`
	Host          string `form:"Host"`
}

type resourceUserListParams struct {
	Country string `form:"Country"`
	Usuario string `form:"Usuario"`
	Recurso string `form:"Recurso"`
}
type resourceUserClaveParams struct {
	Country string `form:"Country"`
	Usuario string `form:"Usuario"`
	Recurso string `form:"Recurso"`
	Opcion  string `form:"Opcion"`
	Email   string `form:"Email"`
}
type paramsServiceAndResources struct {
	Name    string `json:"name"`
	Service string `json:"service"`
}

type paramsServices struct {
	Services []struct {
		Name string `json:"name"`
	} `json:"services"`
}

type responseResource struct {
	Data interface{} `json:"data"`
}

// @Summary Create a resource user
// @Description Create a new resource user
// @Tags Resource
// @Accept json
// @Produce json
// @Param resourceUser body resourceUserCreateParams true "Resource User Parameters"
// @Success 200 {object} responseResource
// @Failure 400 {object} MessageResponse
// @Router autentia/resource/user [post]
func AutentiaPostResourceUserRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Create Resource User
	router.POST("/resource/user", func(c *gin.Context) {

		var params resourceUserCreateParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		list := services.CreateResource(
			params.Country,
			params.Usuario,
			params.Recurso,
			params.Instituciones,
			params.Clave,
			params.Opcion1,
			params.Opcion2,
			params.Opcion3,
			params.Host,
		)

		PrtyParams, _ := events.PrettyParams(list)
		event := &events.EventLog{
			UserNickname: currentUser.NickName,
			Resource:     "Recurso Autentia",
			Event:        fmt.Sprintf("Se ha creado el recurso %s, en la institución, %s", params.Recurso, params.Instituciones),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, responseResource{Data: list.Body.CResultado})
	})
}

// @Summary List resources
// @Description List all resources based on given parameters
// @Tags Resource
// @Accept json
// @Produce json
// @Param resourceListParams body resourceUserListParams true "Resource List Parameters"
// @Success 200 {object} responseResource
// @Failure 400 {object} MessageResponse
// @Router autentia/resource/list [get]
func AutentiaGetResourceListRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Create Resource User
	router.POST("/resource/list", func(c *gin.Context) {

		var params resourceUserListParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		resource := services.ListResource(
			params.Country,
			params.Usuario,
			params.Recurso)

		c.JSON(http.StatusOK, responseResource{Data: resource.Body.ListResp})
	})
}

// @Summary Send resource key
// @Description Sends the resource key by email to the specified user.
// @Tags Resource
// @Accept json
// @Produce json
// @Param resourceUser body resourceUserClaveParams true "Resource user parameters"
// @Success 200 {object} responseResource "Key sent successfully"
// @Failure 400 {object} MessageResponse "Invalid request or empty key"
// @Router /autentia/resource/clave [post]
func AutentiaGetResourceClaveRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Send resource
	router.POST("/resource/clave", func(c *gin.Context) {

		var params resourceUserClaveParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		resourceKey := cfg.KeyResouce()

		resource := services.SendMessage(
			params.Country,
			params.Usuario,
			params.Recurso,
			resourceKey,
		)
		if resource.Body.ListResp.Recs[0].Clave == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Clave en blanco",
				Code:    http.StatusBadRequest,
			})
			return
		}
		mailParams := mail.ResourceClaveParams{
			Recurso: params.Recurso,
			Email:   params.Email,
			Pass:    resource.Body.ListResp.Recs[0].Clave,
		}
		mail.SendPass(currentUser, mailParams)
		PrtyParams, _ := events.PrettyParams(c.Params)
		event := &events.EventLog{
			UserNickname: currentUser.NickName,
			Resource:     "Recurso Autentia",
			Event:        fmt.Sprintf("Se ha enviado el recurso %s, en al correo, %s", params.Recurso, params.Email),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, responseResource{"Correo enviado exitosamente"})
	})
}

// @Summary Create services
// @Description Create new services
// @Tags Service
// @Accept json
// @Produce json
// @Param services body paramsServices true "Services Parameters"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router autentia/services [post]
func AutentiaPostService(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//POST Service
	router.POST("/services", func(c *gin.Context) {
		var params paramsServices

		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		for _, service := range params.Services {
			_, err := db.GetAutentiaServiceByName(service.Name)
			if err == nil {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: fmt.Sprintf("Servicio '%s' ya existe", service.Name),
					Code:    http.StatusBadRequest,
				})
				return
			}

			nameAutenti := &data.AutentiaService{
				Name: service.Name,
			}
			_, err = db.CreateAutentiaService(nameAutenti)
			if err != nil {
				logrus.Print("Error al crear servicio", err.Error())
				c.JSON(http.StatusInternalServerError, MessageResponse{
					Details: fmt.Sprintf("Error al crear el servicio '%s'", service.Name),
					Code:    http.StatusInternalServerError,
				})
				return
			}

			PrtyParams, _ := events.PrettyParams(service)
			event := &events.EventLog{
				UserNickname: currentUser.NickName,
				Resource:     "Autentia Resource",
				Event:        fmt.Sprintf("Se crea el servicio %s", service.Name),
				Params:       PrtyParams,
			}
			event.Write()
		}

		c.JSON(http.StatusOK, MessageResponse{
			Details: "Servicios autentia creados con éxito",
			Code:    http.StatusOK,
		})
	})
}

// @Summary Create resources for a service
// @Description Create new resources for a specific service
// @Tags Resource
// @Accept json
// @Produce json
// @Param serviceResources body []paramsServiceAndResources true "Service and Resources Parameters"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router autentia/resource [post]
func AutentiaPostResource(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// POST resource
	router.POST("/resource", func(c *gin.Context) {
		var resources []paramsServiceAndResources

		if err := c.ShouldBindJSON(&resources); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		for _, params := range resources {
			service, errSer := db.GetAutentiaServiceByName(params.Service)
			if errSer != nil {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "Servicio no existe",
					Code:    http.StatusBadRequest,
				})
				return
			}

			_, err := db.GetAutentiaResourceByNameAndID(params.Name, service.ID)
			if err == nil {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "Recurso ya existe",
					Code:    http.StatusBadRequest,
				})
				return
			}

			nameAutenti := &data.AutentiaResource{
				Name:      params.Name,
				ServiceID: service.ID,
			}

			if err := db.CreateAutentiaResource(nameAutenti); err != nil {
				c.JSON(http.StatusInternalServerError, MessageResponse{
					Details: "Error al crear el recurso",
					Code:    http.StatusInternalServerError,
				})
				return
			}

			PrtyParams, _ := events.PrettyParams(params)
			event := &events.EventLog{
				UserNickname: currentUser.NickName,
				Resource:     "Autentia Resource",
				Event:        fmt.Sprintf("Se crea el recurso %s", params.Name),
				Params:       PrtyParams,
			}
			event.Write()
		}

		c.JSON(http.StatusOK, MessageResponse{
			Details: "Recursos autentia creados con éxito",
			Code:    http.StatusOK,
		})
	})
}

// @Summary Get all services
// @Description Retrieve a list of all services
// @Tags Service
// @Produce json
// @Success 200 {array} data.AutentiaService
// @Failure 404 {object} MessageResponse
// @Router autentia/service [get]
func AutentiaServiceGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//get service
	router.GET("/service", func(c *gin.Context) {

		role, err := db.GetAllService()
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Println("error al obtener los servicios")
			return
		}
		c.JSON(http.StatusOK, role)
	})
}

// @Summary Get a resource by ID
// @Description Retrieve a specific resource by its ID
// @Tags Resource
// @Produce json
// @Param id path string true "Resource ID"
// @Success 200 {object} data.AutentiaResource
// @Failure 404 {object} MessageResponse
// @Router autentia/resource/{id} [get]
func AutentiaResourceGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//get service
	router.GET("/resource/:id", func(c *gin.Context) {
		idservice := c.Param("id")
		serviName, _ := db.GetAutentiaServiceByName(idservice)
		service, err := db.GetAutentiaResourceByName(serviName.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Println("error al obtener los recursos")
			return
		}
		c.JSON(http.StatusOK, service)
	})
}
