package routes

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/imedcl/manager-api/pkg/config"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type OperacionesParams struct {
	Institucion string `form:"institucion" binding:"required" `
	DateStart   string `form:"date_start" binding:"required" json:"date_start"`
	DateEnd     string `form:"date_end" binding:"required" json:"date_end"`
}

type AuditoriaWalmartParams struct {
	Rut       string `form:"rut" binding:"required" `
	DateStart string `form:"date_start" binding:"required" json:"date_start"`
	DateEnd   string `form:"date_end" binding:"required" json:"date_end"`
}

type KeyBigQuery struct {
	TypeBig        string `json:"type"`
	ProjectId      string `json:"project_id"`
	PrivateKeyId   string `json:"private_key_id"`
	PrivateKey     string `json:"private_key"`
	ClientEmail    string `json:"client_email"`
	ClientId       string `json:"client_id"`
	AuthUri        string `json:"auth_uri"`
	TokenUri       string `json:"token_uri"`
	AuthProvider   string `json:"auth_provider_x509_cert_url"`
	Client_Cert    string `json:"client_x509_cert_url"`
	UniverseDomain string `json:"universe_domain"`
}

// Estructura de bigquery, se les pone NullInt64 por si vienen datos nulos, probe con int64 y devuelve todo en 0
type ResultRow struct {
	Fecha               bigquery.NullDate  `json:"fecha"`
	TotalVerificaciones bigquery.NullInt64 `json:"total_verificaciones"`
	TotalEnrolamientos  bigquery.NullInt64 `json:"total_enrolamientos"`
	TotalOperacionesInt bigquery.NullInt64 `json:"total_operaciones_internas"`
	TotalTrx            bigquery.NullInt64 `json:"total_trx"`
}

type ResultRowWalmart struct {
	Rut          bigquery.NullString `json:"rut"`
	NroAuditoria bigquery.NullString `json:"nro_auditoria"`
	Fecha        bigquery.NullString `json:"fecha_hora"`
	Resultado    bigquery.NullString `json:"resultado"`
}

func createJsonKey() {

	cfg := config.New()

	privateKeyId := cfg.PrivateKeyIdBigQuery()
	clientId := cfg.ClientIdBigQuery()
	privateKey := cfg.PrivateKeyBigQuery()
	decodedKeyBytes, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		log.Fatalf("error al decodificar la clave privada: %v", err)
	}

	decodedKey := string(decodedKeyBytes)

	jsonBigQuery := KeyBigQuery{"service_account",
		"sandbox-179114",
		privateKeyId,
		decodedKey,
		"manager-sandbox@sandbox-179114.iam.gserviceaccount.com",
		clientId,
		"https://accounts.google.com/o/oauth2/auth",
		"https://oauth2.googleapis.com/token",
		"https://www.googleapis.com/oauth2/v1/certs",
		"https://www.googleapis.com/robot/v1/metadata/x509/manager-sandbox%40sandbox-179114.iam.gserviceaccount.com",
		"googleapis.com"}

	keyJson, err := json.Marshal(jsonBigQuery)
	if err != nil {
		fmt.Printf("Error codificando jsonBigQuery: %v", err)
	}

	dir, _ := os.Getwd()
	path := dir + "/pkg/mail/templates/key.json"

	err = os.WriteFile(path, keyJson, os.ModePerm)
	if err != nil {
		fmt.Printf("Error al escribir archivo: %v", err)
	}
}

// se le agrego el  TO_JSON_STRING(STRUCT( para la data, ademas los sum buscan evitar posibles null "IFNULL" y , 0
func getBigQueryData(ctx context.Context, client *bigquery.Client, institucion string, dateStart string, dateEnd string) ([]ResultRow, error) {

	query := client.Query(`
        SELECT 
            TO_JSON_STRING(STRUCT(
                fecha, 
                SUM(IFNULL(total_verificaciones, 0)) AS total_verificaciones,
                SUM(IFNULL(total_enrolamientos, 0)) AS total_enrolamientos,
                SUM(IFNULL(total_operaciones_internas, 0)) AS total_operaciones_internas,
                SUM(IFNULL(total_trx, 0)) AS total_trx
            )) AS jso_data
        FROM
            reporteria_manager.trx_lector
        WHERE
            fecha >= @fechastart
            AND fecha < @fechafinish
            AND institucion = @institucion
        GROUP BY fecha
    `)

	query.Parameters = []bigquery.QueryParameter{
		{Name: "fechastart", Value: dateStart},
		{Name: "fechafinish", Value: dateEnd},
		{Name: "institucion", Value: institucion},
	}

	it, err := query.Read(ctx)
	if err != nil {
		return nil, err
	}

	var results []ResultRow
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var result ResultRow
		jsonStr, ok := row[0].(string)
		if !ok {
			fmt.Println("Error al convertir el valor a string")
			continue
		}
		err = json.Unmarshal([]byte(jsonStr), &result)
		if err != nil {
			fmt.Println("Error al decodificar el JSON:", err)
			continue
		}

		results = append(results, result)
	}
	return results, nil
}

