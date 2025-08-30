package routes

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/imedcl/manager-api/pkg/config"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"github.com/imedcl/manager-api/pkg/mail"
	"github.com/imedcl/manager-api/pkg/services"
)

type PersonResponse struct {
	Data *data.Role `json:"data"`
}

type AutentiaPersonParams struct {
	Id              string `form:"id"`
	Dni             string `form:"dni" binding:"required"`
	Country         string `form:"country"`
	Name            string `form:"name" `
	Names           string `form:"names"`
	MiddleName      string `form:"middle_name" json:"middle_name"`
	LastName        string `form:"last_name" json:"last_name"`
	Institution     string `form:"institution"`
	Gender          string `form:"gender"`
	Birthdate       string `form:"birthdate"`
	VersionChange   string `form:"version_change"`
	NroAudit        string `form:"nro_audit" json:"nro_audit"`
	Description     string `form:"description"`
	InstitutionBloq string `form:"institution_bloq" json:"institution_bloq"`
	TypeBloq        string `form:"type_bloq" json:"type_bloq"`
	UserId          string `form:"users_id" json:"users_id"`
}

type PersonVerificationParams struct {
	Dni      string `form:"dni" binding:"required"`
	Ambiente string `form:"ambiente"`
}

type PeopleFileAddParams struct {
	People      *multipart.FileHeader `form:"people"`
	Delimiter   string                `form:"delimiter"`
	Name        string                `form:"name"`
	Uri         string                `form:"uri"`
	Description string                `form:"descriptions"`
	Dni         string                `form:"dni"`
	PersonId    string                `form:"person_id" json:"person_id"`
}
type responsePerson struct {
	Data interface{} `json:"data"`
}

type responseStorage struct {
	Status  int           `json:"status"`
	Message string        `json:"message"`
	Result  storageResult `json:"result"`
}

type storageResult struct {
	Code string `json:"code"`
	Url  string `json:"url"`
}

type responseFile struct {
	Status int        `json:"status"`
	Result fileResult `json:"result"`
}

type fileResult struct {
	File string `json:"file"`
}

var MEN_SEX = []string{"M", "MASCULINO"}
var WOMAN_SEX = []string{"F", "FEMENINO"}
var NOBIN_SEX = []string{"X", "NO BINARIO"}
var ALL_SEX = []string{"M", "F", "X", "FEMENINO", "MASCULINO", "NO BINARIO"}
var cfg = config.New()

func dateFormated(date string) (newDate string) {
	dateSplit := strings.Split(strings.Split(date, " ")[0], "-")
	if len(dateSplit) != 3 {
		dateSplit = strings.Split(date, "/")
	}
	if len(dateSplit) == 3 {
		if len(dateSplit[0]) == 4 {
			newDate = fmt.Sprintf("%s-%s-%s", dateSplit[0], dateSplit[1], dateSplit[2])
		} else if len(dateSplit[2]) == 4 {
			newDate = fmt.Sprintf("%s-%s-%s", dateSplit[2], dateSplit[1], dateSplit[0])
		}
	}
	return
}

