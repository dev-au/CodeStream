package api

import (
	"CodeStream/src/resources"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateSession(c *gin.Context) {
	var body struct {
		CaptchaResponse string `json:"captcha" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !resources.ValidateCaptcha(body.CaptchaResponse) {
		c.JSON(400, gin.H{"error": "Captcha error"})
		return
	}

	cache := resources.NewCacheContext()
	if !resources.CanCreateSession(cache, c.ClientIP()) {
		c.JSON(429, gin.H{"error": "Too many sessions"})
		return
	}
	interview, err, _ := resources.CreateInterviewSession(cache)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"session_id": interview.SessionID})
	return
}
