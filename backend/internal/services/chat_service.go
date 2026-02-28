package services

import (
	"gotest/config"
	"gotest/internal/models"
)

type ChatService struct{}

func (s *ChatService) SaveMessage(msg *models.Message) error {
	return config.DB.Create(msg).Error
}

// GetHistory 获取两人之间的聊天记录
func (s *ChatService) GetHistory(userID, targetID uint) ([]models.Message, error) {
	var messages []models.Message
	// 查询 sender=A&receiver=B 或 sender=B&receiver=A
	err := config.DB.Where(
		"(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, targetID, targetID, userID,
	).Order("created_at asc").Find(&messages).Error
	return messages, err
}

// GetContacts 获取最近联系人列表
func (s *ChatService) GetContacts(userID uint) ([]map[string]interface{}, error) {
	// 这是一个简化的实现：查询所有涉及该用户的消息，然后按对方ID去重
	var messages []models.Message
	config.DB.Where("sender_id = ? OR receiver_id = ?", userID, userID).Order("created_at desc").Find(&messages)

	contactMap := make(map[uint]models.Message)
	var contacts []map[string]interface{}

	for _, m := range messages {
		targetID := m.ReceiverID
		if m.ReceiverID == userID {
			targetID = m.SenderID
		}

		// 如果这个联系人还没加进来，就加进来（因为是倒序，第一条就是最新的）
		if _, exists := contactMap[targetID]; !exists {
			contactMap[targetID] = m

			var user models.User
			config.DB.First(&user, targetID)

			contacts = append(contacts, map[string]interface{}{
				"user":         user,
				"last_message": m,
				"timestamp":    m.CreatedAt,
			})
		}
	}
	return contacts, nil
}
