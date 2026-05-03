package controllers

import "testing"

func TestNormalizeReportReason(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "trims whitespace", input: "  图片与实物不符  ", want: "图片与实物不符"},
		{name: "rejects empty", input: "   ", wantErr: true},
		{name: "rejects too short", input: "坏", wantErr: true},
		{name: "rejects too long", input: string(make([]byte, 201)), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeReportReason(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestProductWarningLevel(t *testing.T) {
	tests := []struct {
		count int64
		level string
	}{
		{count: 0, level: "green"},
		{count: 5, level: "green"},
		{count: 6, level: "yellow"},
		{count: 10, level: "yellow"},
		{count: 11, level: "red"},
		{count: 15, level: "red"},
		{count: 16, level: "red"},
	}

	for _, tt := range tests {
		if got := productWarningLevel(tt.count); got != tt.level {
			t.Fatalf("count %d: expected %q, got %q", tt.count, tt.level, got)
		}
	}
}
