package services

import (
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/imedcl/manager-api/pkg/config"
)

type sensorRequest struct {
	XMLName    xml.Name `xml:"SensorReq"`
	WsUser     string   `xml:"wsUsuario"`
	WsPass     string   `xml:"wsClave"`
	Country    string   `xml:"Pais"`
	Operator   string   `xml:"RutOper"`
	Code       string   `xml:"CodInterno"`
	Technology string   `xml:"Tecnologia"`
}
type sensorAddRequest struct {
	XMLName     xml.Name `xml:"SensorReqAdd"`
	WsUser      string   `xml:"wsUsuario"`
	WsPass      string   `xml:"wsClave"`
	Country     string   `xml:"Pais"`
	Operator    string   `xml:"RutOper"`
	Code        string   `xml:"CodInterno"`
	External    string   `xml:"CodExterno"`
	Location    string   `xml:"Ubicacion"`
	Technology  string   `xml:"Tecnologia"`
	Institution string   `xml:"Institucion"`
	State       int      `xml:"Estado" `
	Logon       int      `xml:"TipoLogon"`
	Brand       string   `xml:"Marca" `
	Model       string   `xml:"Modelo" `
}

type requestWsSensor struct {
	XMLName   xml.Name `xml:"urn:wssensor"`
	NameSpace sensorRequest
}

type requestWsSensorAdd struct {
	XMLName   xml.Name `xml:"urn:wssensoradd"`
	NameSpace sensorAddRequest
}

// Response
type addSensorsResponse struct {
	Body addSensorsBody `xml:"Body"`
}

type addSensorsBody struct {
	Result struct {
		Error  string `xml:"Err"`
		Detail string `xml:"Glosa"`
	} `xml:"CResultado"`
}

type sensorsResponse struct {
	Body *sensorsBody `xml:"Body"`
}
type sensorsBody struct {
	Response *sensorsData `xml:"CSensorResp"`
}
type sensorsData struct {
	Result struct {
		Error  string `xml:"Err" json:"error"`
		Detail string `xml:"Glosa" json:"glosa"`
	} `xml:"Resultado" json:"result"`
	Technology   string      `xml:"Tecnologia" json:"technology"`
	Code         string      `xml:"CodInterno" json:"code"`
	External     string      `xml:"CodExterno" json:"external_code"`
	Institution  string      `xml:"Institucion" json:"institution"`
	Location     string      `xml:"Ubicacion" json:"location"`
	LocationCode string      `xml:"CodLugar" json:"location_code"`
	Ubication    interface{} `json:"ubication"`
	Description  string      `xml:"Descripcion" json:"description"`
	RegisterAt   string      `xml:"Registrado" json:"register_at"`
	LogonType    int         `xml:"TipoLogon" json:"logon_type"`
	LogonDNI     string      `xml:"RutLogon" json:"logon_dni"`
	LastOperator string      `xml:"LastOper" json:"last_operator"`
	DateFrom     string      `xml:"FechaDesde" json:"date_from"`
	State        int         `xml:"Estado" json:"state"`
	DateTo       string      `xml:"FechaHasta" json:"date_to"`
	Brand        string      `xml:"Marca" json:"brand"`
	Model        string      `xml:"Modelo" json:"model"`
}

func GetSensor(country string, code string, technology string) *sensorsData {
	cfg := config.New()
	NameSpace := sensorRequest{
		WsUser:     cfg.WsUser(country),
		WsPass:     cfg.WsPass(country),
		Operator:   cfg.WsOper(country),
		Technology: technology,
		Country:    country,
		Code:       code,
	}
	response := soapCall(
		country,
		"autentia-sensor.cgi",
		"wssensor",
		requestWsSensor{
			NameSpace: NameSpace,
		})

	var resp sensorsResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}

	return resp.Body.Response
}

func SetSensor(country string, code string, external string, institution string, technology string, location string, state int, logon int, brand string, model string) (bool, error) {
	cfg := config.New()
	NameSpace := sensorAddRequest{
		WsUser:      cfg.WsUser(country),
		WsPass:      cfg.WsPass(country),
		Operator:    cfg.WsOper(country),
		Technology:  technology,
		Country:     country,
		Code:        code,
		External:    external,
		Location:    location,
		Institution: institution,
		State:       state,
		Logon:       logon,
		Brand:       brand,
		Model:       model,
	}
	response := soapCall(
		country,
		"autentia-sensor.cgi",
		"wssensor",
		requestWsSensorAdd{
			NameSpace: NameSpace,
		},
	)

	// Verificación si el cuerpo de la respuesta está vacío
	if len(response) == 0 {
		return false, errors.New("la respuesta SOAP está vacía")
	}
	var resp addSensorsResponse
	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error al deserializar la respuesta SOAP:", err.Error())
		return false, fmt.Errorf("error al deserializar la respuesta SOAP: %w", err)
	}

	if resp.Body.Result.Error == "" {
		return false, errors.New("la respuesta SOAP no contiene el campo 'Error' esperado")
	}

	if resp.Body.Result.Error != "0" {
		return false, fmt.Errorf(resp.Body.Result.Detail)
	}

	return true, nil

}
