package controllers

import (
	"gotest/config"
	"gotest/core/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type PresenceController struct{}

func (pc *PresenceController) Heartbeat(c *gin.Context) {
	uid, _ := c.Get("userID")
	userID := uid.(uint)

	if err := services.MarkUserOnline(config.DB, userID, time.Now()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "在线状态更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"online": true,
	})
}

func (pc *PresenceController) Offline(c *gin.Context) {
	uid, _ := c.Get("userID")
	userID := uid.(uint)

	if err := services.MarkUserOffline(config.DB, userID, time.Now()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "离线状态更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"online": false,
	})
}
