package shipping

import "errors"

var (
	ErrPostalCodeNotCovered = errors.New("POSTAL_CODE_NOT_COVERED")
	ErrNoBranchHasStock     = errors.New("NO_BRANCH_HAS_STOCK")
	ErrNoRateForZone        = errors.New("NO_RATE_FOR_ZONE")
)
