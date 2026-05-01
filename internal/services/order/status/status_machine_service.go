package status

const (
	Pending   = "PENDING"
	Confirmed = "CONFIRMED"
	Shipped   = "SHIPPED"
	Delivered = "DELIVERED"
	Cancelled = "CANCELLED"
)

var validTransitions = map[string][]string{
	Pending:   {Confirmed, Cancelled},
	Confirmed: {Shipped, Cancelled},
	Shipped:   {Delivered},
	Delivered: {},
	Cancelled: {},
}

func IsValidTransition(currentStatus, newStatus string) bool {
	validNextStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, status := range validNextStatuses {
		if status == newStatus {
			return true
		}
	}

	return false
}

func GetNextStatus(currentStatus string) string {
	if len(validTransitions[currentStatus]) == 0 {
		return currentStatus
	}
	return validTransitions[currentStatus][0]
}
