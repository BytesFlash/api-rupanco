package services

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/imedcl/manager-api/pkg/config"
)

type listUsersRequest struct {
	XMLName     xml.Name `xml:"urn:Req"`
	WsUser      string   `xml:"wsUsuario"`
	WsPass      string   `xml:"wsClave"`
	Country     string   `xml:"Pais"`
	Institution string   `xml:"CodInstit"`
	Operator    string   `xml:"RutOper"`
	Table       string   `xml:"Tabla"`
	Action      string   `xml:"Accion"`
	FlagDec     int      `xml:"FlagDec"`
	Offset      int      `xml:"LimitOffset"`
	Limit       int      `xml:"LimitCount"`
	Check       int      `xml:"ChkVigencia"`
}

type addRoleRequest struct {
	XMLName     xml.Name `xml:"urn:Req"`
	WsUser      string   `xml:"wsUsuario"`
	WsPass      string   `xml:"wsClave"`
	Country     string   `xml:"Pais"`
	Institution string   `xml:"CodInstit"`
	Operator    string   `xml:"RutOper"`
	Table       string   `xml:"Tabla"`
	Action      string   `xml:"Accion"`
	Dni         string   `xml:"Rut"`
	Role        string   `xml:"Rol"`
}

type changePasswordRequest struct {
	XMLName     xml.Name `xml:"urn:Req"`
	WsUser      string   `xml:"wsUsuario"`
	WsPass      string   `xml:"wsClave"`
	Country     string   `xml:"Pais"`
	Institution string   `xml:"CodInstit"`
	Operator    string   `xml:"RutOper"`
	Table       string   `xml:"Tabla"`
	Action      string   `xml:"Accion"`
	Dni         string   `xml:"Rut"`
	System      string   `xml:"Sistema"`
	NewValue    string   `xml:"NuevoValor"`
}

type getUserRolesRequest struct {
	XMLName     xml.Name `xml:"urn:Req"`
	WsUser      string   `xml:"wsUsuario"`
	WsPass      string   `xml:"wsClave"`
	Country     string   `xml:"Pais"`
	Institution string   `xml:"CodInstit"`
	Operator    string   `xml:"RutOper"`
	Table       string   `xml:"Tabla"`
	Action      string   `xml:"Accion"`
	Dni         string   `xml:"Rut"`
	Role        string   `xml:"Rol"`
	Check       int      `xml:"ChkVigencia"`
}

type addRequest struct {
	XMLName   xml.Name `xml:"urn:wsadmusr"`
	NameSpace addRoleRequest
}

type passwordRequest struct {
	XMLName   xml.Name `xml:"urn:wsadmusr"`
	NameSpace changePasswordRequest
}

type request struct {
	XMLName   xml.Name `xml:"urn:wsadmusr"`
	NameSpace listUsersRequest
}

type getRolesRequest struct {
	XMLName   xml.Name `xml:"urn:wsadmusr"`
	NameSpace getUserRolesRequest
}

type usersListResponse struct {
	NameSpace listUsersRequest
}

// Response
type soapResponse struct {
	Body responseBody `xml:"Body"`
}
type responseBody struct {
	Response responseData `xml:"CAdmUsrResp"`
}
type responseData struct {
	Result struct {
		Error  string `xml:"Err" json:"error"`
		Detail string `xml:"Glosa" json:"glosa"`
	} `xml:"Resultado" json:"result"`
	Rows int    `xml:"nRows"`
	List []list `xml:"List" json:"institutions"`
}

type list struct {
	Country        string `xml:"Pais" json:"country"`
	Institution    string `xml:"CodInstit" json:"institution"`
	RutInstitution string `xml:"RutInstit" json:"dni_institution"`
	Place          string `xml:"CodLugar" json:"place"`
	System         string `xml:"Sistema" json:"system"`
	Rut            string `xml:"Rut" json:"dni"`
	Role           string `xml:"Rol" json:"role"`
	Email          string `xml:"email" json:"email"`
	Phone          string `xml:"celular" json:"phone"`
	RoleFrom       string `xml:"RolDesde" json:"role_from"`
	RoleTo         string `xml:"RolHasta" json:"role_to"`
	PasswordDate   string `xml:"FecClave" json:"password_date"`
	FlagDec        string `xml:"FlagDec" json:"flag_dec"`
	Name           string `xml:"Nombre" json:"name"`
}

