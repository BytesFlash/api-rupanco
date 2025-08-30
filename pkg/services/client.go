package services

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/imedcl/manager-api/pkg/config"
)

type soapHeader struct {
	XMLName xml.Name `xml:"x:Header"`
}

type soapBody struct {
	XMLName   xml.Name `xml:"x:Body"`
	NameSpace interface{}
}

type soapRoot struct {
	XMLName xml.Name `xml:"x:Envelope"`
	X       string   `xml:"xmlns:x,attr"`
	Sch     string   `xml:"xmlns:sch,attr"`
	Urn     string   `xml:"xmlns:urn,attr"`
	Header  soapHeader
	Body    soapBody
}

type DefaultValues struct {
	WsUser   string `xml:"wsUsuario"`
	WsPass   string `xml:"wsClave"`
	Country  string `xml:"Pais"`
	Operator string `xml:"RutOper"`
}

func soapCall(country string, service string, namespace string, request interface{}) []byte {
	cfg := config.New()
	var root = soapRoot{}
	root.X = "http://schemas.xmlsoap.org/soap/envelope/"
	root.Sch = "http://www.n11.com/ws/schemas"
	root.Urn = fmt.Sprintf("urn:%s", namespace)
	root.Header = soapHeader{}
	root.Body = soapBody{}
	root.Body.NameSpace = request
	out, _ := xml.MarshalIndent(&root, " ", "  ")
	body := string(out)
	client := &http.Client{
		Transport: &http.Transport{
			IdleConnTimeout: 30 * time.Minute,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	url := fmt.Sprintf("%scgi-bin/%s%s", cfg.WsUrl(country), cfg.Path(country), service)
	response, err := client.Post(url, "text/xml", bytes.NewBufferString(body))
	if err != nil {
		fmt.Println("error", err)
	}

	defer response.Body.Close()
	content, _ := ioutil.ReadAll(response.Body)

	return content
}
