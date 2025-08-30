package routes

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/charmap"
)

type ubicationParams struct {
	Name        string `form:"name" binding:"required"`
	Country     string `form:"country" binding:"required"`
	Institution string `form:"institution" binding:"required"`
	Code        string `form:"code" binding:"required"`
	Description string `form:"description"`
	Entitity    string `form:"entitity"`
	Address     string `form:"address"`
	State       string `form:"state"`
}

type batchUbicationParams struct {
	Ubications  *multipart.FileHeader `form:"ubications"`
	Country     string                `form:"country" binding:"required"`
	Institution string                `form:"institution"`
	Delimiter   string                `form:"delimiter"`
}

type ubicationsResponse struct {
	Data *[]data.Ubication `json:"data"`
}

type ubicationResponse struct {
	Data *data.Ubication `json:"data"`
}

type updateUbicationParams struct {
	Name        string `form:"name" binding:"required"`
	Country     string `form:"country" binding:"required"`
	Institution string `form:"institution" binding:"required"`
	Description string `form:"description"`
	Entitity    string `form:"entitity"`
	Address     string `form:"address"`
	State       string `form:"state"`
}

type deleteUbicationParams struct {
	Country     string `form:"country" binding:"required"`
	Institution string `form:"institution" binding:"required"`
}

type updateResponse struct {
	Data bool `json:"data"`
}

type deleteResponse struct {
	Data bool `json:"data"`
}

