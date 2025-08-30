package routes

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"github.com/imedcl/manager-api/pkg/services"
)

type SoapEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    SoapBody `xml:"Body"`
}

type SoapBody struct {
	CResultado CResultado `xml:"urn:wsgenaudit CResultado"`
}

type CResultado struct {
	Err      string `xml:"Err"`
	Glosa    string `xml:"Glosa"`
	NroAudit string `xml:"NroAudit"`
}

type RequestVigencia struct {
	Run              string `json:"Run"`
	CredentialNumber string `json:"credential_number"`
}

type RequestRegistro struct {
	Run              string `json:"Run"`
	CredentialNumber string `json:"credential_number"`
	Institution      string `json:"institution"`
}

type ResponseVigencia struct {
	Code      int    `json:"code"`
	Resultado string `json:"resultado"`
	Run       string `json:"run"`
	Serie     string `json:"serie"`
	Auditoria string `json:"auditoria"`
}

func validarSerie(cadena string) bool {
	// Expresión regular para validar el formato "ZNNNNNNNN"

	pattern := `^[A-Za-z0-9]{9}$`

	// Compilar la expresión regular
	regex := regexp.MustCompile(pattern)

	// Verificar si la cadena cumple con el patrón
	return regex.MatchString(cadena)
}

func generarAuditoria(country string, rut string, serie string, institucion string, texto string, resultado string, glosa string, nameOper string, name string) string {
	// La URL del servicio SOAP

	url := fmt.Sprintf("%scgi-bin/autentia-genaud2.fcgi", cfg.WsUrl(country))
	// Cuerpo del mensaje SOAP en formato XML
	soapRequest := `<soapenv:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:wsgenaudit">
		<soapenv:Header/>
		<soapenv:Body>
				<urn:GenAudit soapenv:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
					<ReqGen xsi:type="urn:CAuditData">
							<wsUsuario>` + cfg.WsUser(country) + `</wsUsuario>
							<wsClave>` + cfg.WsPass(country) + `</wsClave>
							<Sistema xsi:type="urn:CDatosSistema">
								<Institucion xsi:type="xsd:string">` + institucion + `</Institucion>
								<Origen xsi:type="xsd:string">AutentiaManager</Origen>
								<Operacion xsi:type="xsd:string">VALIDARUT</Operacion>
								<RutOper xsi:type="xsd:string">0000001000-X</RutOper>
								<Resultado xsi:type="xsd:string">` + resultado + `</Resultado>
								<Descripcion xsi:type="xsd:string">` + glosa + `</Descripcion>
								<Version xsi:type="xsd:string">1.1</Version>
							</Sistema>
							<Audit xsi:type="urn:CDatosAuditados">
								<Rut xsi:type="xsd:string">` + rut + `</Rut>
								<Dedo-id xsi:type="xsd:string">0</Dedo-id>
								<Texto1 xsi:type="xsd:string">` + serie + `</Texto1>
								<Texto2 xsi:type="xsd:string">` + texto + `</Texto2>
								<Texto3 xsi:type="xsd:string">` + institucion + `</Texto3>
								<S-Serie xsi:type="xsd:string">wsvalidarutaudit</S-Serie>
								<Valor1 xsi:type="xsd:string">` + nameOper + `</Valor1>
								<Valor2 xsi:type="xsd:string">` + name + `</Valor2>
								<Bmp xsi:type="xsd:base64Binary"/>
								<Wsq xsi:type="xsd:base64Binary"/>
							</Audit>
					</ReqGen>
				</urn:GenAudit>
		</soapenv:Body>
	    </soapenv:Envelope>`

	fmt.Printf("Url: %s\n", url)
	fmt.Printf("soapRequest: %s\n", soapRequest)
	// Crear una solicitud POST
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(soapRequest)))
	if err != nil {
		fmt.Printf("Error al crear la solicitud: %v", err)
	}

	// Establecer encabezados necesarios para SOAP
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")

	// Crear un transporte personalizado que desactive la verificación de certificados SSL
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Desactivar verificación de SSL
		},
	}

	// Crear un cliente HTTP y enviar la solicitud
	client := &http.Client{
		Transport: tr,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error al enviar la solicitud: %v", err)
	}
	defer resp.Body.Close()

	// Leer la respuesta
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error al leer el cuerpo de la respuesta: %v", err)
	}

	// Mostrar el código de estado y el cuerpo de la respuesta
	fmt.Printf("Código de estado : %d\n", resp.StatusCode)
	fmt.Printf("Respuesta:\n%s\n", body)

	var envelope SoapEnvelope

	decoder := xml.NewDecoder(bytes.NewReader(body))
	// Leer el XML y tratar con los espacios de nombres
	err = decoder.Decode(&envelope)
	if err != nil {
		fmt.Printf("Error al analizar la respuesta XML: %v", err)
	}

	// Mostrar el valgor de NroAudit
	if envelope.Body.CResultado.NroAudit != "" {
		fmt.Printf("NroAudit: %s\n", envelope.Body.CResultado.NroAudit)
	} else {
		log.Println("El valor de NroAudit está vacío.")
	}
	return envelope.Body.CResultado.NroAudit
}

func vigencia(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {

	router.POST("/vigencia-cedula", func(c *gin.Context) {

		var params RequestRegistro
		fmt.Println("params-err", params)

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
		if !validarSerie(params.CredentialNumber) {

			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Número de Documento no válido",
				Code:    http.StatusBadRequest,
			})
			return
		}
		data := RequestVigencia{
			Run:              params.Run,
			CredentialNumber: params.CredentialNumber,
		}

		// Convertir los datos a JSON
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Fatal("Error al convertir los datos a JSON:", err)
		}

		// Realizar la solicitud HTTP
		url := cfg.WsRegistroCivil()
		fmt.Println("url-err", url)
		req, err := http.NewRequest("GET", url, bytes.NewBuffer(jsonData))
		fmt.Println("req-err", req)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		// Agregar los encabezados de la solicitud
		req.Header.Set("Content-Type", "application/json")

		// Crear el cliente HTTP y hacer la solicitud
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al hacer la solicitud:" + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		defer resp.Body.Close()

		// Leer la respuesta
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al leer la respuesta:" + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		result := string(body)[13:]
		person := services.GetPerson(usuario.Country.Name, params.Run)
		personAdd := ""
		if person.Names != "" && person.MiddleName != "" {
			personAdd = fmt.Sprintf("%s %s %s", person.Names, person.MiddleName, person.LastName)
		} else if person.Name != "" {
			personAdd = person.Name
		}
		auditoria := generarAuditoria("CHILE", params.Run, params.CredentialNumber, params.Institution, string(body), strconv.Itoa(resp.StatusCode), result, usuario.NickName, personAdd)

		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Vigencia de Cédula",
			Event:        fmt.Sprintf("Rut: %s, Nro. Documento: %s, Cod. Respuesta: %s, Respuesta: %s", params.Run, params.CredentialNumber, strconv.Itoa(resp.StatusCode), result),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, ResponseVigencia{
			Code:      http.StatusOK,
			Resultado: result,
			Auditoria: auditoria,
			Run:       params.Run,
			Serie:     params.CredentialNumber,
		})
		/* 	fmt.Println(strconv.Itoa(resp.StatusCode))
		fmt.Println(string(body)) */

	})
}
