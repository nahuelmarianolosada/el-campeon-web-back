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
		{"Pending to Approved", Pending, Approved, true},
		{"Pending to Cancelled", Pending, Cancelled, true},
		{"Pending to Rejected", Pending, Rejected, true},
		{"Confirmed to Refunded", Approved, Refunded, true},
		{"Approved to Cancelled", Approved, Cancelled, true},
		{"Pending to Refunded", Pending, Refunded, false},
		{"Approved to Approved", Approved, Approved, false},
		{"Rejected to any", Rejected, Approved, false},
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
		{"Next from Pending", Pending, Approved},
		{"Next from Approved", Approved, Cancelled},
		{"Next from Refunded", Refunded, Refunded},
		{"Next from Cancelled", Cancelled, Cancelled},
		{"Next from Rejected", Rejected, Rejected},
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
