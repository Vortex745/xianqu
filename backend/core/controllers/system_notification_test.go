package controllers

import (
	"strings"
	"testing"

	"gotest/core/models"
)

func TestBuildShipmentNotification(t *testing.T) {
	product := models.Product{ID: 7, Name: "机械键盘"}
	order := models.Order{ID: 11, SellerID: 3, ProductID: product.ID, Price: 129.5, Status: 2}

	got := buildShipmentNotification(order, product)

	if got.UserID != order.SellerID {
		t.Fatalf("expected seller %d to receive notification, got %d", order.SellerID, got.UserID)
	}
	if got.Type != notificationTypeShipment {
		t.Fatalf("expected type %q, got %q", notificationTypeShipment, got.Type)
	}
	if got.OrderID == nil || *got.OrderID != order.ID {
		t.Fatalf("expected order id %d, got %#v", order.ID, got.OrderID)
	}
	if got.ProductID == nil || *got.ProductID != product.ID {
		t.Fatalf("expected product id %d, got %#v", product.ID, got.ProductID)
	}
	if got.IsRead {
		t.Fatalf("new shipment notification should be unread")
	}
	if !strings.Contains(got.Content, product.Name) || !strings.Contains(got.Content, "发货") {
		t.Fatalf("content should mention product and shipment action, got %q", got.Content)
	}
}

func TestShouldNotifyProductTakedown(t *testing.T) {
	tests := []struct {
		name      string
		oldStatus int
		newStatus int
		want      bool
	}{
		{name: "active to takedown", oldStatus: 1, newStatus: productStatusViolationTakedown, want: true},
		{name: "sold to takedown", oldStatus: 2, newStatus: productStatusViolationTakedown, want: true},
		{name: "same takedown does not duplicate", oldStatus: productStatusViolationTakedown, newStatus: productStatusViolationTakedown, want: false},
		{name: "other status does not notify", oldStatus: 1, newStatus: 2, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldNotifyProductTakedown(tt.oldStatus, tt.newStatus); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestBuildProductTakedownNotification(t *testing.T) {
	product := models.Product{ID: 15, UserID: 6, Name: "复古相机"}

	got := buildProductTakedownNotification(product)

	if got.UserID != product.UserID {
		t.Fatalf("expected owner %d to receive notification, got %d", product.UserID, got.UserID)
	}
	if got.Type != notificationTypeProductTakedown {
		t.Fatalf("expected type %q, got %q", notificationTypeProductTakedown, got.Type)
	}
	if got.ProductID == nil || *got.ProductID != product.ID {
		t.Fatalf("expected product id %d, got %#v", product.ID, got.ProductID)
	}
	if got.OrderID != nil {
		t.Fatalf("takedown notification should not reference order, got %#v", got.OrderID)
	}
	if got.IsRead {
		t.Fatalf("new takedown notification should be unread")
	}
	if !strings.Contains(got.Content, product.Name) || !strings.Contains(got.Content, "下架") {
		t.Fatalf("content should mention product and takedown action, got %q", got.Content)
	}
}
