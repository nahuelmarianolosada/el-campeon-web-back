package status

const (
	Pending   = "PENDING"
	Approved  = "APPROVED"
	Rejected  = "REJECTED"
	Cancelled = "CANCELLED"
	Refunded  = "REFUNDED"
)

var validTransitions = map[string][]string{
	Pending:   {Approved, Rejected, Cancelled},
	Approved:  {Cancelled, Refunded},
	Rejected:  {},
	Cancelled: {},
	Refunded:  {},
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
