package services

import (
	"encoding/xml"
	"fmt"

	"github.com/imedcl/manager-api/pkg/config"
)

//add

type requestWsRecAccess struct {
	XMLName   xml.Name `xml:"urn:Insert"`
	NameSpace resourceAddUserRequest
}

type resourceAddUserRequest struct {
	XMLName   xml.Name `xml:"InsertReq"`
	WsUser    string   `xml:"wsUsuario"`
	WsPass    string   `xml:"wsClave"`
	RecAccess resourceInsertReqRequest
}

type resourceInsertReqRequest struct {
	XMLName       xml.Name `xml:"RecAccess"`
	Usuario       string   `xml:"Usuario"`
	Recurso       string   `xml:"Recurso"`
	Instituciones string   `xml:"Instituciones"`
	Clave         string   `xml:"Clave"`
	Opcion1       string   `xml:"Opcion1"`
	Opcion2       string   `xml:"Opcion2"`
	Opcion3       string   `xml:"Opcion3"`
	Host          string   `xml:"Host"`
}

// List

type requesListWsRecAccess struct {
	XMLName   xml.Name `xml:"urn:List"`
	NameSpace resourceListRequest
}

type resourceListRequest struct {
	XMLName xml.Name `xml:"ListReq"`
	WsUser  string   `xml:"wsUsuario"`
	WsPass  string   `xml:"wsClave"`
	Usuario string   `xml:"Usuario,omitempty"`
	Recurso string   `xml:"Recurso,omitempty"`
}

// Response Add

type addResourceResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName    xml.Name `xml:"Body"`
		CResultado struct {
			XMLName xml.Name `xml:"CResultado"`
			Err     int      `xml:"Err" json:"error"`
			Glosa   string   `xml:"Glosa" json:"glosa"`
			NHost   int      `xml:"nHost" json:"nHost"`
			NHostOk int      `xml:"nHostOk" json:"nHostOk"`
			NRows   int      `xml:"nRows" json:"nRows"`
		} `xml:"CResultado" json:"Result"`
	} `xml:"Body" json:"Body"`
}

// Response List

type listResourceResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName  xml.Name `xml:"Body"`
		ListResp struct {
			XMLName xml.Name `xml:"ListResp"`
			Res     struct {
				XMLName xml.Name `xml:"Res"`
				Err     int      `xml:"Err" json:"error"`
				Glosa   string   `xml:"Glosa" json:"glosa"`
				NHost   int      `xml:"nHost" json:"nHost"`
				NHostOk int      `xml:"nHostOk" json:"nHostOk"`
				NRows   int      `xml:"nRows" json:"nRows"`
			} `xml:"Res" json:"Res"`
			Count int `xml:"Count" json:"count"`
			Recs  []struct {
				XMLName       xml.Name `xml:"Recs"`
				Usuario       string   `xml:"Usuario"`
				Recurso       string   `xml:"Recurso"`
				Instituciones string   `xml:"Instituciones"`
				Clave         string   `xml:"Clave"`
				Opcion1       int      `xml:"Opcion1"`
				Opcion2       int      `xml:"Opcion2"`
				Opcion3       int      `xml:"Opcion3"`
				Host          int      `xml:"Host"`
			} `xml:"Recs" json:"Recs"`
		} `xml:"ListResp" json:"ListResp"`
	} `xml:"Body" json:"Body"`
}

// Send Email

type requesClaveWsRecAccess struct {
	XMLName   xml.Name `xml:"urn:List"`
	NameSpace resourceClaveRequest
}

type resourceClaveRequest struct {
	XMLName xml.Name `xml:"ListReq"`
	WsUser  string   `xml:"wsUsuario"`
	WsPass  string   `xml:"wsClave"`
	Usuario string   `xml:"Usuario,omitempty"`
	Recurso string   `xml:"Recurso,omitempty"`
	Opcion  string   `xml:"Opcion,omitempty"`
}

func CreateResource(
	country string,
	usuario string,
	recurso string,
	instituciones string,
	clave string,
	opcion1 string,
	opcion2 string,
	opcion3 string,
	host string,
) *addResourceResponse {

	cfg := config.New()
	NameSpace := resourceAddUserRequest{
		WsUser: cfg.WsUser(country),
		WsPass: cfg.WsPass(country),
		RecAccess: resourceInsertReqRequest{
			Usuario:       usuario,
			Recurso:       recurso,
			Instituciones: instituciones,
			Clave:         clave,
			Opcion1:       opcion1,
			Opcion2:       opcion2,
			Opcion3:       opcion3,
			Host:          host,
		},
	}

	response := soapCall(
		country,
		"autentia-recaccess.cgi",
		"wsrecaccess",
		requestWsRecAccess{
			NameSpace: NameSpace,
		})

	var resp addResourceResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}

	return &resp
}

func ListResource(
	country string,
	usuario string,
	recurso string,

) *listResourceResponse {

	cfg := config.New()
	NameSpace := resourceListRequest{
		WsUser:  cfg.WsUser(country),
		WsPass:  cfg.WsPass(country),
		Usuario: usuario,
		Recurso: recurso,
	}

	response := soapCall(
		country,
		"autentia-recaccess.cgi",
		"wsrecaccess",
		requesListWsRecAccess{
			NameSpace: NameSpace,
		})
	var resp listResourceResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}

	return &resp
}

func SendMessage(
	country string,
	usuario string,
	recurso string,
	opcion string,
) *listResourceResponse {

	cfg := config.New()
	NameSpace := resourceClaveRequest{
		WsUser:  cfg.WsUser(country),
		WsPass:  cfg.WsPass(country),
		Usuario: usuario,
		Recurso: recurso,
		Opcion:  opcion,
	}

	response := soapCall(
		country,
		"autentia-recaccess.cgi",
		"wsrecaccess",
		requesClaveWsRecAccess{
			NameSpace: NameSpace,
		})

	var resp listResourceResponse

	err := xml.Unmarshal(response, &resp)
	if err != nil {
		fmt.Println("Error!", err.Error())
	}

	return &resp
}
