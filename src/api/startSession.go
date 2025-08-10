package api

import (
	"github.com/gin-gonic/gin"
)

func StartSession(c *gin.Context) {
	c.HTML(200, "ground.html", gin.H{})
	return
}
