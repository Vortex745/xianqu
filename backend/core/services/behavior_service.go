package services

import (
	"strings"
	"time"

	"gotest/core/models"

	"gorm.io/gorm"
)

const behaviorDedupeWindow = 10 * time.Minute

var behaviorWeightMap = map[string]float64{
	"view": 1,
	"like": 2,
	"cart": 3,
	"buy":  5,
}

func NormalizeBehaviorAction(action string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(action))
	_, ok := behaviorWeightMap[key]
	return key, ok
}

func BehaviorWeight(action string) float64 {
	key, ok := NormalizeBehaviorAction(action)
	if !ok {
		return 0
	}
	return behaviorWeightMap[key]
}

// RecordUserBehavior 写入用户行为并执行 10 分钟去重。
func RecordUserBehavior(db *gorm.DB, userID, productID uint, action, source string) error {
	normalizedAction, ok := NormalizeBehaviorAction(action)
	if !ok || db == nil || userID == 0 || productID == 0 {
		return nil
	}

	threshold := time.Now().Add(-behaviorDedupeWindow)

	var existingID uint
	err := db.Model(&models.UserBehavior{}).
		Select("id").
		Where("user_id = ? AND product_id = ? AND action = ? AND created_at >= ?", userID, productID, normalizedAction, threshold).
		Order("created_at DESC").
		Limit(1).
		Scan(&existingID).Error
	if err != nil {
		return err
	}
	if existingID > 0 {
		return nil
	}

	behavior := models.UserBehavior{
		UserID:    userID,
		ProductID: productID,
		Action:    normalizedAction,
		Weight:    BehaviorWeight(normalizedAction),
		Source:    strings.TrimSpace(source),
	}

	return db.Create(&behavior).Error
}