// @Summary autentia person post old
// @Description autentia person post old
// @Tags autentia persons
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responsePerson
// @failure 400 {object} MessageResponse
// @Router /autentia/persons/manager/old [post]
// @Param person body AutentiaPersonParams true "person"
func AutentiaPersonsOldRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Create History Person Autentia in Manager Old
	router.POST("/persons/manager/old", func(c *gin.Context) {

		var params AutentiaPersonParams
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

		gender := strings.ToUpper(params.Gender)

		if gender == NOBIN_SEX[1] {
			gender = NOBIN_SEX[0]
		}
		if gender == MEN_SEX[1] {
			gender = MEN_SEX[0]
		}
		if gender == WOMAN_SEX[1] {
			gender = WOMAN_SEX[0]
		}

		personsManager, versionErro := db.CountPeopleVersion(params.Dni)

		var versionstring = 0
		if versionErro != nil {
			versionstring = 1
		} else {
			versionstring = personsManager.Version + 1
		}

		personsManager = &data.AutentiaPerson{
			Country:         params.Country,
			Dni:             params.Dni,
			Name:            params.Name,
			Names:           params.Names,
			MiddleName:      params.MiddleName,
			LastName:        params.LastName,
			Institution:     params.Institution,
			Gender:          gender,
			Birthdate:       params.Birthdate,
			Description:     params.Description,
			InstitutionBloq: params.InstitutionBloq,
			TypeBloq:        params.TypeBloq,
			NroAudit:        params.NroAudit,
			UserID:          params.UserId,
			Version:         versionstring,
			VersionChange:   "Anterior",
		}

		db.CreatePersonManager(personsManager)
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Personas",
			Event:        fmt.Sprintf("Se ha registrado y realizado cambios en el rut %s", params.Dni),
			Params:       PrtyParams,
			PersonDni:    personsManager.Dni,
		}
		event.Write()
		c.JSON(http.StatusOK, responsePerson{Data: personsManager})
	})
}

// @Summary autentia person post new
// @Description autentia person post new
// @Tags autentia persons
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responsePerson
// @failure 400 {object} MessageResponse
// @Router /autentia/persons/manager/new [post]
// @Param person body AutentiaPersonParams true "person"
func AutentiaPersonsNewRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {

	// Create History Person Autentia in Manager New
	router.POST("/persons/manager/new", func(c *gin.Context) {

		var params AutentiaPersonParams
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

		gender := strings.ToUpper(params.Gender)

		if gender == NOBIN_SEX[1] {
			gender = NOBIN_SEX[0]
		}
		if gender == MEN_SEX[1] {
			gender = MEN_SEX[0]
		}
		if gender == WOMAN_SEX[1] {
			gender = WOMAN_SEX[0]
		}
		personsManager, _ := db.CountPeopleVersion(params.Dni)
		var versionstring = personsManager.Version
		personsManager = &data.AutentiaPerson{
			Country:         params.Country,
			Dni:             params.Dni,
			Name:            params.Name,
			Names:           params.Names,
			MiddleName:      params.MiddleName,
			LastName:        params.LastName,
			Institution:     params.Institution,
			Gender:          gender,
			Birthdate:       params.Birthdate,
			Description:     params.Description,
			InstitutionBloq: params.InstitutionBloq,
			TypeBloq:        params.TypeBloq,
			NroAudit:        params.NroAudit,
			UserID:          params.UserId,
			Version:         versionstring,
			VersionChange:   "Actual",
		}

		db.CreatePersonManager(personsManager)
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Personas",
			Event:        fmt.Sprintf("Se han realizado cambios en el rut %s", params.Dni),
			Params:       PrtyParams,
			PersonDni:    personsManager.Dni,
		}
		event.Write()
		c.JSON(http.StatusOK, responsePerson{Data: personsManager})
	})
}

