package api

import (
	"CodeStream/src/resources"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateSession(c *gin.Context) {
	var body struct {
		SessionID string `json:"session_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.SessionID == "" {
		c.JSON(400, gin.H{"error": "session_id is empty"})
		return
	}
	cache := resources.NewCacheContext()
	interview, err, created := resources.CreateInterviewSession(cache, body.SessionID)
	if !created {
		c.JSON(400, gin.H{"error": "session already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"session_id": interview.SessionID})
	return
}
