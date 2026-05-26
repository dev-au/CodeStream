package http

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type webSocketHandler func(*gin.Context, *websocket.Upgrader)

func (s *Server) wsRoute(handler webSocketHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler(ctx, s.wsUpgrader)
	}
}

func (s *Server) registerRoutes() {
	apiGroup := s.Engine.Group("/api")

	apiGroup.GET("/ping", s.pingCtrl.Ping)
	apiGroup.GET("/room", s.wsRoute(s.roomCtrl.RoomSession))
}