const cgi_name = "autentia-admusr4.cgi"
const cgi_namespace = "wsadmusr"

func ListUsersRoles(country string, institution string, offset int, limit int) *responseData {
	cfg := config.New()
	namespace := listUsersRequest{
		WsUser:      cfg.WsUser(country),
		WsPass:      cfg.WsPass(country),
		Country:     country,
		Institution: institution,
		Operator:    cfg.WsOper(country),
		Table:       config.READ,
		Action:      config.LIST,
		Check:       1,
		FlagDec:     -1,
		Offset:      offset,
		Limit:       limit * 10,
	}
	response := soapCall(
		namespace.Country,
		cgi_name,
		cgi_namespace,
		request{
			NameSpace: namespace,
		})

	var resp soapResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}
	return &resp.Body.Response
}

func ListUserRoles(dni string, country string, institution string) *[]list {
	cfg := config.New()
	codInstitution := "(todos)"
	if institution != "" {
		codInstitution = institution
	}
	if country == config.COLOMBIA && !strings.Contains(dni, "-C") {
		dni = fmt.Sprintf("%s-C", dni)
	}
	namespace := getUserRolesRequest{
		WsUser:      cfg.WsUser(country),
		WsPass:      cfg.WsPass(country),
		Country:     country,
		Dni:         dni,
		Institution: codInstitution,
		Check:       1,
		Operator:    cfg.WsOper(country),
		Table:       config.READ,
		Action:      config.LIST,
	}
	response := soapCall(
		namespace.Country,
		cgi_name,
		cgi_namespace,
		getRolesRequest{
			NameSpace: namespace,
		})

	var resp soapResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}
	return &resp.Body.Response.List
}

func CreateRole(dni string, name string, country string, institution string) bool {
	cfg := config.New()
	if country == config.COLOMBIA && !strings.Contains(dni, "-C") {
		dni = fmt.Sprintf("%s-C", dni)
	}
	namespace := addRoleRequest{
		WsUser:      cfg.WsUser(country),
		WsPass:      cfg.WsPass(country),
		Country:     country,
		Institution: institution,
		Dni:         dni,
		Role:        name,
		Operator:    cfg.WsOper(country),
		Table:       config.READ,
		Action:      config.ADD,
	}
	response := soapCall(
		namespace.Country,
		cgi_name,
		cgi_namespace,
		addRequest{
			NameSpace: namespace,
		})

	var resp soapResponse
	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}
	return resp.Body.Response.Result.Error == "0"
}

func DeleteRole(dni string, name string, country string, institution string) bool {
	cfg := config.New()
	namespace := addRoleRequest{
		WsUser:      cfg.WsUser(country),
		WsPass:      cfg.WsPass(country),
		Country:     country,
		Institution: institution,
		Dni:         dni,
		Role:        name,
		Operator:    cfg.WsOper(country),
		Table:       config.READ,
		Action:      config.DELETE,
	}
	response := soapCall(
		namespace.Country,
		cgi_name,
		cgi_namespace,
		addRequest{
			NameSpace: namespace,
		})

	var resp soapResponse
	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}

	fmt.Println("RESP DELETE ROLE:", resp)
	return resp.Body.Response.Result.Error == "0"
}

func ChangePassword(dni string, password string, country string, institution string, system string) bool {
	cfg := config.New()
	namespace := changePasswordRequest{
		WsUser:      cfg.WsUser(country),
		WsPass:      cfg.WsPass(country),
		Country:     country,
		Institution: institution,
		Dni:         dni,
		NewValue:    password,
		System:      system,
		Operator:    cfg.WsOper(country),
		Table:       config.CHANGE,
		Action:      config.PASSWORD,
	}
	response := soapCall(
		namespace.Country,
		cgi_name,
		cgi_namespace,
		passwordRequest{
			NameSpace: namespace,
		})

	var resp soapResponse
	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}
	return resp.Body.Response.Result.Error == "0"
}
