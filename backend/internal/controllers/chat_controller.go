package controllers

import (
	"fmt"
	"gotest/config"
	"gotest/internal/middleware"
	"gotest/internal/models"
	"gotest/pkg/ws"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatController struct {
	Hub *ws.Hub
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Connect 建立 WebSocket 连接
func (cc *ChatController) Connect(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 不能为空"})
		return
	}

	claims, err := middleware.ParseToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket Upgrade Error:", err)
		return
	}

	client := &ws.Client{
		Hub:    cc.Hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: claims.UserID,
	}

	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

// GetHistory 获取历史消息
func (cc *ChatController) GetHistory(c *gin.Context) {
	uid, _ := c.Get("userID")
	userID := uid.(uint)

	targetIDStr := c.Query("target_id")
	targetID, _ := strconv.Atoi(targetIDStr)

	var messages []models.Message

	err := config.DB.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, targetID, targetID, userID,
	).Order("created_at asc").Find(&messages).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取消息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}

// GetContacts 获取最近联系人列表 (真实逻辑修复版)
func (cc *ChatController) GetContacts(c *gin.Context) {
	uid, _ := c.Get("userID")
	userID := uid.(uint) // 现在这里使用了 userID，不会报 unused error

	// 1. 查出所有与我相关的消息 (按时间倒序)
	var messages []models.Message
	config.DB.Where("sender_id = ? OR receiver_id = ?", userID, userID).
		Order("created_at desc").
		Find(&messages)

	// 2. 逻辑去重
	contactMap := make(map[uint]bool)
	var contacts []map[string]interface{}

	for _, msg := range messages {
		targetID := msg.SenderID
		if msg.SenderID == userID {
			targetID = msg.ReceiverID
		}

		if contactMap[targetID] {
			continue
		}
		contactMap[targetID] = true

		var user models.User
		config.DB.First(&user, targetID)

		contacts = append(contacts, map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"last_msg": msg.Content,
			"time":     msg.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": contacts})
}