// @Summary autentia person post document
// @Description autentia person post document
// @Tags autentia persons
// @security BarerToken
// @Accept multipart/form-data
// @Produce json
// @Success 200 {object} responseStorage
// @failure 400 {object} MessageResponse
// @Router /autentia/persons/document [post]
// @Param document formData PeopleFileAddParams true "document"
// @Param file formData file true "people"
func AutentiaPersonsDocumentRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Update document
	router.POST("/persons/document", func(c *gin.Context) {
		var params PeopleFileAddParams
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

		file, headers, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "Error al leer el archivo",
				Code:    http.StatusBadRequest,
			})
			return
		}

		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "Error al leer el archivo",
				Code:    http.StatusBadRequest,
			})
			return
		}

		rawDecodedText := base64.StdEncoding.EncodeToString(buf.Bytes())

		dataValues := url.Values{}
		dataValues.Set("name", headers.Filename)
		dataValues.Set("file_mime", headers.Header["Content-Type"][0])
		dataValues.Set("institution", "AUTENTIAMANAGER")
		dataValues.Set("file", rawDecodedText)
		encodedData := dataValues.Encode()
		responseURL := fmt.Sprintf("%s/api/v1/s3/documents/upload", cfg.UrlStorage())
		response := config.PostRequest(responseURL, "application/x-www-form-urlencoded", encodedData)

		responseJson := responseStorage{}

		err = json.Unmarshal([]byte(response), &responseJson)

		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		documentRegister := &data.PersonDocument{
			Name:             config.GenerateUniqueFilename(headers.Filename),
			Uri:              responseJson.Result.Code,
			AutentiaPersonId: params.PersonId,
		}

		db.CreateDocumentRegisterPerson(documentRegister)
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Personas",
			Event:        fmt.Sprintf("Se registro el archivo %s", headers.Filename),
			Params:       PrtyParams,
			PersonDni:    params.Dni,
		}
		event.Write()

		c.JSON(http.StatusOK, responseJson)

	})
}

// @Summary autentia person get
// @Description autentia person get
// @Tags autentia persons
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responsePerson
// @failure 400 {object} MessageResponse
// @Router /autentia/persons/all/{dni} [get]
// @param dni path string true "dni"
func AutentiaPersonsGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get Person bd
	router.GET("/persons/all/:dni", func(c *gin.Context) {
		dni := c.Param("dni")

		if dni == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Dni no puede venir vacio",
				Code:    http.StatusBadRequest,
			})
			return
		}
		_, err := db.ExistsPeopleManagerbyDni(dni)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "Persona no existe",
				Code:    http.StatusBadRequest,
			})
			return
		}
		people, _ := db.ExistsPeopleManagerbyDniVersion(dni)

		c.JSON(http.StatusOK, responsePerson{Data: people})
	})
}

// @Summary autentia person document get
// @Description autentia person document get
// @Tags autentia persons
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responsePerson
// @failure 400 {object} MessageResponse
// @Router /autentia/persons/documents/{dni} [get]
// @param dni path string true "dni"
func AutentiaPersonsDocumentGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// al date person and document
	router.GET("/persons/documents/:dni", func(c *gin.Context) {
		dni := c.Param("dni")

		if dni == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Dni no puede venir vacio",
				Code:    http.StatusBadRequest,
			})
			return
		}
		_, err := db.ExistsPeopleManagerbyDni(dni)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "Persona no existe",
				Code:    http.StatusBadRequest,
			})
			return
		}
		person, err := db.GetDocumentsByPerson(dni)

		for _, document := range person {
			url := fmt.Sprintf("%s/api/v1/documents?code=%s&institution=AUTENTIAMANAGER&extra=file", cfg.UrlStorage(), document.Uri)
			response := config.GetRequest(url)

			responseJson := responseFile{}

			err = json.Unmarshal([]byte(response), &responseJson)

			if err != nil {
				c.JSON(http.StatusNotFound, MessageResponse{
					Details: err.Error(),
					Code:    http.StatusBadRequest,
				})
				return
			}

			dec, err := base64.StdEncoding.DecodeString(responseJson.Result.File)
			if err != nil {
				c.JSON(http.StatusNotFound, MessageResponse{
					Details: err.Error(),
					Code:    http.StatusBadRequest,
				})
				return
			}

			if _, err := os.Stat("./upload/" + dni + "/"); os.IsNotExist(err) {
				os.MkdirAll("./upload/"+dni, 0777)
			}

			filename := document.Name
			filePath := "/upload/" + dni + "/" + filename

			f, err := os.Create("upload/" + dni + "/" + filename)
			if err != nil {
				c.JSON(http.StatusNotFound, MessageResponse{
					Details: err.Error(),
					Code:    http.StatusBadRequest,
				})
				return
			}
			defer f.Close()

			if _, err := f.Write(dec); err != nil {
				c.JSON(http.StatusNotFound, MessageResponse{
					Details: err.Error(),
					Code:    http.StatusBadRequest,
				})
				return
			}
			if err := f.Sync(); err != nil {
				c.JSON(http.StatusNotFound, MessageResponse{
					Details: err.Error(),
					Code:    http.StatusBadRequest,
				})
				return
			}

			document.Uri = filePath
		}

		c.JSON(http.StatusOK, responsePerson{Data: person})
	})
}

