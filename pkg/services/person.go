package services

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/imedcl/manager-api/pkg/config"
)

type personRequest struct {
	XMLName     xml.Name `xml:"Req"`
	WsUser      string   `xml:"wsUsuario"`
	WsPass      string   `xml:"wsClave"`
	Country     string   `xml:"Pais"`
	Operator    string   `xml:"RutOper"`
	Dni         string   `xml:"Rut"`
	Institution string   `xml:"Institucion"`
}
type personAddRequest struct {
	XMLName     xml.Name `xml:"ReqAdd"`
	WsUser      string   `xml:"wsUsuario"`
	WsPass      string   `xml:"wsClave"`
	Country     string   `xml:"Pais"`
	Operator    string   `xml:"RutOper"`
	Dni         string   `xml:"Rut"`
	Institution string   `xml:"Institucion"`
	Name        string   `xml:"Nombre"`
	Names       string   `xml:"Nombres"`
	MiddleName  string   `xml:"APaterno"`
	LastName    string   `xml:"AMaterno"`
	Email       string   `xml:"EMail"`
	Gender      string   `xml:"Sexo"`
	Birthdate   string   `xml:"FechaNac"`
}

type requestWsPerson struct {
	XMLName   xml.Name `xml:"urn:wspersona"`
	NameSpace personRequest
}

type requestWsPersonAdd struct {
	XMLName   xml.Name `xml:"urn:AddPersona"`
	NameSpace personAddRequest
}

// Response
type addPersonResponse struct {
	Body *addPersonBody `xml:"Body"`
}

type addPersonBody struct {
	Result struct {
		Error  string `xml:"Err" json:"error"`
		Detail string `xml:"Glosa" json:"glosa"`
	} `xml:"CResultado" json:"-"`
}

type personResponse struct {
	Body personBody `xml:"Body"`
}
type personBody struct {
	Response *personData `xml:"CPersonaResp"`
}
type personData struct {
	Result struct {
		Error  string `xml:"Err" json:"error"`
		Detail string `xml:"Glosa" json:"glosa"`
	} `xml:"Resultado" json:"-"`
	Dni        string `xml:"Rut" json:"dni"`
	Name       string `xml:"Nombre" json:"name"`
	Names      string `xml:"Nombres" json:"names"`
	MiddleName string `xml:"APaterno" json:"middle_name"`
	LastName   string `xml:"AMaterno" json:"last_name"`
	Country    string `xml:"Pais" json:"country"`
	Gender     string `xml:"Sexo" json:"gender"`
	Birthdate  string `xml:"FechaNac" json:"birthdate"`
}

// ver problema
func GetPerson(country string, dni string) *personData {
	cfg := config.New()
	if country == config.COLOMBIA && !strings.Contains(dni, "-C") {
		dni = fmt.Sprintf("%s-C", dni)
	}
	NameSpace := personRequest{
		WsUser:   cfg.WsUser(country),
		WsPass:   cfg.WsPass(country),
		Operator: cfg.WsOper(country),
		Dni:      dni,
		Country:  country,
	}
	response := soapCall(
		country,
		"autentia-persona.cgi",
		"wspersona",
		requestWsPerson{
			NameSpace: NameSpace,
		})

	var resp personResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}
	return resp.Body.Response
}

func CreatePerson(
	country string,
	dni string,
	name string,
	names string,
	middleName string,
	lastName string,
	institution string,
	gender string,
	birthdate string,
) *addPersonBody {
	cfg := config.New()
	if country == config.COLOMBIA && !strings.Contains(dni, "-C") {
		dni = fmt.Sprintf("%s-C", dni)
	}
	var nameLast = name
	if name == "" || names != "" {
		nameLast = fmt.Sprintf("%s %s %s", names, middleName, lastName)
	}
	NameSpace := personAddRequest{
		WsUser:      cfg.WsUser(country),
		WsPass:      cfg.WsPass(country),
		Operator:    cfg.WsOper(country),
		Dni:         dni,
		Country:     country,
		Name:        nameLast,
		Names:       names,
		MiddleName:  middleName,
		LastName:    lastName,
		Institution: institution,
		Gender:      gender,
		Birthdate:   birthdate,
	}

	response := soapCall(
		country,
		"autentia-persona.cgi",
		"wspersona",
		requestWsPersonAdd{
			NameSpace: NameSpace,
		})

	var resp addPersonResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}

	return resp.Body
}
