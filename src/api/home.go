package api

import (
	"github.com/gin-gonic/gin"
)

func HomeMenu(c *gin.Context) {
	c.HTML(200, "index.html", gin.H{})
	return
}