func UbicationsRoutes(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {

	// Get Ubications List
	router.GET("/ubications", func(c *gin.Context) {
		country, _ := c.GetQuery("country")
		institutionName, _ := c.GetQuery("institution")
		if country == "" || institutionName == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Bad Request",
				Code:    http.StatusBadRequest,
			})
			return
		}
		countryData, countryErr := db.GetCountryByName(country)
		if countryErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: countryErr.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		institution, err := db.GetInstitution(institutionName, countryData.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		ubications := db.GetUbications(institution)
		c.JSON(http.StatusOK, ubicationsResponse{Data: ubications})
	})

	// Get Ubication
	router.GET("/ubications/:code", func(c *gin.Context) {
		code := c.Param("code")
		country, _ := c.GetQuery("country")
		institutionName, _ := c.GetQuery("institution")
		if country == "" || institutionName == "" || code == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Bad Request",
				Code:    http.StatusBadRequest,
			})
			return
		}
		institution, err := db.GetInstitution(institutionName, country)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		if db.ExistsUbication(code, institution) {
			ubication, _ := db.GetUbication(code, institution)
			c.JSON(http.StatusOK, ubicationResponse{Data: ubication})
		} else {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "No encontrada",
				Code:    http.StatusNotFound,
			})
		}
	})

	// Create Batch Ubications
	router.POST("/ubications/batch", func(c *gin.Context) {
		var params batchUbicationParams
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if err := c.ShouldBind(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		country, countryErr := db.GetCountryByName(params.Country)
		if countryErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: countryErr.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		institution, err := db.GetInstitution(params.Institution, country.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		file, err := params.Ubications.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		defer file.Close()
		detector := chardet.NewTextDetector()
		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		result, err := detector.DetectBest(buf.Bytes())
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		var csvReader *csv.Reader
		if result.Charset == "UTF-8" {
			csvReader = csv.NewReader(buf)
		} else {
			csvReader = csv.NewReader(charmap.ISO8859_1.NewDecoder().Reader(buf))

		}
		if len(params.Delimiter) == 1 {
			delimiter := []rune(params.Delimiter)
			csvReader.Comma = delimiter[0]
		}
		fileContent, err := csvReader.ReadAll()
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(fileContent) <= 1 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El archivo debe tener al menos una linea",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(fileContent[1]) != 6 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "La cantidad de columnas del archivo no coinciden",
				Code:    http.StatusBadRequest,
			})
			return
		}
		batchSize := 1000
		var limit int
		for i := 1; i < len(fileContent); i = i + batchSize {
			limit = i + batchSize
			if len(fileContent) < limit {
				limit = len(fileContent)
			}

			for in, line := range fileContent[i:limit] {
				countstr := strconv.Itoa(in + 2)
				if line[0] == "" {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "No se encontró el código de lugar en la linea " + countstr + ".",
						Code:    http.StatusBadRequest,
					})
					return
				}
				codeRegex := regexp.MustCompile(`^[A-Za-z0-9-\s]+$`)
				if !codeRegex.MatchString(line[0]) {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "Hay un error en el formato del código de lugar en la linea " + countstr + ".",
						Code:    http.StatusBadRequest,
					})
					return
				}
				if line[1] == "" {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "No se encontró la descripción de lugar en la linea " + countstr + ".",
						Code:    http.StatusBadRequest,
					})
					return
				}
				descriptionRegex := regexp.MustCompile(`^[ª”´\x60':#\/a-zA-ZÀ-ÿñÑ()0-9&°º.–\-  ]*$`)
				if !descriptionRegex.MatchString(line[1]) {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "Hay un error en el formato de la descripción de lugar en la linea " + countstr + ".",
						Code:    http.StatusBadRequest,
					})
					return
				}
				if line[2] == "" {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "No se encontró la entidad de lugar en la linea " + countstr + ".",
						Code:    http.StatusBadRequest,
					})
					return
				}
				if line[3] == "" {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "No se encontró el nombre de lugar en la linea " + countstr + ".",
						Code:    http.StatusBadRequest,
					})
					return
				}
				nameRegex := regexp.MustCompile(`^[ª”´\x60':#\/a-zA-ZÀ-ÿñÑ()0-9&°º.–\-  ]*$`)
				if !nameRegex.MatchString(line[3]) {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "Hay un error en el formato del nombre de lugar en la linea " + countstr + ".",
						Code:    http.StatusBadRequest,
					})
					return
				}
				if line[4] == "" {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "No se encontró la dirección de lugar en la linea " + countstr + ".",
						Code:    http.StatusBadRequest,
					})
					return
				}
				if line[5] == "" {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "No se encontró la comuna de lugar en la linea " + countstr + ".",
						Code:    http.StatusBadRequest,
					})
					return
				}
				distritRegex := regexp.MustCompile(`^[ª”´\x60':#\/a-zA-ZÀ-ÿñÑ()0-9&°º.–\-  ]*$`)
				if !distritRegex.MatchString(line[5]) {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "Hay un error en el formato de la comuna de lugar en la linea " + countstr + ".",
						Code:    http.StatusBadRequest,
					})
					return
				}
				code := line[0]
				if db.ExistsUbication(code, institution) {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "La ubicación " + code + " ya se encuentra registrada en la institución " + institution.Name + ", Revisar la linea " + countstr + ".",
						Code:    http.StatusBadRequest,
					})
					return
				}
			}

		}

		var flag bool = false
		flag = setBatchUbications(fileContent[1:limit], institution, usuario.NickName)

		if flag {
			var response responseSensorAddBatch
			response.Data.Status = true
			c.JSON(http.StatusOK, response)
		}

	})

	// Create ubication
	router.POST("/ubications", func(c *gin.Context) {
		var params ubicationParams
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
		institution, err := db.GetInstitution(params.Institution, params.Country)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		var ubication = &data.Ubication{
			Code:        params.Code,
			Description: params.Description,
			Entitity:    params.Entitity,
			Name:        params.Name,
			Address:     params.Address,
			State:       params.State,
			Institution: institution,
		}
		if db.ExistsUbication(ubication.Code, institution) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "La ubicación ya se encuentra registrada en esta institución",
				Code:    http.StatusBadRequest,
			})
			return
		}
		db.CreateUbication(ubication)
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Ubicaciones",
			Event:        fmt.Sprintf("La ubicación %s fue registrada correctamente para la institución %s - %s", ubication.Code, ubication.Institution.Name, ubication.Institution.Country.Name),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusCreated, ubicationResponse{Data: ubication})
	})

	// Update ubication
	router.PUT("/ubications/:code", func(c *gin.Context) {
		code := c.Param("code")
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if code == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Parametros incompletos",
				Code:    http.StatusBadRequest,
			})
			return
		}
		var params updateUbicationParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		institution, err := db.GetInstitution(params.Institution, params.Country)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		var ubication = &data.Ubication{
			Description: params.Description,
			Entitity:    params.Entitity,
			Name:        params.Name,
			Address:     params.Address,
			State:       params.State,
		}
		statusResponse := db.UpdateUbication(code, ubication, institution)
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Ubicaciones",
			Event:        fmt.Sprintf("La ubicación %s fue actualizada correctamente para la institución %s - %s", ubication.Code, ubication.Institution.Name, ubication.Institution.Country.Name),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, updateResponse{Data: statusResponse})
	})

	// Delete Ubication
	router.DELETE("/ubications/:code", func(c *gin.Context) {
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		code := c.Param("code")
		if code == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Parametros faltantes",
				Code:    http.StatusBadRequest,
			})
			return
		}
		var params deleteUbicationParams
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		institution, err := db.GetInstitution(params.Institution, params.Country)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		statusResponse, err := db.RemoveUbication(code, institution)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Ubicaciones",
			Event:        fmt.Sprintf("La ubicación %s fue eliminada correctamente para la institución %s - %s", code, institution.Name, institution.Country.Name),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, deleteResponse{Data: statusResponse})
	})
}

func setBatchUbications(lines [][]string, institution *data.Institution, nickname string) bool {

	for _, line := range lines {
		if len(line) == 6 {
			var ubication = &data.Ubication{
				Code:        line[0],
				Description: line[1],
				Entitity:    line[2],
				Name:        line[3],
				Address:     line[4],
				State:       line[5],
				Institution: institution,
			}
			PrtyParams, _ := events.PrettyParams(ubication)
			event := &events.EventLog{
				UserNickname: nickname,
				Resource:     "Ubicaciones",
				Params:       PrtyParams,
			}
			if !db.ExistsUbication(ubication.Code, institution) {
				db.CreateUbication(ubication)
				event.Event = fmt.Sprintf("La ubicación %s fue registrada correctamente para la institución %s - %s en carga masiva", ubication.Code, institution.Name, institution.Country.Name)
			} else {
				db.UpdateUbication(ubication.Code, ubication, institution)
				event.Event = fmt.Sprintf("La ubicación %s fue actualizada correctamente para la institución %s - %s en carga masiva", ubication.Code, institution.Name, institution.Country.Name)
			}
			event.Write()
		}
	}
	return true
}
