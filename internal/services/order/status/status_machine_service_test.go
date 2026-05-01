package status

import (
	"testing"
)

func TestIsValidTransition(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		newStatus     string
		want          bool
	}{
		{"Pending to Confirmed", Pending, Confirmed, true},
		{"Pending to Cancelled", Pending, Cancelled, true},
		{"Confirmed to Shipped", Confirmed, Shipped, true},
		{"Confirmed to Cancelled", Confirmed, Cancelled, true},
		{"Shipped to Delivered", Shipped, Delivered, true},
		{"Pending to Shipped", Pending, Shipped, false},
		{"Shipped to Cancelled", Shipped, Cancelled, false},
		{"Delivered to any", Delivered, Confirmed, false},
		{"Cancelled to any", Cancelled, Pending, false},
		{"Unknown status", "UNKNOWN", Pending, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidTransition(tt.currentStatus, tt.newStatus); got != tt.want {
				t.Errorf("IsValidTransition(%v, %v) = %v, want %v", tt.currentStatus, tt.newStatus, got, tt.want)
			}
		})
	}
}

func TestGetNextStatus(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		want          string
	}{
		{"Next from Pending", Pending, Confirmed},
		{"Next from Confirmed", Confirmed, Shipped},
		{"Next from Shipped", Shipped, Delivered},
		{"Next from Delivered", Delivered, Delivered},
		{"Next from Cancelled", Cancelled, Cancelled},
		{"Next from Unknown", "UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNextStatus(tt.currentStatus); got != tt.want {
				t.Errorf("GetNextStatus(%v) = %v, want %v", tt.currentStatus, got, tt.want)
			}
		})
	}
}
