package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	pingController "github.com/dev-au/CodeStream/internal/adapter/http/ping"
	roomController "github.com/dev-au/CodeStream/internal/adapter/http/room"
	"github.com/dev-au/CodeStream/internal/config"
	"github.com/dev-au/CodeStream/pkg/logs"
)

type Server struct {
	Engine     *gin.Engine
	httpServer *http.Server
	appCfg     *config.AppConfig
	httpCfg    *config.HTTPConfig
	pingCtrl   *pingController.Controller
	roomCtrl   *roomController.Controller
	wsUpgrader *websocket.Upgrader
}

const (
	RequestIDHeader = "X-Request-ID"
	requestIDKey    = "request_id"
)

func requestIDMiddleware(c *gin.Context) {
	requestID := c.GetHeader(RequestIDHeader)
	if requestID == "" {
		requestID = uuid.New().String()
	}
	c.Set(requestIDKey, requestID)

	detachedCtx := context.WithoutCancel(c.Request.Context())
	ctx := logs.ContextWithRequestID(detachedCtx, requestID)

	c.Request = c.Request.WithContext(ctx)
	c.Header(RequestIDHeader, requestID)
	c.Next()
}

func zapLogger(cfg *config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		duration := end.Sub(start)

		if cfg.Mode == gin.DebugMode {
			logs.InfoCtx(c.Request.Context(), "HTTP Request",
				"status", c.Writer.Status(),
				"method", c.Request.Method,
				"path", path,
				"query", query,
				"ip", c.ClientIP(),
				"user-agent", c.Request.UserAgent(),
				"duration", duration,
			)
		} else {
			logs.InfoCtx(c.Request.Context(), "HTTP Request",
				"status", c.Writer.Status(),
				"method", c.Request.Method,
				"path", path,
				"duration", duration,
				"query", query,
				"responseSize", c.Writer.Size(),
			)
		}
	}
}

func zapRecovery(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			logs.ErrorCtx(c.Request.Context(), "Panic recovered",
				"error", err,
				"path", c.Request.URL.Path,
			)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}()
	c.Next()
}

func NewWebsocketUpgrader() *websocket.Upgrader {
	return &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 10 * 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

func NewServer(
	appCfg *config.AppConfig,
	httpCfg *config.HTTPConfig,
	pingCtrl *pingController.Controller,
	roomCtrl *roomController.Controller,
) (*Server, func(), error) {
	gin.SetMode(appCfg.Mode)
	engine := gin.New()

	engine.Use(requestIDMiddleware)
	engine.Use(zapRecovery)
	engine.Use(zapLogger(appCfg))

	var wsUpgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 10 * 1024,
		CheckOrigin: func(r *http.Request) bool {
			logs.Debug("Hi brother")
			return true
		},
	}

	server := &Server{
		appCfg:     appCfg,
		httpCfg:    httpCfg,
		Engine:     engine,
		pingCtrl:   pingCtrl,
		roomCtrl:   roomCtrl,
		wsUpgrader: wsUpgrader,
	}

	cleanupServer := func() {
		logs.Info("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if server.httpServer != nil {
			_ = server.httpServer.Shutdown(ctx)
		}
	}

	return server, cleanupServer, nil
}

func (s *Server) Run() error {
	s.registerRoutes()
	s.httpServer = &http.Server{
		Addr:    ":" + strconv.Itoa(s.httpCfg.Port),
		Handler: s.Engine,
		ReadHeaderTimeout: 1 * time.Second,
	}
	return s.httpServer.ListenAndServe()
}
