package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
)

type responseDevice struct {
	Data []*data.Device `json:"data"`
}

type paramsDevice struct {
	Name     string `form:"name" json:"name"`
	Datajson string `form:"datajson" json:"datajson"`
}

// @Summary post device
// @Description post device
// @Tags device
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 400 {object} MessageResponse
// @Router /device [post]
// @Param device body paramsDevice true "device"
func DevicePostRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//POST Device
	router.POST("/device", func(c *gin.Context) {
		var params paramsDevice
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

		dataJson := []byte(params.Datajson)
		var objmap map[string]interface{}
		if err := json.Unmarshal(dataJson, &objmap); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		device := &data.Device{
			Name: params.Name,
			Data: objmap,
		}

		db.CreateDevice(device)
		PrtyParams, _ := events.PrettyParams(params)
		event := &events.EventLog{
			UserNickname: usuario.NickName,
			Resource:     "Devices",
			Event:        fmt.Sprintf("Se crea Dispositivo %s", params.Name),
			Params:       PrtyParams,
		}
		event.Write()
		c.JSON(http.StatusOK, MessageResponse{
			Details: "Dispositivo creado con éxito",
			Code:    http.StatusOK,
		})

	})
}

// @Summary get device
// @Description get device
// @Tags device
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} MessageResponse
// @failure 404 {object} MessageResponse
// @Router /device [get]
func DeviceGetRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	//Get all Devices
	router.GET("/device", func(c *gin.Context) {
		role, err := db.GetAllDevice()
		if err != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusNotFound,
			})
			logrus.Println("error al obtener devices")
			return
		}
		c.JSON(http.StatusOK, role)
	})

}
