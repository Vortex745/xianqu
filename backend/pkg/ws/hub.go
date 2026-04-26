package ws

import (
	"encoding/json"
	"gotest/core/models"
	"log"
	"sync"
)

// Hub 维护所有活跃连接并广播消息
// 注意：Client 结构体定义在同包下的 client.go 文件中，这里不需要重复定义
type Hub struct {
	mu sync.RWMutex

	// 注册的客户端 map[Client指针]bool
	Clients map[*Client]bool

	// 用户索引 map[UserID]Client指针
	// 用于通过 UserID 快速找到对应的连接，实现点对点发消息
	UserClients map[uint]*Client

	// 广播通道：接收来自 Client 的 []byte 消息
	Broadcast chan []byte

	// 注册通道
	Register chan *Client

	// 注销通道
	Unregister chan *Client
}

// IsOnline reports whether a user currently has an active WebSocket client.
func (h *Hub) IsOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.UserClients[userID]
	return ok
}

// NewHub 初始化 Hub
func NewHub() *Hub {
	return &Hub{
		Broadcast:   make(chan []byte),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[*Client]bool),
		UserClients: make(map[uint]*Client),
	}
}

// Run 启动 Hub 的主循环 (在一个单独的 goroutine 中运行)
func (h *Hub) Run() {
	for {
		select {
		// 1. 处理用户上线
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.UserClients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("WS: 用户 %d 上线", client.UserID)

		// 2. 处理用户下线
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				delete(h.UserClients, client.UserID)
				close(client.Send) // 关闭通道，通知 WritePump 退出
				log.Printf("WS: 用户 %d 下线", client.UserID)
			}
			h.mu.Unlock()

		// 3. 处理消息转发
		case message := <-h.Broadcast:
			// 这里收到的 message 已经是 client.go 存入数据库后发来的 JSON 字节流
			// 我们需要解析它，看看是发给谁的
			var msgModel models.Message
			if err := json.Unmarshal(message, &msgModel); err != nil {
				log.Println("WS: 广播消息解析失败:", err)
				continue
			}

			// A. 发送给接收者 (如果在线)
			h.mu.RLock()
			receiver, receiverOnline := h.UserClients[msgModel.ReceiverID]
			sender, senderOnline := h.UserClients[msgModel.SenderID]
			h.mu.RUnlock()

			if receiverOnline {
				select {
				case receiver.Send <- message:
				default:
					// 如果发送阻塞（对方掉线或卡死），清理连接
					h.mu.Lock()
					close(receiver.Send)
					delete(h.Clients, receiver)
					delete(h.UserClients, receiver.UserID)
					h.mu.Unlock()
				}
			}

			// B. 发回给发送者 (如果在线)
			// 这一步是为了支持“多端同步”，或者让前端确认消息已发送成功
			if senderOnline {
				select {
				case sender.Send <- message:
				default:
					h.mu.Lock()
					close(sender.Send)
					delete(h.Clients, sender)
					delete(h.UserClients, sender.UserID)
					h.mu.Unlock()
				}
			}
		}
	}
}
