package services

import (
	"encoding/xml"
	"fmt"

	"github.com/imedcl/manager-api/pkg/config"
)

type auditRequest struct {
	XMLName     xml.Name `xml:"WSAuditReadReq"`
	WsUser      string   `xml:"wsUsuario"`
	WsPass      string   `xml:"wsClave"`
	Country     string   `xml:"Pais"`
	Operator    string   `xml:"RutOper"`
	Wsq         bool     `xml:"bWsq"`
	Bmp         bool     `xml:"bBmp"`
	AuditNumber string   `xml:"NroAudit"`
}

type requestWsAudit struct {
	XMLName   xml.Name `xml:"urn:wsaudit"`
	NameSpace auditRequest
}

// Response
type auditResponse struct {
	Body auditBody `xml:"Body"`
}
type auditBody struct {
	Response auditData `xml:"WSAuditReadResp"`
}
type auditData struct {
	Result struct {
		Error  string `xml:"Err" json:"error"`
		Detail string `xml:"Glosa" json:"glosa"`
	} `xml:"Resultado" json:"result"`
	System struct {
		AuditNumber string `xml:"NroAudit" json:"audit_number"`
		Institution string `xml:"Institucion" json:"institution"`
		Origin      string `xml:"Origen" json:"origin"`
		Operation   string `xml:"Operacion" json:"operation"`
		Station     string `xml:"Estacion" json:"station"`
		Description string `xml:"Descripcion" json:"description"`
		Version     string `xml:"Version" json:"version"`
		Result      string `xml:"Resultado" json:"result"`
		OperatorDni string `xml:"RutOper" json:"operator_dni"`
		Registered  string `xml:"Registrado" json:"registered"`
	} `xml:"DatosSistema" json:"system"`
	Audit struct {
		Dni        string `xml:"Rut" json:"dni"`
		Name       string `xml:"Nombre" json:"name"`
		FingerID   string `xml:"Dedo-id" json:"finger_id"`
		FingerDate string `xml:"Dedo-Fecha" json:"finger_date"`
		Operator   string `xml:"NombreOper" json:"operator"`
		Enrollment string `xml:"Enrolado" json:"enrollment"`
		Place      string `xml:"CodLugar" json:"place"`
		Password   string `xml:"ConClave" json:"password"`
		Sensor     string `xml:"S-Serie" json:"sensor"`
		Texto1     string `xml:"Texto1" json:"texto1"`
		Valor1     string `xml:"Valor1" json:"valor1"`
		Valor2     string `xml:"Valor2" json:"valor2"`
	} `xml:"DatosAuditados" json:"audit"`
}

func GetAudit(country string, auditNumber string) *auditData {
	cfg := config.New()
	NameSpace := auditRequest{
		WsUser:      cfg.WsUser(country),
		WsPass:      cfg.WsPass(country),
		Operator:    cfg.WsOper(country),
		AuditNumber: auditNumber,
		Country:     country,
		Bmp:         false,
		Wsq:         false,
	}
	response := soapCall(
		country,
		"autentia-audit.cgi",
		"wsaudit",
		requestWsAudit{
			NameSpace: NameSpace,
		})

	var resp auditResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}

	return &resp.Body.Response
}
