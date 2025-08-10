package main

import (
	"CodeStream/src"
	"CodeStream/src/api"
	"CodeStream/src/resources"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	src.Config.SetupEnv()
	resources.SetupRedis()

	if err := os.MkdirAll(src.Config.CodeWorkDir, 0755); err != nil {
		panic(err)
	}

	gin.SetMode(src.Config.ApplicationMode)
	ginEngine := gin.Default()
	ginEngine.LoadHTMLFiles("templates/index.html", "templates/ground.html")
	ginEngine.GET("/", api.HomeMenu)
	ginEngine.GET("/session/:sessionID", api.StartSession)
	ginEngine.POST("/session", api.CreateSession)
	ginEngine.GET("/ws", api.LiveStreamCoding)

	s := &http.Server{
		Addr:           ":" + src.Config.Port,
		Handler:        ginEngine,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	_ = s.ListenAndServe()

}
