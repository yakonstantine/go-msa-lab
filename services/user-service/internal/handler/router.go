package handler

import (
	"github.com/gin-gonic/gin"
)

func NewUserRoutes(g *gin.RouterGroup, uh *UserHandler) {
	g.GET("/users/:corpKey", uh.GetByCorpKey)
	g.POST("/users", uh.Create)
}
