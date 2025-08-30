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
	"strings"
	"time"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/charmap"

	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"github.com/imedcl/manager-api/pkg/mail"
	"github.com/imedcl/manager-api/pkg/services"
)

type responseSensor struct {
	Data *data.Sensor `json:"data"`
}

type responseSensorExport struct {
	Data []*data.Sensor `json:"data"`
}

type responseSensorList struct {
	Data  []*data.Sensor `json:"data"`
	Total int64          `json:"total"`
}

type responseSensorAdd struct {
	Data struct {
		Status bool `json:"status"`
	} `json:"data"`
}

type responseSensorAddBatch struct {
	//Data data.Sensor `json:"data"`
	Data struct {
		Status bool `json:"status"`
	} `json:"data"`
}

type sensorAddParams struct {
	Profile     string `form:"profile"`
	Code        string `form:"code"`
	External    string `form:"external"`
	Location    string `form:"location"`
	Country     string `form:"country"`
	Institution string `form:"institution"`
	Logon       string `form:"logon"`
	State       string `form:"state"`
	Active      bool   `form:"active"`
	Brand       string `form:"brand"`
	Model       string `form:"model"`
}

type sensorUpdateParams struct {
	Profile     string `form:"profile"`
	Code        string `form:"code"`
	External    string `form:"external"`
	Location    string `form:"location"`
	Country     string `form:"country"`
	Institution string `form:"institution"`
	Logon       string `form:"logon"`
	State       string `form:"state"`
	Active      bool   `form:"active"`
	DniEntity   string `form:"entity"`
	Brand       string `form:"brand"`
	Model       string `form:"model"`
}

type sensorsBatchAddParams struct {
	Country   string                `form:"country"`
	ProfileId string                `form:"profile"`
	Sensors   *multipart.FileHeader `form:"sensors"`
	Delimiter string                `form:"delimiter"`
	Active    bool                  `form:"active"`
	DniEntity string                `form:"entity"`
}

func dateFormat(date string) (newDate string) {
	formattedDate, dateError := time.Parse("2006-01-02T15:04:05.00000-04:00", date)
	if dateError != nil {
		formattedDate, dateError = time.Parse("2006-01-02 15:04:05", date)
		if dateError != nil {
			formattedDate, dateError = time.Parse("2006-01-02", date)
			if dateError != nil {
				return
			}
		}
	}
	dateSplit := strings.Split(strings.Split(formattedDate.String(), " ")[0], "-")
	if len(dateSplit) == 3 {
		if len(dateSplit[0]) == 4 {
			newDate = fmt.Sprintf("%s-%s-%s", dateSplit[0], dateSplit[1], dateSplit[2])
		} else if len(dateSplit[2]) == 4 {
			newDate = fmt.Sprintf("%s-%s-%s", dateSplit[2], dateSplit[1], dateSplit[0])
		}
	}
	return
}

//SENSOR//

// @Summary get autentia sensor
// @Description get autentia sensor
// @Tags autentia sensors
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responseSensorList
// @failure 400 {object} MessageResponse
// @Router /autentia/sensors [get]
// @param country query string true "country"
// @param institution query string true "institution"
func AutentiaGetSensorsRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get Sensors

	router.GET("/sensors", func(c *gin.Context) {
		country := c.Query("country")
		institution := c.Query("institution")

		limit, _ := strconv.Atoi(c.Query("limit"))
		offset, _ := strconv.Atoi(c.Query("offset"))
		if country == "" || institution == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Parametros incorrectos",
				Code:    http.StatusBadRequest,
			})
			return
		}
		sensors := db.GetSensors(institution, country, limit, offset)
		total := db.GetSensorNumber(institution, country)
		c.JSON(http.StatusOK, responseSensorList{Data: sensors, Total: total})
	})
}