// @Summary autentia person document version get
// @Description autentia person document version get
// @Tags autentia persons
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responsePerson
// @failure 400 {object} MessageResponse
// @Router /autentia/persons/document/version/{dni} [get]
// @param dni path string true "dni"
// @param version query string true "version"
func AutentiaPersonsDocumentVersionGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//version person get document
	router.GET("/persons/document/version/:dni", func(c *gin.Context) {
		dni := c.Param("dni")
		version := c.Query("version")

		if dni == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Dni no puede venir vacio",
				Code:    http.StatusBadRequest,
			})
			return
		}
		person, err := db.GetDocumentByPersonVersion(dni, version)
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: "Persona no existe",
				Code:    http.StatusBadRequest,
			})
			return
		}

		for _, document := range person {

			url := fmt.Sprintf("%s/api/v1/documents?code=%s&institution=AUTENTIAMANAGER&extra=file", cfg.UrlStorage(), document.Uri)
			/* url := "https://api-pre.imedlife.com/api-external-storage/api/v1/documents?code=" + document.Uri + "&institution=AUTENTIAMANAGER&extra=file"  */
			response := config.GetRequest(url)

			responseJson := responseFile{}

			err = json.Unmarshal([]byte(response), &responseJson)

			if err != nil {
				c.JSON(http.StatusNotFound, MessageResponse{
					Details: err.Error(),
					Code:    http.StatusBadRequest,
				})
				return
			}

			dec, err := base64.StdEncoding.DecodeString(responseJson.Result.File)
			if err != nil {
				c.JSON(http.StatusNotFound, MessageResponse{
					Details: err.Error(),
					Code:    http.StatusBadRequest,
				})
				return
			}

			if _, err := os.Stat("./upload/" + dni + "/"); os.IsNotExist(err) {
				os.MkdirAll("./upload/"+dni, 0777)
			}

			filename := document.Name
			filePath := "/upload/" + dni + "/" + filename

			f, err := os.Create("upload/" + dni + "/" + filename)
			if err != nil {
				c.JSON(http.StatusNotFound, MessageResponse{
					Details: err.Error(),
					Code:    http.StatusBadRequest,
				})
				return
			}
			defer f.Close()

			if _, err := f.Write(dec); err != nil {
				c.JSON(http.StatusNotFound, MessageResponse{
					Details: err.Error(),
					Code:    http.StatusBadRequest,
				})
				return
			}
			if err := f.Sync(); err != nil {
				c.JSON(http.StatusNotFound, MessageResponse{
					Details: err.Error(),
					Code:    http.StatusBadRequest,
				})
				return
			}

			document.Uri = filePath
		}

		c.JSON(http.StatusOK, responsePerson{Data: person})
	})
}

// @Summary autentia person compare get
// @Description autentia person compare get
// @Tags autentia persons
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responsePerson
// @failure 400 {object} MessageResponse
// @Router /autentia/persons/compare/{dni} [get]
// @param dni path string true "dni"
// @param version query string true "version"
func AutentiaPersonsCompareGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get person compare
	router.GET("/persons/compare/:dni", func(c *gin.Context) {
		dni := c.Param("dni")
		version := c.Query("version")

		if dni == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Dni no puede venir vacío",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if version == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Versión no puede venir vacío",
				Code:    http.StatusBadRequest,
			})
			return
		}
		people, _ := db.SearchPeopleManager(dni, version)

		c.JSON(http.StatusOK, responsePerson{Data: people})
	})
}

