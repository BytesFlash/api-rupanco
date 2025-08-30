package services

import (
	"encoding/xml"
	"fmt"

	"github.com/imedcl/manager-api/pkg/config"
)

type getInstitutionRequest struct {
	XMLName  xml.Name `xml:"GetInfoReq"`
	WsUser   string   `xml:"wsUsuario"`
	WsPass   string   `xml:"wsClave"`
	Country  string   `xml:"Pais"`
	Name     string   `xml:"CodInstit"`
	Operator string   `xml:"RutOper"`
}

type institutionReqBody struct {
	XMLName     xml.Name `xml:"Instit"`
	Country     string   `xml:"Pais"`
	Name        string   `xml:"CodInstit"`
	Nemo        string   `xml:"Nemo"`
	Description string   `xml:"Descripcion"`
	Email       string   `xml:"eMail"`
	State       int      `xml:"Estado"`
	FlagDec     int      `xml:"FlagDec"`
	Dni         string   `xml:"Rut"`
	Operator    string   `xml:"RutOper"`
}

type institutionRequest struct {
	XMLName     xml.Name `xml:"CreateReq"`
	WsUser      string   `xml:"wsUsuario"`
	WsPass      string   `xml:"wsClave"`
	Institution institutionReqBody
}

type listInstitutionsRequest struct {
	XMLName  xml.Name `xml:"ListReq"`
	WsUser   string   `xml:"wsUsuario"`
	WsPass   string   `xml:"wsClave"`
	Country  string   `xml:"Pais"`
	Operator string   `xml:"RutOper"`
}

type requestWsCreateInstit struct {
	XMLName   xml.Name `xml:"urn:Create"`
	NameSpace institutionRequest
}

type requestWsUpdateInstit struct {
	XMLName   xml.Name `xml:"urn:Modify"`
	NameSpace institutionRequest
}

type requestWsInstits struct {
	XMLName   xml.Name `xml:"urn:List"`
	NameSpace listInstitutionsRequest
}

type requestWsInstit struct {
	XMLName   xml.Name `xml:"urn:GetInfo"`
	NameSpace getInstitutionRequest
}

// Response
type institutionsResponse struct {
	Body institutionsBody `xml:"Body"`
}

type institutionResponse struct {
	Body institutionBody `xml:"Body"`
}

type institutionsBody struct {
	Response institutionsData `xml:"ListResp"`
}

type institutionBody struct {
	Response institutionData `xml:"GetInfoResp"`
}

type institutionsData struct {
	Result struct {
		Error  string `xml:"Err" json:"error"`
		Detail string `xml:"Glosa" json:"glosa"`
	} `xml:"Resultado" json:"result"`
	Institutions []*institution `xml:"Instit" json:"institution"`
}

type institutionData struct {
	Result struct {
		Error  string `xml:"Err" json:"error"`
		Detail string `xml:"Glosa" json:"glosa"`
	} `xml:"Resultado" json:"result"`
	Institution *institution `xml:"Instit" json:"institution"`
}

type institution struct {
	Country        string `xml:"Pais" json:"country"`
	Description    string `xml:"Descripcion" json:"description"`
	Email          string `xml:"eMail" json:"email"`
	RutInstitution string `xml:"RutInstit" json:"dni_institution"`
	Place          string `xml:"CodLugar" json:"place"`
	System         string `xml:"Sistema" json:"system"`
	Rut            string `xml:"Rut" json:"dni"`
	Role           string `xml:"Rol" json:"role"`
	Phone          string `xml:"celular" json:"phone"`
	RoleFrom       string `xml:"RolDesde" json:"role_from"`
	RoleTo         string `xml:"RolHasta" json:"role_to"`
	PasswordDate   string `xml:"FecClave" json:"password_date"`
	FlagDec        int    `xml:"FlagDec" json:"flag_dec"`
	Name           string `xml:"CodInstit" json:"name"`
	State          int    `xml:"Estado" json:"state"`
	CreatedAt      string `xml:"Registrado" json:"created_at"`
	Verificated    string `xml:"Verificado" json:"verificated"`
	Expiration     string `xml:"Expiracion" json:"expiration"`
	Nemo           string `xml:"Nemo" json:"nemo"`
}

func ListInstitutions(country string) []*institution {
	cfg := config.New()
	NameSpace := listInstitutionsRequest{
		WsUser:   cfg.WsUser(country),
		WsPass:   cfg.WsPass(country),
		Country:  country,
		Operator: cfg.WsOper(country),
	}
	response := soapCall(
		country,
		"autentia-instit.cgi",
		"wsinstit",
		requestWsInstits{
			NameSpace: NameSpace,
		})

	var resp institutionsResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}

	return resp.Body.Response.Institutions
}

func GetInstitution(country string, name string) *institution {
	cfg := config.New()
	NameSpace := getInstitutionRequest{
		WsUser:   cfg.WsUser(country),
		WsPass:   cfg.WsPass(country),
		Country:  country,
		Name:     name,
		Operator: cfg.WsOper(country),
	}
	response := soapCall(
		country,
		"autentia-instit.cgi",
		"wsinstit",
		requestWsInstit{
			NameSpace: NameSpace,
		})

	var resp institutionResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}

	return resp.Body.Response.Institution
}

func CreateInstitution(
	country string,
	name string,
	nemo string,
	email string,
	state int,
	description string,
	flag_dec int,
	dni string,
) *institution {
	cfg := config.New()
	NameSpace := institutionReqBody{
		Country:     country,
		Name:        name,
		Nemo:        nemo,
		Email:       email,
		State:       state,
		Description: description,
		FlagDec:     flag_dec,
		Dni:         dni,
		Operator:    cfg.WsOper(country),
	}

	response := soapCall(
		country,
		"autentia-instit.cgi",
		"wsinstit",
		requestWsCreateInstit{
			NameSpace: institutionRequest{
				WsUser:      cfg.WsUser(country),
				WsPass:      cfg.WsPass(country),
				Institution: NameSpace,
			},
		})

	var resp institutionResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}

	return resp.Body.Response.Institution
}

func UpdateInstitution(
	country string,
	name string,
	nemo string,
	email string,
	state int,
	description string,
	flag_dec int,
	dni string,
) *institution {
	cfg := config.New()
	NameSpace := institutionReqBody{
		Country:     country,
		Name:        name,
		Email:       email,
		State:       state,
		Description: description,
		FlagDec:     flag_dec,
		Dni:         dni,
		Operator:    cfg.WsOper(country),
	}

	response := soapCall(
		country,
		"autentia-instit.cgi",
		"wsinstit",
		requestWsUpdateInstit{
			NameSpace: institutionRequest{
				WsUser:      cfg.WsUser(country),
				WsPass:      cfg.WsPass(country),
				Institution: NameSpace,
			},
		})

	var resp institutionResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}

	return resp.Body.Response.Institution
}