// @Summary get autentia sensor owner
// @Description get autentia sensor owner
// @Tags autentia sensors
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responseSensorExport
// @failure 400 {object} MessageResponse
// @Router /autentia/sensors/owner/{owner} [get]
// @param owner path string true "owner"
// @param country query string true "country"
func AutentiaGetSensorsOwnerRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get Sensors By Owner
	router.GET("/sensors/owner/:owner", func(c *gin.Context) {
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		owner := c.Param("owner")
		country := c.Query("country")
		if country == "" || owner == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Parametros incorrectos",
				Code:    http.StatusBadRequest,
			})
			return
		}
		sensors := db.GetSensorsByOwner(country, owner)
		PrtyParams, _ := events.PrettyParams(c.Params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Descarga Sensores",
			Event:        fmt.Sprintf("Se procesa la descarga de sensores de %s en %s", owner, country),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, responseSensorExport{Data: sensors})
	})
}

// @Summary get autentia sensor code
// @Description get autentia sensor code
// @Tags autentia sensors
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responseSensor
// @failure 404 {object} MessageResponse
// @Router /autentia/sensors/{code} [get]
// @param code path string true "code"
// @param country query string true "country"
// @param technology query string true "technology"
func AutentiaGetSensorsCodeRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get Sensor
	router.GET("/sensors/:code", func(c *gin.Context) {
		code := c.Param("code")
		country := c.Query("country")
		technology := c.Query("technology")
		// Add institution params for validate permissions

		if country == "" || technology == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Parametros incorrectos",
				Code:    http.StatusBadRequest,
			})
			return
		}
		autentiaSensor := services.GetSensor(country, code, technology)
		if autentiaSensor.Code == "" {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "Sensor no registrado",
				Code:    http.StatusNotFound,
			})
			return
		}
		var sensor *data.Sensor
		if db.ExistsSensor(autentiaSensor.Code) {
			sensor = db.GetSensor(autentiaSensor.Code)
			sensor.Code = autentiaSensor.Code
			sensor.Institution = autentiaSensor.Institution
			sensor.Location = autentiaSensor.Location
			sensor.LogonType = autentiaSensor.LogonType
			sensor.State = autentiaSensor.State
			sensor.Technology = autentiaSensor.Technology
			sensor.Brand = autentiaSensor.Brand
			sensor.Model = autentiaSensor.Model
			_ = db.UpdateSensor(code, sensor)
		} else {

			sensor = &data.Sensor{
				Code:         autentiaSensor.Code,
				Institution:  autentiaSensor.Institution,
				Country:      country,
				Location:     autentiaSensor.Location,
				ExternalCode: autentiaSensor.External,
				LogonType:    autentiaSensor.LogonType,
				State:        autentiaSensor.State,
				Technology:   autentiaSensor.Technology,
				Brand:        autentiaSensor.Brand,
				Model:        autentiaSensor.Model,
			}

			db.CreateSensor(sensor)
		}
		c.JSON(http.StatusOK, responseSensor{Data: sensor})
	})
}

