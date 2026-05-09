package services

import (
	"gotest/core/models"
	"time"

	"gorm.io/gorm"
)

const PresenceTTL = 45 * time.Second

func MarkUserOnline(db *gorm.DB, userID uint, now time.Time) error {
	presence := models.UserPresence{
		UserID:     userID,
		Status:     "online",
		LastSeenAt: now,
	}
	return db.Save(&presence).Error
}

func MarkUserOffline(db *gorm.DB, userID uint, now time.Time) error {
	presence := models.UserPresence{
		UserID:     userID,
		Status:     "offline",
		LastSeenAt: now,
	}
	return db.Save(&presence).Error
}

func IsUserRecentlyOnline(db *gorm.DB, userID uint, now time.Time) bool {
	if db == nil || userID == 0 {
		return false
	}

	var presence models.UserPresence
	if err := db.First(&presence, "user_id = ?", userID).Error; err != nil {
		return false
	}

	return IsPresenceOnline(presence, now)
}

func IsPresenceOnline(presence models.UserPresence, now time.Time) bool {
	return presence.Status == "online" && now.Sub(presence.LastSeenAt) <= PresenceTTL
}
