package routes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/imedcl/manager-api/docs"
	"github.com/imedcl/manager-api/pkg/config"
	"github.com/imedcl/manager-api/pkg/data"
	"github.com/imedcl/manager-api/pkg/events"
	"github.com/imedcl/manager-api/pkg/services"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	db             *data.DB
	connectionAuth data.ConnectionAuth
	currentUser    *data.User
)

func setTraceID(ctx *gin.Context) {
	if ctx.GetHeader("trace-ID") != "" {
		ctx.Set("trace-ID", ctx.GetHeader("trace-ID"))
	} else {
		id := uuid.New()
		ctx.Set("trace-ID", id)
	}
	ctx.Next()
}

func GetUserFromToken(c *gin.Context) (*data.User, error) {
	tokenString := c.GetHeader("Authorization")
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	if tokenString == "" {
		return nil, fmt.Errorf("Token no proporcionado")
	}

	cfg := config.New()
	claims := &Claims{}

	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.SignKey()), nil
	})

	if err != nil || !parsedToken.Valid {
		return nil, fmt.Errorf("Token inválido")
	}

	return claims.User, nil
}

func logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[%s] %s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.Keys["trace-ID"],
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

func Create(cfg *config.Config) {
	// force colors in terminal
	gin.ForceConsoleColor()

	// Register a route.
	router := gin.New()

	// CORS Configuration
	origin := strings.TrimSuffix(cfg.AppUrl(), "/")
	if gin.Mode() == "debug" {
		origin = "*"
	}
	// Temporal old domain accept: Remove when migration is finished
	oldDomain := "https://autentia-admin-dev.autentia.io"

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{origin, oldDomain},
		AllowMethods:     []string{"PUT", "OPTIONS", "GET", "POST", "DELETE"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "accept", "origin", "Access-Control-Allow-Origin"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add Trace ID to context
	router.Use(setTraceID)

	// Logger
	router.Use(logger())

	// Recovery from server error
	router.Use(gin.Recovery())

	// Migrations of DB
	migrateDB(cfg)
	logrus.Println("Migration complete!")
	// Add health check route
	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	logrus.Println("Healthz registered!")

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public Routes
	LoginRoute(router, db)
	LoginMicrosoftRoute(router, db)
	LoginGuestRoute(router, db)
	RecoveryRoute(router, db)
	ReconfirmationRoute(router, db)
	ConfirmRoute(router, db)
	ValidateRoute(router, db)
	ActivateRoute(router, db)

	//Codigo de lugar

	router.Static("/upload", "./upload")

	// Private Routes
	authorized := router.Group("/", VerifyToken)

	//user
	UserRolesRoute(authorized, db)
	UsersRolesRoute(authorized, db)
	UserPostRoute(authorized, db)
	UsersRolesEmailRoute(authorized, db)
	UserInstrospectionRoute(authorized, db)
	UserGetRoute(authorized, db)
	UserGetRouteNickname(authorized, db)
	UserPutRoute(authorized, db)
	UserRolesPostRoute(authorized, db)
	UserRolesPutRoute(authorized, db)
	UserRolesGetRoute(authorized, db)
	UserRolesDeleteRoute(authorized, db)
	UserFeedbackRoute(authorized, db)
	LogoutRoute(authorized, db)
	DeleteUserRoute(authorized, db)

	//module
	ModulesGetRoute(authorized, db)
	ModulesDeleteRoute(authorized, db)

	//log
	LogGetRoute(authorized, db)
	LogPostRoute(authorized, db)

	logrus.Println("Authorized routes!")

	logrus.Println("Autentia routes!")

	// Start the server
	logrus.Println("Start server!")
	if err := router.Run(cfg.Port()); err != nil {
		logrus.Fatalf("%v", err)
	}
}

func migrateDB(cfg *config.Config) {
	connectionAuth.Database = cfg.DbName()
	connectionAuth.UserName = cfg.DbUserName()
	connectionAuth.Password = cfg.DbPassword()
	connectionAuth.Port = cfg.DbPort()
	connectionAuth.Host = cfg.DbHost()
	connectionAuth.SSL = cfg.DbSSL()
	connectionAuth.SSLCa = cfg.DbSSLCa()
	connectionAuth.SSLCert = cfg.DbSSLCert()
	connectionAuth.SSLKey = cfg.DbSSLKey()
	connectionAuth.TimeZone = cfg.DbTimeZone()
	db, _ = connectionAuth.Connect()
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
	err := data.Migrate(*db)
	if err != nil {
		logrus.Print("Error al migrar", err.Error())
	}
	events.Start(db)
	db.CreateDefaultBrandsAndModels()
	db.CreateDefaultOwners()
	InitUserSuperAdminManager()

	countries, countriesErr := db.ListAllCountries()

	if countriesErr != nil {
		logrus.Print("Error al cargar países")
	} else {
		for _, country := range countries {
			logrus.Printf("Sync country: %s", country.Name)
			_ = syncInstitutions(country)
		}
	}

}

type Claims struct {
	User *data.User `json:"user"`
	jwt.StandardClaims
}

type MessageResponse struct {
	Code    int    `json:"code"`
	Details string `json:"details"`
}

type PasswordMessageResponse struct {
	Code         int                       `json:"code"`
	Details      string                    `json:"details"`
	Requirements config.PasswordValidation `json:"requirements"`
}

func VerifyToken(c *gin.Context) {
	currentUser = &data.User{}
	r := c.Request
	cfg := config.New()
	reqToken := r.Header.Get("Authorization")
	if reqToken == "" {
		c.JSON(http.StatusBadRequest, MessageResponse{
			Details: "Usuario no autorizado",
			Code:    http.StatusBadRequest,
		})
		c.Abort()
		return
	}
	splitToken := strings.Split(reqToken, "Bearer ")
	if len(splitToken) <= 1 {
		c.JSON(http.StatusBadRequest, MessageResponse{
			Details: "Usuario no autorizado",
			Code:    http.StatusBadRequest,
		})
		c.Abort()
		return
	}
	reqToken = splitToken[1]

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(reqToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.SignKey()), nil
	})
	if err == nil && token.Valid {
		currentUser = claims.User
		if db.IsActive(currentUser.ID) {
			var userData, _ = db.GetUserByID(currentUser.ID)
			if userData.LastTokenSession != reqToken {
				c.JSON(http.StatusUnauthorized, MessageResponse{
					Details: "Usuario no autorizado",
					Code:    http.StatusUnauthorized,
				})
				c.Abort()
				return
			}
			db.ExtendToken(currentUser)
			c.Next()
			return
		} else {
			logrus.Print("user inactive")
			c.JSON(http.StatusUnauthorized, config.SetError("Sesión expirada"))
			c.Abort()
			return
		}
	} else {
		logrus.Print("invalid token")
		c.JSON(http.StatusUnauthorized, config.SetError("Token inválido"))
		c.Abort()
		return
	}
}
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func syncInstitutions(country *data.Country) (institutionList []*data.Institution) {
	list := services.ListInstitutions(country.Name)

	for _, institution := range list {
		var inst *data.Institution
		if !db.InstitutionExists(institution.Name, country.ID) {
			logrus.Printf("Creando institución: %s", institution.Name)
			owner, _ := db.GetOwner(config.DEFAULT_OWNER)
			inst = &data.Institution{
				Name:        institution.Name,
				Nemo:        institution.Nemo,
				Country:     country,
				Dni:         strings.ToLower(institution.Rut),
				Description: institution.Description,
				Email:       institution.Email,
				FlagDec:     institution.FlagDec,
				State:       institution.State,
				Owner:       owner,
			}
			db.CreateLocalInstitution(inst)
			institutionList = append(institutionList, inst)

		} else {
			logrus.Printf("Obteniendo institución: %s", institution.Name)
			inst, _ = db.GetInstitution(institution.Name, country.ID)
			if inst.OwnerID == "" {
				inst.Owner, _ = db.GetOwner(config.DEFAULT_OWNER)
			} else {
				inst.Owner, _ = db.GetOwnerByID(inst.OwnerID)
			}
			db.AddConfigOwner(inst)
			institutionList = append(institutionList, inst)
		}

	}

	return
}

