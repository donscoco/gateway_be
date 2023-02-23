package http_proxy_service

import (
	"github.com/donscoco/gateway_be/internal/controller"
	"github.com/donscoco/gateway_be/internal/middleware/manager_middleware"
	"github.com/donscoco/gateway_be/internal/middleware/proxy_middleware/http_proxy_middleware"
	"github.com/gin-gonic/gin"
)

func InitRouter(middlewares ...gin.HandlerFunc) *gin.Engine {
	//router := gin.Default()
	router := gin.New()
	router.Use(middlewares...)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	oauth := router.Group("/oauth")
	oauth.Use(manager_middleware.TranslationMiddleware())
	{
		controller.OAuthRegister(oauth)
	}

	router.Use(
		http_proxy_middleware.HTTPAccessModeMiddleware(),
		http_proxy_middleware.HTTPFlowCountMiddleware(),
		http_proxy_middleware.HTTPFlowLimitMiddleware(),
		http_proxy_middleware.HTTPJwtAuthTokenMiddleware(),
		http_proxy_middleware.HTTPJwtFlowCountMiddleware(),
		http_proxy_middleware.HTTPJwtFlowLimitMiddleware(),
		http_proxy_middleware.HTTPWhiteListMiddleware(),
		http_proxy_middleware.HTTPBlackListMiddleware(),
		http_proxy_middleware.HTTPHeaderTransferMiddleware(),
		http_proxy_middleware.HTTPStripUriMiddleware(),
		http_proxy_middleware.HTTPUrlRewriteMiddleware(),
		http_proxy_middleware.HTTPReverseProxyMiddleware())

	return router
}
