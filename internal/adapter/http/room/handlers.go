package room

import (
	"github.com/dev-au/CodeStream/pkg/logs"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// type webSession struct {
// 	domainRoom *room.Room
// 	users      []*account.User
// 	wsConns    []*websocket.Conn
// }

// var sessions = map[string]*webSession{}

type Controller struct {
}

func (c *Controller) RoomSession(ctx *gin.Context, upgrader *websocket.Upgrader) {

	// conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	// if err != nil {
	// 	logs.ErrorCtx(ctx, "Failed to upgrade connection: %v", err)
	// 	return
	// }

	// for {
		
	// }

}

func NewController() *Controller {
	return &Controller{}
}