// @Summary post autentia sensor
// @Description post autentia sensor
// @Tags autentia sensors
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 "Sensor registrado correctamente"
// @failure 400 {object} MessageResponse
// @Router /autentia/sensors [post]
// @Param sensor body sensorAddParams true "sensor"
func AutentiaPostSensorRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Create/update Sensor
	router.POST("/sensors", func(c *gin.Context) {
		var params sensorAddParams
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
		var userData, _ = db.GetUserByID(params.Profile)

		codeRegex := regexp.MustCompile(`^[{a-zA-Z0-9-}]{4,42}$`)
		if !codeRegex.MatchString(params.Code) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Serial incorrecto, debe tener entre 4 y 40 carácteres, además de las llaves {}",
				Code:    http.StatusBadRequest,
			})
			return
		}

		count, _ := db.GetCountryByName(params.Country)
		insti, errInst := db.GetInstitution(params.Institution, count.ID)
		if errInst != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: ("La institución no existe"),
				Code:    http.StatusNotFound,
			})
			return
		} else {
			if params.Active == false {
				_, errormatch := db.GetUserAndInstitution(params.Profile, insti.ID)

				if errormatch != nil {
					c.JSON(http.StatusNotFound, MessageResponse{
						Details: ("No cuenta con permisos en esta institución"),
						Code:    http.StatusNotFound,
					})
					return
				}
			}
		}

		location := regexp.MustCompile(`^[{0-9}]{1,150}$`)
		if !location.MatchString(params.Location) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Debe ingresar código de lugar correcto",
				Code:    http.StatusBadRequest,
			})
			return
		}

		logonType, _ := strconv.Atoi(params.Logon)
		if logonType != 1 && logonType != 4 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El tipo logon no es válido",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if params.Brand != "" {
			_, errBrand := db.GetBrandByName(params.Brand)
			if errBrand != nil {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "Marca no existe",
					Code:    http.StatusBadRequest,
				})
				return
			}
			if params.Model != "" {
				_, errModel := db.GetModelByName(params.Model)
				if errModel != nil {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "Modelo no existe",
						Code:    http.StatusBadRequest,
					})
					return
				}
			}
		}

		state, _ := strconv.Atoi(params.State)

		technologies := []string{"uareu", "uareu-gold"}

		for _, technolgy := range technologies {
			//create sensor cgi
			success, errSen := services.SetSensor(
				params.Country,
				params.Code,
				params.External,
				params.Institution,
				technolgy,
				params.Location,
				state,
				logonType,
				params.Brand,
				params.Model,
			)
			if success != true {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: fmt.Sprintf("%s", errSen),
					Code:    http.StatusBadRequest,
				})
				return
			}
			//create sensor
			sensor := &data.Sensor{
				Code:         params.Code,
				Institution:  params.Institution,
				Country:      params.Country,
				Location:     params.Location,
				ExternalCode: params.External,
				LogonType:    logonType,
				Technology:   technolgy,
				State:        state,
				Brand:        params.Brand,
				Model:        params.Model,
			}
			db.CreateSensor(sensor)

			//create event sensor
			var eventSensor = &data.EventSensor{
				Code:        params.Code,
				Institution: params.Institution,
				User:        userData.NickName,
				Action:      "Registrar",
				Glosa:       "Sensor se registro correctamente",
			}
			db.CreateEventSensor(eventSensor)

			PrtyParams, _ := events.PrettyParams(params)
			event := &events.EventLog{
				UserNickname: usuario.NickName,
				Resource:     "Sensor",
				Event:        fmt.Sprintf("El sensor %s fue registrado correctamente", sensor.Code),
				Params:       PrtyParams,
				Sensor:       sensor.Code,
			}
			event.Write()

		}

		var response responseSensorAddBatch
		response.Data.Status = true
		c.JSON(http.StatusOK, MessageResponse{
			Details: ("Sensor registrado correctamente"),
			Code:    http.StatusOK,
		})
	})
}

// @Summary put autentia sensor
// @Description put autentia sensor
// @Tags autentia sensors
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 "Sensor actualizado correctamente"
// @failure 400 {object} MessageResponse
// @Router /autentia/sensors/update [put]
// @Param sensor body sensorUpdateParams true "sensor"
func AutentiaPutSensorRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Update Sensor
	router.PUT("/sensors/update", func(c *gin.Context) {
		var params sensorUpdateParams
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
		var userData, _ = db.GetUserByID(params.Profile)

		codeRegex := regexp.MustCompile(`^[{a-zA-Z0-9-}]{4,42}$`)
		if !codeRegex.MatchString(params.Code) {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Serial incorrecto, debe tener entre 4 y 40 carácteres, además de las llaves {}",
				Code:    http.StatusBadRequest,
			})
			return
		}

		count, _ := db.GetCountryByName(params.Country)
		insti, errInst := db.GetInstitution(params.Institution, count.ID)
		if errInst != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: ("La institución no existe"),
				Code:    http.StatusNotFound,
			})
			return
		} else {
			if params.Active == false {
				_, errormatch := db.GetUserAndInstitution(params.Profile, insti.ID)

				if errormatch != nil {
					c.JSON(http.StatusNotFound, MessageResponse{
						Details: ("No cuenta con permisos en esta institución"),
						Code:    http.StatusNotFound,
					})
					return
				}
			}
		}

		logonType, _ := strconv.Atoi(params.Logon)
		if logonType != 1 && logonType != 4 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "El tipo logon no es válido",
				Code:    http.StatusBadRequest,
			})
			return
		}

		if params.Brand != "" {
			_, errBrand := db.GetBrandByName(params.Brand)
			if errBrand != nil {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "Marca no existe",
					Code:    http.StatusBadRequest,
				})
				return
			}
			if params.Model != "" {
				_, errModel := db.GetModelByName(params.Model)
				if errModel != nil {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "Modelo no existe",
						Code:    http.StatusBadRequest,
					})
					return
				}
			}
		}

		state, _ := strconv.Atoi(params.State)
		technologies := []string{"uareu", "uareu-gold"}

		for _, technolgy := range technologies {

			//create sensor cgi
			success, errSen := services.SetSensor(
				params.Country,
				params.Code,
				params.External,
				params.Institution,
				technolgy,
				params.Location,
				state,
				logonType,
				params.Brand,
				params.Model,
			)
			if success != true {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: fmt.Sprintf("%s", errSen),
					Code:    http.StatusBadRequest,
				})
				return
			}
			//create sensor
			sensor := &data.Sensor{
				Code:         params.Code,
				Institution:  params.Institution,
				Country:      params.Country,
				Location:     params.Location,
				ExternalCode: params.External,
				LogonType:    logonType,
				Technology:   technolgy,
				State:        state,
				Brand:        params.Brand,
				Model:        params.Model,
			}
			db.UpdateSensor(params.Code, sensor)

			//create event sensor
			var eventSensor = &data.EventSensor{
				Code:        params.Code,
				Institution: params.Institution,
				User:        userData.NickName,
				Action:      "Actualizar",
				Glosa:       "Sensor se actualizó correctamente",
			}
			db.CreateEventSensor(eventSensor)

			PrtyParams, _ := events.PrettyParams(params)
			event := &events.EventLog{
				UserNickname: usuario.NickName,
				Resource:     "Sensor",
				Event:        fmt.Sprintf("El sensor %s fue actualizado correctamente", sensor.Code),
				Params:       PrtyParams,
				Sensor:       sensor.Code,
			}
			event.Write()
		}
		var response responseSensorAddBatch
		response.Data.Status = true
		c.JSON(http.StatusOK, MessageResponse{
			Details: ("Sensor actualizado correctamente"),
			Code:    http.StatusOK,
		})
	})
}

