package routers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"web_app/logger"
)

func Setup() *gin.Engine {
	r:=gin.New()
	r.Use(logger.GinLogger(),logger.GinRecovery(true))

	r.GET("/", func(context *gin.Context) {
		context.String(http.StatusOK,"ok")
	})
	
	return r 
}
