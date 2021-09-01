package types

const TX_STATUS_SUBMITTED string = "submitted"
const TX_STATUS_PROCESSING string = "processing"
const TX_STATUS_COMPLETED string = "completed"
const TX_STATUS_ERROR string = "error"

func ValidateTransactionStatus(status string) bool {
	switch status {
	case TX_STATUS_SUBMITTED:
		return true
	case TX_STATUS_PROCESSING:
		return true
	case TX_STATUS_COMPLETED:
		return true
	case TX_STATUS_ERROR:
		return true
	default:
		return false
	}
}