// @Summary post batch autentia sensor
// @Description post batch autentia sensor
// @Tags autentia sensors
// @security BarerToken
// @Accept multipart/form-data
// @Produce json
// @Success 200 "Sensor registrado correctamente"
// @failure 400 {object} MessageResponse
// @Router /autentia/sensors/batch [post]
// @Param sensor formData sensorsBatchAddParams true "sensor"
// @Param sensors formData file true "sensors"
func AutentiaPostSensorBatchRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// create batch Sensors
	router.POST("/sensors/batch", func(c *gin.Context) {
		var params sensorsBatchAddParams
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
		file, err := params.Sensors.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "La plantilla no se pudo abrir",
				Code:    http.StatusBadRequest,
			})
			return
		}

		defer file.Close()
		detector := chardet.NewTextDetector()
		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error inesperado en la plantilla, no se pudo procesar",
				Code:    http.StatusBadRequest,
			})
			return
		}
		result, err := detector.DetectBest(buf.Bytes())
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Datos internos del cvs se encuentran en mal estado, favor descargar la plantilla nuevamente",
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
				Details: "Datos internos del cvs se encuentran en mal estado, favor descargar la plantilla nuevamente",
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
		if len(fileContent[1]) != 9 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "La cantidad de columnas del archivo no coinciden",
				Code:    http.StatusBadRequest,
			})
			return
		}

		var size int

		batchSize := 1000
		for i := 1; i < len(fileContent); i = i + batchSize {
			limit := i + batchSize
			size = len(fileContent)
			if len(fileContent) < limit {
				limit = len(fileContent)
			}

			for in, line := range fileContent[i:limit] {
				in++
				count, _ := db.GetCountryByName(params.Country)
				insti, errInst := db.GetInstitution(line[6], count.ID)

				countstr := strconv.Itoa(in + 1)

				codeRegex := regexp.MustCompile(`^[{a-zA-Z0-9-}]{4,42}$`)
				if !codeRegex.MatchString(line[0]) {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "Serial incorrecto, debe tener entre 4 y 40 carácteres, además de las llaves {}, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}
				technology := strings.ToLower(line[1])
				if technology == "uareu" {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "La tecnología uareu no es soportada en el registro masivo, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}
				if technology != "uareu-gold" {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "La tecnología no corresponde, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}

				logonType, _ := strconv.Atoi(line[2])

				if logonType != 1 && logonType != 4 {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "El tipo logon no es válido, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}

				location := regexp.MustCompile(`^[{0-9}]{1,150}$`)
				if !location.MatchString(line[4]) {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "Debe ingresar código de lugar correcto, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}

				stateString := line[5]

				if stateString != "0" {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "El estado no es válido, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}

				state, _ := strconv.Atoi(line[5])

				if state != 0 {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "El tipo status no es válido, debe registrar un sensor en 0, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}

				if errInst != nil {
					c.JSON(http.StatusNotFound, MessageResponse{
						Details: ("La institución no existe, revisar línea " + countstr + " de su archivo csv"),
						Code:    http.StatusNotFound,
					})
					return
				} else {
					if params.Active == false {
						_, errormatch := db.GetUserAndInstitution(params.ProfileId, insti.ID)

						if errormatch != nil {
							c.JSON(http.StatusNotFound, MessageResponse{
								Details: ("No cuenta con permisos en esta institución " + insti.Name + ", revisar línea " + countstr + " de su archivo csv"),
								Code:    http.StatusNotFound,
							})

							return
						}
					}
				}
				if line[7] != "" {
					brandId, err := db.GetBrandByName(line[7])
					if err != nil {
						c.JSON(http.StatusBadRequest, MessageResponse{
							Details: "La marca no existe revisar la línea " + countstr,
							Code:    http.StatusBadRequest,
						})
						return
					}
					models, err := db.ListModelsByBrand(brandId.ID)

					if err != nil {
						c.JSON(http.StatusBadRequest, MessageResponse{
							Details: "El modelos no existe revisar la línea " + countstr,
							Code:    http.StatusBadRequest,
						})
						return
					}

					if line[8] != "" {
						matchCount := 0
						for _, model := range models {
							if model.Name == line[8] {
								matchCount++
							}
						}

						if matchCount == 0 {
							c.JSON(http.StatusBadRequest, MessageResponse{
								Details: "No hay coincidencias con el modelo y la marca, revisar la línea " + countstr,
								Code:    http.StatusBadRequest,
							})
							return
						}

					}
				}
				if line[4] != "0" {
					if params.DniEntity != "0076957430-1" {
						codigoLugar := line[4]
						dniEntity, _ := db.GetCodigoById(params.DniEntity)
						if len(dniEntity) == 0 {
							c.JSON(http.StatusNotFound, MessageResponse{
								Details: ("No cuenta con los permisos para actualizar el código de lugar " + line[4] + ", revisar línea " + countstr + " de su archivo csv, solicite ayuda con su usuario"),
								Code:    http.StatusNotFound,
							})
							return
						}
						found := false
						for _, codigo := range dniEntity {
							if codigoLugar == codigo.CodigoLugar {
								found = true
								break
							}
						}

						if !found {
							c.JSON(http.StatusNotFound, MessageResponse{
								Details: ("No cuenta con los permisos para actualizar el código de lugar " + line[4] + ", revisar línea " + countstr + " de su archivo csv"),
								Code:    http.StatusNotFound,
							})
							return
						}
					}
				}
			}
		}
		go func() {
			setBatchRegisterSensors(fileContent[1:size], params.Country, params.ProfileId, usuario.NickName)
		}()
		var response responseSensorAddBatch
		response.Data.Status = true
		c.JSON(http.StatusOK, MessageResponse{
			Details: ("Sensor registrado correctamente"),
			Code:    http.StatusOK,
		})
	})
}

// @Summary post batch autentia sensor
// @Description post batch autentia sensor
// @Tags autentia sensors
// @security BarerToken
// @Accept multipart/form-data
// @Produce json
// @Success 200 "Sensor actualizado correctamente"
// @failure 400 {object} MessageResponse
// @Router /autentia/sensors/update/batch [post]
// @Param sensor formData sensorsBatchAddParams true "sensor"
// @Param sensors formData file true "sensors"
func AutentiaUpdateSensorBatchRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.POST("/sensors/update/batch", func(c *gin.Context) {
		var params sensorsBatchAddParams
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

		file, err := params.Sensors.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "La plantilla no se pudo abrir",
				Code:    http.StatusBadRequest,
			})
			return
		}

		defer file.Close()
		detector := chardet.NewTextDetector()
		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error inesperado en la plantilla, no se pudo procesar",
				Code:    http.StatusBadRequest,
			})
			return
		}
		result, err := detector.DetectBest(buf.Bytes())
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Datos internos del cvs se encuentran en mal estado, favor descargar la plantilla nuevamente",
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
				Details: "Datos internos del cvs se encuentran en mal estado, favor descargar la plantilla nuevamente",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(fileContent) <= 1 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "La plantilla debe tener al menos una linea",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(fileContent[1]) != 9 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "La cantidad de columnas del archivo no coinciden",
				Code:    http.StatusBadRequest,
			})
			return
		}

		var size int

		batchSize := 1000
		for i := 1; i < len(fileContent); i = i + batchSize {

			limit := i + batchSize
			size = len(fileContent)
			if len(fileContent) < limit {
				limit = len(fileContent)
			}
			for in, line := range fileContent[i:limit] {
				in++
				count, _ := db.GetCountryByName(params.Country)
				insti, errInst := db.GetInstitution(line[6], count.ID)

				countstr := strconv.Itoa(in + 1)

				codeRegex := regexp.MustCompile(`^[{a-zA-Z0-9-}]{4,42}$`)
				if !codeRegex.MatchString(line[0]) {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "Serial incorrecto, debe tener entre 4 y 40 carácteres, además de las llaves {}, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}
				technology := strings.ToLower(line[1])
				if technology == "uareu" {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "La tecnología uareu no es soportada en el registro masivo, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}
				if technology != "uareu-gold" {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "La tecnología no corresponde, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}

				location := regexp.MustCompile(`^[{0-9}]{1,150}$`)
				if !location.MatchString(line[4]) {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "Debe ingresar código de lugar correcto, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}

				logonType, _ := strconv.Atoi(line[2])

				if logonType != 1 && logonType != 4 {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "El tipo logon no es válido, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}

				stateString := line[5]

				if stateString != "0" && stateString != "1" {
					c.JSON(http.StatusBadRequest, MessageResponse{
						Details: "El estado no es válido, revisar línea " + countstr + " de su archivo csv",
						Code:    http.StatusBadRequest,
					})
					return
				}

				if errInst != nil {
					c.JSON(http.StatusNotFound, MessageResponse{
						Details: ("La institución no existe, revisar línea " + countstr + " de su archivo csv"),
						Code:    http.StatusNotFound,
					})
					return
				} else {
					if params.Active == false {
						_, errormatch := db.GetUserAndInstitution(params.ProfileId, insti.ID)

						if errormatch != nil {
							c.JSON(http.StatusNotFound, MessageResponse{
								Details: ("No cuenta con permisos en esta institución " + insti.Name + ", revisar línea " + countstr + " de su archivo csv"),
								Code:    http.StatusNotFound,
							})

							return
						}
					}
				}

				if line[7] != "" {
					brandId, err := db.GetBrandByName(line[7])
					if err != nil {
						c.JSON(http.StatusBadRequest, MessageResponse{
							Details: "La marca no existe revisar la línea " + countstr,
							Code:    http.StatusBadRequest,
						})
						return
					}
					models, err := db.ListModelsByBrand(brandId.ID)

					if err != nil {
						c.JSON(http.StatusBadRequest, MessageResponse{
							Details: "El modelos no existe revisar la línea " + countstr,
							Code:    http.StatusBadRequest,
						})
						return
					}

					if line[8] != "" {
						matchCount := 0
						for _, model := range models {
							if model.Name == line[8] {
								matchCount++
							}
						}

						if matchCount == 0 {
							c.JSON(http.StatusBadRequest, MessageResponse{
								Details: "No hay coincidencias con el modelo y la marca, revisar la línea " + countstr,
								Code:    http.StatusBadRequest,
							})
							return
						}

					}
				}
				if line[4] != "0" {
					if params.DniEntity != "0076957430-1" {
						codigoLugar := line[4]
						dniEntity, _ := db.GetCodigoById(params.DniEntity)
						if len(dniEntity) == 0 {
							c.JSON(http.StatusNotFound, MessageResponse{
								Details: ("No cuenta con los permisos para actualizar el código de lugar " + line[4] + ", revisar línea " + countstr + " de su archivo csv, solicite ayuda con su usuario"),
								Code:    http.StatusNotFound,
							})
							return
						}
						found := false
						for _, codigo := range dniEntity {
							if codigoLugar == codigo.CodigoLugar {
								found = true
								break
							}
						}

						if !found {
							c.JSON(http.StatusNotFound, MessageResponse{
								Details: ("No cuenta con los permisos para actualizar el código de lugar " + line[4] + ", revisar línea " + countstr + " de su archivo csv"),
								Code:    http.StatusNotFound,
							})
							return
						}
					}
				}
			}
		}
		go func() {
			setBatchUpdatedSensors(fileContent[1:size], params.Country, params.ProfileId, usuario.NickName)
		}()

		var response responseSensorAddBatch
		response.Data.Status = true
		c.JSON(http.StatusOK, MessageResponse{
			Details: ("Sensor actualizado correctamente"),
			Code:    http.StatusOK,
		})

	})
}

func setBatchRegisterSensors(lines [][]string, country string, userId string, nickname string) bool {
	var userData, _ = db.GetUserByID(userId)
	var sensorNotRegistered []mail.SensorEmail
	for _, line := range lines {
		counter := 0
		counter++
		if len(line) == 9 {

			autentiaSensorUare := services.GetSensor(country, strings.TrimSpace(line[0]), "UareU")

			autentiaSensorUareG := services.GetSensor(country, strings.TrimSpace(line[0]), "UareU-gold")

			if autentiaSensorUareG.Code == "" && autentiaSensorUare.Code == "" {
				var sensorEmailRegister = mail.SensorEmail{
					Code:  strings.TrimSpace(line[0]),
					Glosa: "Registrado",
				}
				sensorNotRegistered = append(sensorNotRegistered, sensorEmailRegister)

				technology := strings.ToLower(strings.TrimSpace(line[1]))
				logonType, _ := strconv.Atoi(strings.TrimSpace(line[2]))
				state, _ := strconv.Atoi(strings.TrimSpace(line[5]))
				var sensor = &data.Sensor{
					Code:         strings.TrimSpace(line[0]),
					Technology:   strings.TrimSpace(technology),
					LogonType:    logonType,
					ExternalCode: strings.TrimSpace(line[3]),
					Location:     strings.TrimSpace(line[4]),
					Institution:  strings.TrimSpace(line[6]),
					State:        state,
					Country:      strings.TrimSpace(country),
					Brand:        strings.TrimSpace(line[7]),
					Model:        strings.TrimSpace(line[8]),
				}
				db.CreateSensor(sensor)

				//create event sensor
				var eventSensor = &data.EventSensor{
					Code:        strings.TrimSpace(line[0]),
					Institution: strings.TrimSpace(line[6]),
					User:        userData.NickName,
					Action:      "Registrar",
					Glosa:       "Sensor se registro correctamente",
				}
				db.CreateEventSensor(eventSensor)

				//create log
				PrtyParams, _ := events.PrettyParams(sensor)
				event := &events.EventLog{
					UserNickname: nickname,
					Resource:     "Sensor",
					Event:        fmt.Sprintf("El sensor %s fue registrado correctamente en una carga masiva", sensor.Code),
					Params:       PrtyParams,
					Sensor:       sensor.Code,
				}
				event.Write()

				//create sensor cgi
				services.SetSensor(
					sensor.Country,
					sensor.Code,
					sensor.ExternalCode,
					sensor.Institution,
					technology,
					sensor.Location,
					state,
					logonType,
					sensor.Brand,
					sensor.Model,
				)
			} else {
				var sensorEmailRegister = mail.SensorEmail{
					Code:  strings.TrimSpace(line[0]),
					Glosa: "Registro existente.",
				}
				sensorNotRegistered = append(sensorNotRegistered, sensorEmailRegister)

				var eventSensor = &data.EventSensor{
					Code:        strings.TrimSpace(line[0]),
					Institution: strings.TrimSpace(line[6]),
					User:        userData.NickName,
					Action:      "Registrar",
					Glosa:       "Sensor ya está registrado",
				}
				db.CreateEventSensor(eventSensor)

				PrtyParams, _ := events.PrettyParams(eventSensor)
				event := &events.EventLog{
					UserNickname: nickname,
					Resource:     "Sensor",
					Event:        fmt.Sprintf("El sensor %s ya esta registrado", strings.TrimSpace(line[0])),
					Sensor:       strings.TrimSpace(line[0]),
					Params:       PrtyParams,
				}
				event.Write()
			}
		}
	}
	mail.SendSensorNotRegistered(userData, sensorNotRegistered)
	return true
}

func setBatchUpdatedSensors(lines [][]string, country string, userId string, nickname string) bool {
	var userData, _ = db.GetUserByID(userId)
	var sensorNotRegistered []mail.SensorEmail
	for _, line := range lines {
		counter := 0
		counter++
		if len(line) == 9 {
			autentiaSensorUare := services.GetSensor(country, strings.TrimSpace(line[0]), "UareU")

			autentiaSensorUareG := services.GetSensor(country, strings.TrimSpace(line[0]), "UareU-gold")

			if autentiaSensorUareG.Code == "" && autentiaSensorUare.Code == "" {
				var sensorEmailRegister = mail.SensorEmail{
					Code:  strings.TrimSpace(line[0]),
					Glosa: "Sensor no está registrado",
				}
				sensorNotRegistered = append(sensorNotRegistered, sensorEmailRegister)

				var eventSensor = &data.EventSensor{
					Code:        strings.TrimSpace(line[0]),
					Institution: strings.TrimSpace(line[6]),
					User:        userData.NickName,
					Action:      "Actualizar",
					Glosa:       "Sensor no está registrado",
				}
				db.CreateEventSensor(eventSensor)

				PrtyParams, _ := events.PrettyParams(eventSensor)
				event := &events.EventLog{
					UserNickname: nickname,
					Resource:     "Sensor",
					Event:        fmt.Sprintf("El sensor %s no está registrado", strings.TrimSpace(line[0])),
					Sensor:       strings.TrimSpace(line[0]),
					Params:       PrtyParams,
				}
				event.Write()

			} else {
				var sensorEmailRegister = mail.SensorEmail{
					Code:  strings.TrimSpace(line[0]),
					Glosa: "Actualizado",
				}
				sensorNotRegistered = append(sensorNotRegistered, sensorEmailRegister)
				//update sensor
				technology := strings.ToLower(strings.TrimSpace(line[1]))
				logonType, _ := strconv.Atoi(strings.TrimSpace(line[2]))
				state, _ := strconv.Atoi(strings.TrimSpace(line[5]))
				var sensor = &data.Sensor{
					Code:         strings.TrimSpace(line[0]),
					Technology:   strings.TrimSpace(technology),
					LogonType:    logonType,
					ExternalCode: strings.TrimSpace(line[3]),
					Location:     strings.TrimSpace(line[4]),
					Institution:  strings.TrimSpace(line[6]),
					State:        state,
					Country:      strings.TrimSpace(country),
					Brand:        strings.TrimSpace(line[7]),
					Model:        strings.TrimSpace(line[8]),
				}
				db.UpdateSensor(sensor.Code, sensor)

				//create event sensor
				var eventSensor = &data.EventSensor{
					Code:        strings.TrimSpace(line[0]),
					Institution: strings.TrimSpace(line[6]),
					User:        userData.NickName,
					Action:      "Actualizar",
					Glosa:       "Sensor se actualizó correctamente",
				}
				db.CreateEventSensor(eventSensor)

				//update log
				PrtyParams, _ := events.PrettyParams(sensor)
				event := &events.EventLog{
					UserNickname: nickname,
					Resource:     "Sensor",
					Event:        fmt.Sprintf("El sensor %s fue actualizado correctamente en una carga masiva", sensor.Code),
					Params:       PrtyParams,
					Sensor:       sensor.Code,
				}
				event.Write()

				//update sensor cgi
				services.SetSensor(
					sensor.Country,
					sensor.Code,
					sensor.ExternalCode,
					sensor.Institution,
					technology,
					sensor.Location,
					state,
					logonType,
					sensor.Brand,
					sensor.Model,
				)
			}
		}
	}
	mail.SendSensorNotUpdated(userData, sensorNotRegistered)
	return true
}
