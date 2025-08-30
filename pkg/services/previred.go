package services

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

type ejecutaRequest struct {
	XMLName       xml.Name `xml:"can:ejecuta"`
	EncodingStyle string   `xml:"soapenv:encodingStyle,attr"`
	XML           cdataXML `xml:"xml"`
}

type cdataXML struct {
	XMLContent string `xml:",cdata"`
}

type soapEnvelope struct {
	XMLName xml.Name        `xml:"soapenv:Envelope"`
	Xmlns1  string          `xml:"xmlns:soapenv,attr"`
	Xmlns2  string          `xml:"xmlns:can,attr"`
	Header  *struct{}       `xml:"soapenv:Header"`
	Body    soapBodyEjecuta `xml:"soapenv:Body"`
}

type soapBodyEjecuta struct {
	Ejecuta ejecutaRequest
}

type SoapResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Response struct {
			Return string `xml:"ejecutaReturn"`
		} `xml:"ejecutaResponse"`
	} `xml:"Body"`
}

type RespuestaPrevired struct {
	XMLName xml.Name `xml:"respuesta"`
	Control struct {
		Codigo string `xml:"codigo,attr"`
	} `xml:"respuestaservicio>control"`

	RespuestaAUT struct {
		Llave string `xml:"llave"`
	} `xml:"respuestaservicio>respuestaaut"`

	RespuestaCCX struct {
		Encabezado struct {
			Rut             string `xml:"rut,attr"`
			Nombres         string `xml:"nombres,attr"`
			ApellidoPaterno string `xml:"apellidopaterno,attr"`
			ApellidoMaterno string `xml:"apellidomaterno,attr"`
		} `xml:"respuestaccx_encabezado"`

		Detalles []struct {
			Mes                   string `xml:"mes,attr"`
			TipoMovimiento        string `xml:"tipomovimiento,attr"`
			FechaPago             string `xml:"fechapago,attr"`
			RemuneracionImponible int    `xml:"remuneracionimponible,attr"`
			Monto                 int    `xml:"monto,attr"`
			RutEmpleador          string `xml:"rutempleador,attr"`
			Afp                   string `xml:"afp,attr"`
		} `xml:"respuestaccx_detalle"`
	} `xml:"respuestaservicio>respuestaccx"`
}

func generarXMLPrevired(usuario, password, rut, codigo, tipoServicio string) string {
	return fmt.Sprintf(`<peticion llave=''>
<peticionservicio tipo='AUT'>
  <parametro valor='%s' nombre='usuario'/>
  <parametro valor='%s' nombre='password'/>
</peticionservicio>
<peticionservicio tipo='CCX'>
  <parametro nombre='rut' valor='%s'/>
  <parametro nombre='codautoriza' valor='%s'/>
  <parametro nombre="tipoServicio" valor="%s" />
</peticionservicio>
</peticion>`, usuario, password, rut, codigo, tipoServicio)
}

func EnviarPeticionPrevired(usuario, password, rut, codAutoriza, tipoServicio string) ([]byte, error) {
	xmlContent := generarXMLPrevired(usuario, password, rut, codAutoriza, tipoServicio)

	request := soapEnvelope{
		Xmlns1: "http://schemas.xmlsoap.org/soap/envelope/",
		Xmlns2: "http://canales.monitorservicios.previred.com",
		Header: &struct{}{},
		Body: soapBodyEjecuta{
			Ejecuta: ejecutaRequest{
				EncodingStyle: "http://schemas.xmlsoap.org/soap/encoding/",
				XML:           cdataXML{XMLContent: xmlContent},
			},
		},
	}

	buffer := new(bytes.Buffer)
	if err := xml.NewEncoder(buffer).Encode(request); err != nil {
		return nil, fmt.Errorf("error codificando XML: %w", err)
	}

	req, err := http.NewRequest("POST", "https://cotizaciones.previred.com/axis/services/MonitorPrevired", buffer)
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud HTTP: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", "ejecuta")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error enviando solicitud: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
