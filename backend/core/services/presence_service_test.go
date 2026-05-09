package services

import (
	"gotest/core/models"
	"testing"
	"time"
)

func TestIsPresenceOnlineUsesTTL(t *testing.T) {
	now := time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		presence models.UserPresence
		want     bool
	}{
		{
			name:     "recent online heartbeat is online",
			presence: models.UserPresence{Status: "online", LastSeenAt: now.Add(-30 * time.Second)},
			want:     true,
		},
		{
			name:     "expired heartbeat is offline",
			presence: models.UserPresence{Status: "online", LastSeenAt: now.Add(-46 * time.Second)},
			want:     false,
		},
		{
			name:     "explicit offline is offline",
			presence: models.UserPresence{Status: "offline", LastSeenAt: now},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPresenceOnline(tt.presence, now); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
