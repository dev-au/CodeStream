package ping

import "github.com/gin-gonic/gin"


type Controller struct {
}


func (p *Controller) Ping(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"health": "ok ishladi birodar vauuuu"})
}


func NewController() *Controller {
	return &Controller{}
}