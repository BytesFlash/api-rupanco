package routes

import (
	"net/http"
	"path/filepath"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/imedcl/manager-api/pkg/data"
)

type paramsTrxHtml struct {
	NameHtml    string `form:"name_html" json:"name_html"`
	NameTrx     string `form:"name_trx" json:"name_trx"`
	Uri         string `form:"Uri" json:"Uri"`
	Description string `form:"description" json:"description"`
	Institution string `form:"institution" json:"institution"`
}

// Estructura para incluir la URL pre-firmada
type paramsTrxHtmlWithPresigned struct {
	NameHtml     string `json:"name_html"`
	NameTrx      string `json:"name_trx"`
	Uri          string `json:"Uri"`
	PresignedUri string `json:"presigned_uri"`
	Description  string `json:"description"`
	Institution  string `json:"institution"`
}

// @Summary post trxhtml
// @Description post trxhtml
// @Tags trxhtml
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 400 {object} MessageResponse
// @Router /device [post]
// @Param device body paramsTrxHtml true "device"
func TrxHtmlPostRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// POST trxhtml
	router.POST("/trxhtml", func(c *gin.Context) {
		var params paramsTrxHtml
		if err := c.ShouldBind(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Archivo HTML es requerido",
				Code:    http.StatusBadRequest,
			})
			return
		}
		_, nameHtmlErr := db.GetNameHtmlByName(params.NameHtml)

		if nameHtmlErr == nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Nombre de Html existe",
				Code:    http.StatusBadRequest,
			})
			return
		}

		_, nameTrxErr := db.GetNameTrxByName(params.NameTrx)

		if nameTrxErr == nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Nombre de Transacción existe",
				Code:    http.StatusBadRequest,
			})
			return
		}

		_, fileNameErr := db.GetFileNameByName(file.Filename)
		if fileNameErr == nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Nombre del archivo existe",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if filepath.Ext(file.Filename) != ".html" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Solo se permiten archivos HTML",
				Code:    http.StatusBadRequest,
			})
			return
		}

		fileContent, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, MessageResponse{
				Details: "Error al abrir el archivo",
				Code:    http.StatusInternalServerError,
			})
			return
		}
		defer fileContent.Close()

		s3Url, err := uploadFileToS3(fileContent, file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, MessageResponse{
				Details: cfg.BucketNameStorage(),
				Code:    http.StatusInternalServerError,
			})
			return
		}

		trxHtml := &data.TrxHtml{
			NameHtml:    params.NameHtml,
			NameTrx:     params.NameTrx,
			Uri:         s3Url,
			Description: params.Description,
			Institution: params.Institution,
		}
		db.CreateTrxHtml(trxHtml)

		c.JSON(http.StatusOK, MessageResponse{
			Details: "Html agregado con éxito",
			Code:    http.StatusOK,
		})
	})
}

// @Summary get device
// @Description get device
// @Tags device
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 404 {object} MessageResponse
// @Router /device [get]
func TrxHtmlGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//Get all Devices
	router.GET("/trxhtml", func(c *gin.Context) {
		trxHtml, err := db.GetAllTrxHtml()
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Println("error al obtener devices")
			return
		}
		c.JSON(http.StatusOK, trxHtml)
	})

}

// @Summary get device
// @Description get device
// @Tags device
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 404 {object} MessageResponse
// @Router /device [get]
func TrxHtmlGetRouteUri(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.GET("/trxhtml/:uri", func(c *gin.Context) {
		uri := c.Param("uri")
		if uri == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Error al obtener registros"})
			logrus.Println("error al obtener registros")
			return
		}

		s3Client, err := createS3Session()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando sesión de S3"})
			return
		}

		localFilePath := filepath.Join(uploadDir+"/AUTENTIA/", uri)

		erro := downloadFileFromS3(s3Client, uri, localFilePath)
		if erro != nil {
			logrus.Errorf("Error al descargar archivo desde S3: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al descargar archivo desde S3"})
			return
		}

		c.JSON(http.StatusOK, MessageResponse{
			Details: "Descargado",
			Code:    http.StatusOK,
		})
	})
}