func Operacionesxdia(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {

	router.POST("/reportes/operaciones", func(c *gin.Context) {
		var params OperacionesParams
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

		institucion := strings.ToUpper(params.Institucion)
		datestart := strings.ToUpper(params.DateStart)
		dateend := strings.ToUpper(params.DateEnd)

		date1, err := time.Parse("2006-01-02", datestart)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Fecha de inicio no válida",
				Code:    http.StatusBadRequest,
			})
			return
		}

		date2, err := time.Parse("2006-01-02", dateend)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Fecha final no válida",
				Code:    http.StatusBadRequest,
			})
			return
		}

		fmt.Println("Fechas:", date1, date2)
		createJsonKey()
		dir, _ := os.Getwd()
		path := dir + "/pkg/mail/templates/key.json"
		ctx := context.Background()
		client, err := bigquery.NewClient(ctx, "sandbox-179114", option.WithCredentialsFile(path))
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al crear el cliente de BigQuery",
				Code:    http.StatusBadRequest,
			})
			return
		}
		defer client.Close()

		results, err := getBigQueryData(ctx, client, institucion, datestart, dateend)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al ejecutar la consulta",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if results == nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "No existen registros",
				Code:    http.StatusBadRequest,
			})
			return
		}

		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Reporte de Operaciones",
			Event:        fmt.Sprintf("Reporte de Operaciones, Institución: %s, Fecha de Inicio: %s, Fecha de Término: %s", institucion, datestart, dateend),
			Params:       PrtyParams,
		}
		event.Write()

		c.JSON(http.StatusOK, results)

	})
}

func getDataAuditoriaWalmart(ctx context.Context, client *bigquery.Client, rut string, dateStart string, dateEnd string) ([]ResultRowWalmart, error) {
	query := client.Query(`
	SELECT 
	TO_JSON_STRING(STRUCT(
	rut,
	nro_auditoria,
	fecha_hora,
	resultado)) AS jso_data
	 from sandbox-179114.reporteria_manager.walmart_auditorias
	where rut= @rut
	and date(fecha_hora) >= @fechastart
    and date(fecha_hora) <= @fechafinish
	ORDER BY fecha_hora
    `)

	query.Parameters = []bigquery.QueryParameter{
		{Name: "rut", Value: rut},
		{Name: "fechastart", Value: dateStart},
		{Name: "fechafinish", Value: dateEnd},
	}

	it, err := query.Read(ctx)
	if err != nil {
		return nil, err
	}

	var results []ResultRowWalmart
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var result ResultRowWalmart

		jsonStr, ok := row[0].(string)
		if !ok {
			fmt.Println("Error al convertir el valor a string")
			continue
		}
		err = json.Unmarshal([]byte(jsonStr), &result)
		if err != nil {
			fmt.Println("Error al decodificar el JSON:", err)
			continue
		}

		results = append(results, result)
	}
	return results, nil
}

func AuditoriasWalmart(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {

	router.POST("/reportes/auditoriasWalmart", func(c *gin.Context) {
		var params AuditoriaWalmartParams
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

		rut := strings.ToUpper(params.Rut)
		datestart := strings.ToUpper(params.DateStart)
		dateend := strings.ToUpper(params.DateEnd)

		_, err = time.Parse("2006-01-02", datestart)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Fecha de inicio no válida",
				Code:    http.StatusBadRequest,
			})
			return
		}

		_, err = time.Parse("2006-01-02", dateend)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Fecha final no válida",
				Code:    http.StatusBadRequest,
			})
			return
		}

		createJsonKey()
		dir, _ := os.Getwd()
		path := dir + "/pkg/mail/templates/key.json"
		ctx := context.Background()
		client, err := bigquery.NewClient(ctx, "sandbox-179114", option.WithCredentialsFile(path))
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al crear el cliente de BigQuery",
				Code:    http.StatusBadRequest,
			})
			return
		}
		defer client.Close()

		results, err := getDataAuditoriaWalmart(ctx, client, rut, datestart, dateend)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Error al ejecutar la consulta",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if results == nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "No existen registros",
				Code:    http.StatusBadRequest,
			})
			return
		}

		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Reporte de Auditoría Walmart",
			Event:        fmt.Sprintf("Reporte de Auditoría Walmart, Rut: %s, Fecha de Inicio: %s, Fecha de Término: %s", rut, datestart, dateend),
			Params:       PrtyParams,
		}
		event.Write()

		c.JSON(http.StatusOK, results)

	})
}
