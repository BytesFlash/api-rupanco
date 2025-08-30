package routes

import (
	"fmt"
	"net/http"

	"github.com/Unleash/unleash-client-go/v3"
	"github.com/gin-gonic/gin"

	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
)

type roleAutentiaParams struct {
	Id   string `form:"id"`
	Name string `form:"name"`
}

type userRoleAutentiaCreateParams struct {
	UserID string               `json:"user_id" binding:"required"`
	Role   []roleAutentiaParams `json:"role_id" binding:"required"`
}

// @Summary /user/role/autentia/all
// @Description get all user role autentia
// @Tags user role autentia
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} responseUserRoleAutentia
// @failure 400 {object} MessageResponse
// @Router /user/role/autentia [get]
func UserGetAllUserRoleAutentiaRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// List user role autentia
	router.GET("/user/role/autentia/all", func(c *gin.Context) {
		idUser := c.Query("id_user")
		userRole, err := db.GetRoleAutentiabyId(idUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}
		c.JSON(http.StatusOK, userRole)
	})
}

// @Summary roleUserAutentia get
// @Description roleUserAutentia get
// @Tags roleUserAutentia
// @security BarerToken
// @Accept json
// @Produce json
// @Success 200 {object} response
// @failure 404 {object} MessageResponse
// @Router /roleUserAutentia/{id} [get]
// @param id path string true "id"
func GetRoleIdUserAutentia(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	// Get User Dec by Id
	router.GET("/roleUserAutentia", func(c *gin.Context) {
		idUser := c.Query("user_id")
		userDec, userDecErr := db.GetUserRoleAutentia(idUser)

		if userDecErr != nil {
			c.JSON(http.StatusNotFound, MessageResponse{
				Details: userDecErr.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		c.JSON(http.StatusOK, userDec)
	})
}

// @Summary UserRoleAutentia
// @Description create UserRoleAutentia
// @Tags UserRoleAutentia
// @security BarerToken
// @Accept json
// @Produce json
// @Param user body UserRoleAutentia true "UserRoleAutentia"
// @Success 201 "Rol Autentia asociado al usuario Creado"
// @failure 400 {object} MessageResponse
// @Router /UserRoleAutentia [post]
func UserRoleAutentiaPostRoute(router *gin.RouterGroup, db *data.DB, client *unleash.Client) {
	router.POST("/UserRoleAutentia", func(c *gin.Context) {

		var params userRoleAutentiaCreateParams
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

		userId := (params.UserID)
		RoleId := (params.Role)
		UserSelect, _ := db.GetUserByID(userId)

		if len(userId) == 0 {
			c.JSON(http.StatusBadRequest, MessageResponse{
				Details: "Debe contener un usuario",
				Code:    http.StatusBadRequest,
			})
			return
		}
		if len(RoleId) == 0 {
			db.DeleteAllUserRolAutentia(userId)
			PrtyParams, _ := events.PrettyParams(params)
			event := &events.EventLog{
				UserNickname: usuario.NickName,
				Resource:     "Usuario rol autentia",
				Event:        fmt.Sprintf("Se han eliminado los roles autentia, al usuario%s", UserSelect.Name),
				Params:       PrtyParams,
			}
			event.Write()
			c.JSON(http.StatusOK, MessageResponse{
				Details: "Roles eliminados del usuario con éxito",
				Code:    http.StatusOK,
			})
			return
		}

		db.DeleteAllUserRolAutentia(userId)
		for _, roleJson := range RoleId {

			userRole := &data.UserRoleAutentia{
				UserID:   userId,
				RoleName: roleJson.Name,
				RoleId:   roleJson.Id,
			}
			db.CreateUserRolAutentia(userRole)

			PrtyParams, _ := events.PrettyParams(params)
			event := &events.EventLog{
				UserNickname: usuario.NickName,
				Resource:     "Usuario rol autentia",
				Event:        fmt.Sprintf("Se ha autorizado al usuario %s, el rol autentia %s", UserSelect.Name, userRole.RoleName),
				Params:       PrtyParams,
			}
			event.Write()
		}

		c.JSON(http.StatusOK, MessageResponse{
			Details: "Usuario asociado con éxito",
			Code:    http.StatusOK,
		})
	})

}
