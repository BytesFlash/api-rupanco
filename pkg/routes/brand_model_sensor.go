package routes

import (
	"net/http"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"

	"github.com/imedcl/manager-api/pkg/data"
)

type responseBrandList struct {
	Data []*data.Brand `json:"data"`
}

type responseBrand struct {
	Data data.Brand `json:"data"`
}

type responseModelList struct {
	Data []*data.Model `json:"data"`
}
type BrandAddParams struct {
	Name string `form:"name"`
}
type ModelAddParams struct {
	Name  string `form:"name"`
	Brand string `form:"brand"`
}
type MessageResponseBrand struct {
	Details string `json:"details"`
	Code    int    `json:"code"`
}

// @Summary Get all brands
// @Description Get a list of all brands from the database
// @Tags brands
// @Accept json
// @Produce json
// @Success 200 {object} responseBrandList
// @Failure 400 {object} MessageResponse
// @Router /brands [get]
func GetAllBrandRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Obtener todas las marcas
	router.GET("/brands", func(c *gin.Context) {
		// Llamar al método ListAllBrand
		brands, err := db.ListAllBrand()

		if err != nil {
			// Si hay error, devolver una respuesta con código 400
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al obtener las marcas",
				Code:    http.StatusBadRequest,
			})
			return
		}

		// Si la consulta es exitosa, devolver las marcas en formato JSON
		c.JSON(http.StatusOK, responseBrandList{
			Data: brands,
		})
	})
}

// @Summary Get all models for a specific brand
// @Description Get a list of models for a specific brand from the database
// @Tags models
// @Accept json
// @Produce json
// @Param brandName path string true "Brand Name"
// @Success 200 {object} responseModelList
// @Failure 400 {object} MessageResponse
// @Router /brands/{brandName}/models [get]
func GetModelsByBrandRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Obtener los modelos para una marca específica
	router.GET("/models/:brandName", func(c *gin.Context) {
		brandName := c.Param("brandName") // Obtener el nombre de la marca desde el parámetro de la URL

		// Llamar al método ListModelsByBrand
		brandId, err := db.GetBrandByName(brandName)
		if err != nil {
			// Si hay error, devolver una respuesta con código 400
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al obtener la marca",
				Code:    http.StatusBadRequest,
			})
			return
		}

		models, err := db.ListModelsByBrand(brandId.ID)

		if err != nil {
			// Si hay error, devolver una respuesta con código 400
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al obtener los modelos",
				Code:    http.StatusBadRequest,
			})
			return
		}

		// Si la consulta es exitosa, devolver los modelos en formato JSON
		c.JSON(http.StatusOK, responseModelList{
			Data: models,
		})
	})
}

// @Summary Create a new brand
// @Description Add a new brand to the database
// @Tags brands
// @Accept json
// @Produce json
// @Param params body BrandAddParams true "Brand to create"
// @Success 200 {object} MessageResponseBrand
// @Failure 400 {object} MessageResponse
// @Router /brands [post]
func CreateBrandRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.POST("/brand", func(c *gin.Context) {
		var params BrandAddParams

		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		_, errBrand := db.GetBrandByName(params.Name)

		if errBrand == nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: ("La marca ya existe"),
				Code:    http.StatusNotFound,
			})
			return
		}

		brand := &data.Brand{
			Name: params.Name,
		}

		db.CreateBrand(brand)

		c.JSON(http.StatusOK, MessageResponseBrand{
			Details: "Marca creada con éxito",
			Code:    http.StatusOK,
		})
	})
}

// @Summary Update a brand
// @Description Update an existing brand by ID
// @Tags brands
// @Accept json
// @Produce json
// @Param id path string true "Brand ID"
// @Param brand body Brand true "Brand data"
// @Success 200 {object} responseBrand
// @Failure 400 {object} MessageResponse
// @Router /brands/{id} [put]
func UpdateBrandRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.PUT("/brand/:id", func(c *gin.Context) {
		id := c.Param("id")
		var params BrandAddParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Datos inválidos para actualizar",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if id == "" {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "ID no puede ser vacío",
				Code:    http.StatusNotFound,
			})
			return
		}

		if params.Name == "" {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "Nombre no puede ser vacío",
				Code:    http.StatusNotFound,
			})
			return
		}

		// Verificar que la marca exista
		_, err := db.GetBrandById(id)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "La marca no existe",
				Code:    http.StatusNotFound,
			})
			return
		}
		_, errName := db.GetBrandByName(params.Name)
		if errName == nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "El nombre de la marca ya existe",
				Code:    http.StatusNotFound,
			})
			return
		}

		// Enlazar nuevo contenido del body

		brand := &data.Brand{
			ID:   id,
			Name: params.Name,
		}

		// Actualizar la marca
		if err := db.UpdateBrand(brand); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al actualizar la marca",
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusOK, MessageResponse{
			Details: "Actualizado correctamente",
			Code:    http.StatusOK,
		})
	})

}

// @Summary Create a new model
// @Description Add a new model to the database
// @Tags model
// @Accept json
// @Produce json
// @Param params body ModelAddParams true "Model to create"
// @Success 200 {object} MessageResponseBrand
// @Failure 400 {object} MessageResponse
// @Router /model [post]
func CreateModeldRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.POST("/model", func(c *gin.Context) {
		var params ModelAddParams

		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if params.Name == "" {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: ("El modelo no puede estar vacio"),
				Code:    http.StatusNotFound,
			})
			return
		}

		if params.Brand == "" {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: ("La marca no puede estar vacia"),
				Code:    http.StatusNotFound,
			})
			return
		}

		Idbrand, errBrand := db.GetBrandByName(params.Brand)

		if errBrand != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: ("La marca no existe"),
				Code:    http.StatusNotFound,
			})
			return
		}

		_, err := db.GetModelByNameBrandByID(params.Name, Idbrand.ID)

		if err == nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: ("La asociación ya existe"),
				Code:    http.StatusNotFound,
			})
			return
		}

		model := &data.Model{
			Name:    params.Name,
			BrandID: Idbrand.ID,
		}

		db.CreateModel(model)

		c.JSON(http.StatusOK, MessageResponseBrand{
			Details: "Modelo creado con éxito",
			Code:    http.StatusOK,
		})
	})
}
