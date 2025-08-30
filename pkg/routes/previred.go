package routes

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/imedcl/manager-api/pkg/config"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"github.com/imedcl/manager-api/pkg/services"
	"golang.org/x/net/html/charset"
)

type previredSoapParams struct {
	RutOper     string `form:"rut_oper" json:"rut_oper"`
	Rutclient   string `form:"rut_client" json:"rut_client"`
	CodAutoriza string `form:"cod_autoriza" json:"cod_autoriza"`
}

func PreviRedgetCallSoapRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Call soap previred
	router.POST("previred/call", func(c *gin.Context) {
		var params previredSoapParams
		cfg := config.New()
		if err := c.ShouldBindJSON(&params); err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		usuario, err := GetUserFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		resp, err := services.EnviarPeticionPrevired(cfg.UserPrevired(), cfg.PassPrevi(), params.Rutclient, params.CodAutoriza, "12")

		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if !bytes.Contains(resp, []byte("<soapenv:Envelope")) {
			fmt.Println("Respuesta no es SOAP. HTML recibido:")
			fmt.Println(string(resp))
			return
		}

		// Intentar parsear como XML SOAP
		var soapResp services.SoapResponse
		if err := xml.Unmarshal(resp, &soapResp); err != nil {
			fmt.Println("Error al parsear respuesta SOAP:", err)
			fmt.Println("Contenido recibido:", string(resp))
			return
		}
		cdata := soapResp.Body.Response.Return

		// Paso 2: parsear el contenido interno (CDATA)
		var respuesta services.RespuestaPrevired
		decoder := xml.NewDecoder(strings.NewReader(cdata))
		decoder.CharsetReader = charset.NewReaderLabel

		if err := decoder.Decode(&respuesta); err != nil {
			fmt.Println("Error al parsear XML interno:", err)
			return
		}
		mensajeCCX := respuesta.Control.Codigo

		estadoCCX := "Error"
		if mensajeCCX == "9050" || mensajeCCX == "9000" {
			estadoCCX = "Exito"
		}

		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "PreviRed",
			Event:        fmt.Sprintf("%s al solicitar el listado de cotizaciones %s, del rut, codigo previred: %s", estadoCCX, params.Rutclient, mensajeCCX),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, gin.H{
			"rut":             respuesta.RespuestaCCX.Encabezado.Rut,
			"nombre_completo": fmt.Sprintf("%s %s %s", respuesta.RespuestaCCX.Encabezado.Nombres, respuesta.RespuestaCCX.Encabezado.ApellidoPaterno, respuesta.RespuestaCCX.Encabezado.ApellidoMaterno),
			"cotizaciones":    respuesta.RespuestaCCX.Detalles,
			"Codigo":          mensajeCCX,
		})

	})

}
