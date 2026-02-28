package ws

import (
	"encoding/json"
	"gotest/config"
	"gotest/internal/models"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client 代表一个 WebSocket 连接用户
type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	UserID uint
}

// 接收前端发来的消息格式
type InputMessage struct {
	ReceiverID uint   `json:"receiver_id"`
	Content    string `json:"content"`
	Type       int    `json:"type"` // 1:文字 2:图片
}

// ReadPump 循环读取前端发来的消息
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// 1. 解析消息
		var input InputMessage
		if err := json.Unmarshal(message, &input); err != nil {
			log.Println("消息格式错误:", err)
			continue
		}

		// 2. 构造数据库模型
		msgModel := models.Message{
			SenderID:   c.UserID,
			ReceiverID: input.ReceiverID,
			Content:    input.Content,
			Type:       input.Type,
			// CreatedAt 由 GORM 自动生成
		}

		// 3. ★★★ 存入 SQLite 数据库 ★★★
		if err := config.DB.Create(&msgModel).Error; err != nil {
			log.Println("消息存库失败:", err)
			continue
		}

		// 4. 构造要广播的数据 (带上 ID 和时间，回传给前端)
		broadcastMsg, _ := json.Marshal(msgModel)

		// 5. 发送给 Hub 进行转发
		c.Hub.Broadcast <- broadcastMsg
	}
}

// WritePump 将消息发送给前端
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 如果队列里还有消息，一次性写完
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// 发送心跳包
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
