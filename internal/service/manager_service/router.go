package manager_service

import (
	"github.com/donscoco/gateway_be/conf"
	"github.com/donscoco/gateway_be/internal/controller"
	"github.com/donscoco/gateway_be/internal/middleware/manager_middleware"
	docs "github.com/donscoco/gateway_be/internal/swagger" //swag init 生成的文件
	"github.com/donscoco/gateway_be/pkg/iron_config"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server celler server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @query.collection.format multi

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @securitydefinitions.oauth2.application OAuth2Application
// @tokenUrl https://example.com/oauth/token
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.implicit OAuth2Implicit
// @authorizationurl https://example.com/oauth/authorize
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.password OAuth2Password
// @tokenUrl https://example.com/oauth/token
// @scope.read Grants read access
// @scope.write Grants write access
// @scope.admin Grants read and write access to administrative information

// @securitydefinitions.oauth2.accessCode OAuth2AccessCode
// @tokenUrl https://example.com/oauth/token
// @authorizationurl https://example.com/oauth/authorize
// @scope.admin Grants read and write access to administrative information

// @x-extension-openapi {"example": "value on a json format"}

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {

	config := iron_config.Conf
	router := gin.Default()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// programatically set swagger info
	docs.SwaggerInfo.Title = config.GetString("/swagger/title")
	docs.SwaggerInfo.Description = config.GetString("/swagger/desc")
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = config.GetString("/swagger/host")
	docs.SwaggerInfo.BasePath = config.GetString("/swagger/base_path")
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	store, err := sessions.NewRedisStore(10, "tcp",
		config.GetString("/session/redis_addr"),
		config.GetString("/session/redis_pwd"),
		[]byte("secret"),
	)
	if err != nil {
		log.Fatalf("sessions.NewRedisStore err:%v", err)
	}

	adminLoginRouter := router.Group("/admin_login")
	adminLoginRouter.Use(
		sessions.Sessions("mysession", store),
		manager_middleware.RecoveryMiddleware(),
		manager_middleware.RequestLog(),
		manager_middleware.TranslationMiddleware(), // 设置翻译
	)
	{
		controller.AdminLoginRegister(adminLoginRouter)
	}

	adminRouter := router.Group("/admin")
	adminRouter.Use(
		sessions.Sessions("mysession", store), // 设置cookie 名
		manager_middleware.RecoveryMiddleware(),
		manager_middleware.RequestLog(),
		manager_middleware.SessionAuthMiddleware(),
		manager_middleware.TranslationMiddleware())
	{
		controller.AdminRegister(adminRouter)
	}

	serviceRouter := router.Group("/service")
	serviceRouter.Use(
		sessions.Sessions("mysession", store),
		manager_middleware.RecoveryMiddleware(),
		manager_middleware.RequestLog(),
		manager_middleware.SessionAuthMiddleware(),
		manager_middleware.TranslationMiddleware())
	{
		controller.ServiceRegister(serviceRouter)
	}

	appRouter := router.Group("/app")
	appRouter.Use(
		sessions.Sessions("mysession", store),
		manager_middleware.RecoveryMiddleware(),
		manager_middleware.RequestLog(),
		manager_middleware.SessionAuthMiddleware(),
		manager_middleware.TranslationMiddleware())
	{
		controller.APPRegister(appRouter)
	}

	dashRouter := router.Group("/dashboard")
	dashRouter.Use(
		sessions.Sessions("mysession", store),
		manager_middleware.RecoveryMiddleware(),
		manager_middleware.RequestLog(),
		manager_middleware.SessionAuthMiddleware(),
		manager_middleware.TranslationMiddleware())
	{
		controller.DashboardRegister(dashRouter)
	}

	// 前端静态资源
	//router.Static("/dist", "./dist")

	router.Static("/dist", conf.Path("../dist")) // dist 目录在相对 gataway_be/conf 的 ../

	return router
}