// @Summary autentia person dni get
// @Description autentia person dni get
// @Tags autentia persons
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responsePerson
// @failure 400 {object} MessageResponse
// @Router /autentia/persons/{dni} [get]
// @param dni path string true "dni"
// @param country query string true "country"
func AutentiaPersonGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// old
	// Get Person CGI
	router.GET("/persons/:dni", func(c *gin.Context) {

		dni := c.Param("dni")
		country := c.Query("country")
		if country == "" || dni == "" {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Parametros incorrectos",
				Code:    http.StatusBadRequest,
			})
			return
		}
		person := services.GetPerson(country, dni)

		gender := strings.ToUpper(person.Gender)

		if contains(MEN_SEX, gender) {
			person.Gender = MEN_SEX[0]
		}
		if contains(WOMAN_SEX, gender) {
			person.Gender = WOMAN_SEX[0]
		}
		if contains(NOBIN_SEX, gender) {
			person.Gender = NOBIN_SEX[0]
		}

		c.JSON(http.StatusOK, responsePerson{Data: person})
	})
}

// @Summary autentia person document delete
// @Description autentia person document delete
// @Tags autentia persons
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 400 {object} MessageResponse
// @Router /autentia/persons/documents [delete]
func AutentiaPersonDeleteRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Delete documents
	router.DELETE("/persons/documents", func(c *gin.Context) {

		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		person, err := db.GetDocuments()
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "No hay documentos",
				Code:    http.StatusBadRequest,
			})
			return
		}

		for _, document := range person {
			if db.DeleteDocument(document) != nil {
				c.JSON(http.StatusBadRequest, MessageResponse{
					Details: "Error al eliminar documentos, intente nuevamente",
					Code:    http.StatusBadRequest,
				})
				return
			}
		}
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Personas",
			Event:        fmt.Sprintf("Se han eliminado los documentos"),
		}
		event.Write()
		c.JSON(http.StatusOK, MessageResponse{Details: "Se han eliminado los documentos", Code: http.StatusOK})
	})
}

// @Summary autentia person status get
// @Description autentia person status get
// @Tags autentia persons
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} bool
// @failure 400 {object} MessageResponse
// @Router /autentia/persons/document/status [get]
func AutentiaPersonStatusGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Document api status
	router.GET("/persons/document/status", func(c *gin.Context) {
		status := true

		dataValues := url.Values{}
		encodedData := dataValues.Encode()

		responseURL := fmt.Sprintf("%s/api/v1/s3/documents/upload", cfg.UrlStorage())
		response := config.PostRequest(responseURL, "application/x-www-form-urlencoded", encodedData)

		responseJson := responseStorage{}

		err := json.Unmarshal([]byte(response), &responseJson)

		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if responseJson.Message == "no Route matched with those values" {
			status = false
		}

		c.JSON(http.StatusOK, status)
	})
}

// @Summary autentia person post validation
// @Description autentia person post validation
// @Tags autentia persons
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responsePerson
// @failure 400 {object} MessageResponse
// @Router /autentia/persons/validation [post]
// @Param PersonVerificationParams body Ok
func AutentiaPersonValidationRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Delete documents
	router.POST("/persons/validation", func(c *gin.Context) {
		var params PersonVerificationParams
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
		mailParams := mail.VerificationParams{
			Dni:      params.Dni,
			Ambiente: params.Ambiente,
		}
		mail.SendPeopleValidation(currentUser, mailParams)
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Validacion persona",
			Params:       PrtyParams,
			Event:        fmt.Sprintf("Se ha reportado el rut %s", params.Dni),
		}
		event.Write()
		c.JSON(http.StatusOK, MessageResponse{Details: "Se ha enviado un reporte", Code: http.StatusOK})
	})
}
