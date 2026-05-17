package status

import "log"

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
	log.Printf("[paymentStatus.IsValidTransition] INFO: Validating transition - current=%s, new=%s", currentStatus, newStatus)
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
	log.Printf("[paymentStatus.GetNextStatus] INFO: Getting next status - current=%s", currentStatus)
	if len(validTransitions[currentStatus]) == 0 {
		return currentStatus
	}
	return validTransitions[currentStatus][0]
}