func createRoleModule(moduleName string, functionName string, roleName string) {
	var role *data.Role
	module, _ := db.CreateDefaultModule(moduleName)
	if roleName == "SUPER ADMIN-MANAGER" || roleName == "ADMIN MANAGER (PAÍS)" || roleName == "Operador lector externo" || roleName == "INVITADO" {
		role, _ = db.CreateDefaultRole(roleName)
	}

	_, err := db.CreateDefaultFunction(functionName, module.ID)
	if err != nil {
		logrus.Print("Error al crear default function", err.Error())
	}
	if roleName == "SUPER ADMIN-MANAGER" || roleName == "ADMIN MANAGER (PAÍS)" || roleName == "Operador lector externo" || roleName == "INVITADO" {
		_, err = db.CreateDefaultRoleModule(role.ID, module.ID)
		if err != nil {
			logrus.Print("Error al crear default module", err.Error())
		}
	}
}

func InitUserSuperAdminManager() {
	createRoleModule("Roles MGR", "Nuevo Rol MGR", "SUPER ADMIN-MANAGER")
	createRoleModule("Roles MGR", "Actualizar Rol MGR", "SUPER ADMIN-MANAGER")
	createRoleModule("Roles Autentia", "Nuevo Rol Autentia", "SUPER ADMIN-MANAGER")
	createRoleModule("Roles Autentia", "Actualizar Rol Autentia", "SUPER ADMIN-MANAGER")
	createRoleModule("Usuarios MGR", "Nuevo Usuario MGR (país)", "SUPER ADMIN-MANAGER")
	createRoleModule("Usuarios MGR", "Actualizar Usuario MGR (país)", "SUPER ADMIN-MANAGER")
	createRoleModule("Usuarios MGR", "Nuevo Usuario MGR (país)", "ADMIN MANAGER (PAÍS)")
	createRoleModule("Usuarios MGR", "Actualizar Usuario MGR (país)", "ADMIN MANAGER (PAÍS)")
	createRoleModule("Instituciones", "Nueva Institución", "Operador Instituciones")
	createRoleModule("Instituciones", "Actualizar Institución", "Operador Instituciones")
	createRoleModule("Datos Personas (DP)", "Actualizar DP", "Operador Datos Persona")
	createRoleModule("Datos Personas (DP)", "Auditar DP", "Operador Datos Persona")
	createRoleModule("Lectores", "Nuevo Lector", "Operador Lectores")
	createRoleModule("Lectores", "Actualizar Lector", "Operador Lectores")
	createRoleModule("Lectores", "Carga masiva Lectores", "Operador Lectores")
	createRoleModule("Lectores", "Reporte Lectores", "Operador Lectores")
	createRoleModule("Lectores", "Consultar Lector", "Operador lector externo")
	createRoleModule("Enrol-Verif", "Enrolar usuario Autentia", "Operador enrol-verif")
	createRoleModule("Enrol-Verif", "Verificar usuario Autentia CAP", "Operador enrol-verif")
	createRoleModule("Enrol-Verif", "Pasar rut PROD a CAP", "Operador enrol-verif")
	createRoleModule("Enrol-Verif", "Reporte enrol-verif", "Operador enrol-verif")
	createRoleModule("Usuarios Sistema Autentia", "Nuevo usuario Autentia", "Jefe proyecto")
	createRoleModule("Usuarios Sistema Autentia", "Consultar usuario Autentia", "Jefe proyecto")
	createRoleModule("Operaciones en capacitación", "Eliminar Persona", "Operador Capacitación")
	createRoleModule("Operaciones en capacitación", "Pasar Rut Prod a QA", "Operador Capacitación")
	createRoleModule("LME", "Desbloqueo LME", "Operador LME")
	createRoleModule("LME", "Bloqueo LME", "Operador LME")

	createRoleModule("Lectores invitado", "Consultar Lector invitado", "INVITADO")
	createRoleModule("Enrol-Verif", "Enrolar usuario Autentia", "INVITADO")
	createRoleModule("Enrol-Verif", "Verificar usuario Autentia CAP", "INVITADO")
	createRoleModule("Enrol-Verif", "Pasar rut PROD a CAP", "INVITADO")
	createRoleModule("Enrol-Verif", "Reporte enrol-verif", "INVITADO")
	createRoleModule("Enrolamiento", "Enrolamiento contra base", "JP")
	createRoleModule("Enrolamiento", "Enrolamiento contra cedula", "JP")
	createRoleModule("Log", "Consultar Log", "Lector Log")
	createRoleModule("Brand-Model", "Brand-Model Sensor", "Operador Brand-Model")
	//Ejecutar html
	createRoleModule("TRX", "Ver Modulo Trx", "Modulo TRX")
	createRoleModule("Reporte Operaciones", "Reporte de Operaciones", "Operador Reporte-Operaciones")
	createRoleModule("Reporte Auditoría Walmart", "Reporte de Auditoría Walmart", "Operador Reporte-AuditoriaWalmart")
	createRoleModule("Vigencia de Cédula", "Vigencia de Cédula", "Vigencia de Cédula")
	createRoleModule("Servicios Autentia", "Servicios Autentia", "Servicios Autentia")
	createRoleModule("Servicios Previred", "Servicios Previred", "Servicios Previred")
	createRoleModule("Validacion Persona", "validar persona", "Operador Validar Persona")

	country, _ := db.CreateDefaultCountry("*.*")
	chile, _ := db.CreateDefaultCountry("CHILE")
	ecuador, _ := db.CreateDefaultCountry("ECUADOR")
	colombia, _ := db.CreateDefaultCountry("COLOMBIA")
	rdominicana, _ := db.CreateDefaultCountry("RDOMINICANA")
	mexico, _ := db.CreateDefaultCountry("MEXICO")

	db.Exec("UPDATE institutions SET format_pdf = ? WHERE format_pdf IS NULL OR format_pdf = ''", "default")

	db.CreateDefaultOwners()

	db.CreateDefaultInstitution(chile.ID, "*.chile*")
	db.CreateDefaultInstitution(ecuador.ID, "*.ecuador*")
	db.CreateDefaultInstitution(colombia.ID, "*.colombia*")
	db.CreateDefaultInstitution(rdominicana.ID, "*.rdominicana*")
	db.CreateDefaultInstitution(mexico.ID, "*.mexico*")

	institution, _ := db.CreateDefaultInstitution(country.ID, "*.*")

	user, _ := db.CreateDefaultUser(country.ID, "sam", "Super Admin", "superadmin@autentia.cl", "admin")
	guestUser, _ := db.CreateDefaultUser(chile.ID, "guest", "Invitado", "guest@autentia.cl", "guest")
	agenteUser, _ := db.CreateDefaultUser(chile.ID, "agente", "agente", "agente@autentia.cl", "Agente2023")

	roleAdmin, _ := db.GetRole("SUPER ADMIN-MANAGER")
	_, err := db.CreateDefaultUserRoleInstitution(user.ID, roleAdmin.ID, institution.ID)
	if err != nil {
		logrus.Print(": Error! = " + err.Error())
	}

	roleGuest, _ := db.GetRole("INVITADO")
	_, errGuest := db.CreateDefaultUserRoleInstitution(guestUser.ID, roleGuest.ID, institution.ID)
	if errGuest != nil {
		logrus.Print(": Error! = " + err.Error())
	}

	_, errAgente := db.CreateDefaultUserRoleInstitution(agenteUser.ID, roleGuest.ID, institution.ID)
	if errAgente != nil {
		logrus.Print(": Error! = " + err.Error())
	}

	db.CreateDefaultAutentiaRole("OPER")
	db.CreateDefaultAutentiaRole("ADMIN")

	/* 	result := db.Exec(`
	   	        DELETE FROM users
	   	        WHERE id NOT IN (
	   	            SELECT MIN(id)
	   	            FROM users
	   	            GROUP BY email
	   	        )
	   	    `)
	   	if result.Error != nil {
	   		panic(result.Error)
	   	}
	   	fmt.Println("Duplicated emails removed successfully.") */
}
