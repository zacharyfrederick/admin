package types

type CapitalAccount struct {
	DocType             string            `json:"docType"`
	ID                  string            `json:"id"`
	Fund                string            `json:"fund"`
	Investor            string            `json:"name"`
	Number              int               `json:"number"`
	CurrentPeriod       int               `json:"currentPeriod"`
	ClosingValue        map[string]string `json:"periodClosingValue"`
	OpeningValue        map[string]string `json:"periodOpeningValue"`
	FixedFees           map[string]string `json:"fixedFees"`
	Deposits            map[string]string `json:"deposits"`
	OwnershipPercentage map[string]string `json:"ownershipPercentage"`
	HighWaterMark       HighWaterMark     `json:"highWaterMark"`
	PeriodUpdated       bool              `json:"periodUpdated"`
}

type CapitalAccountAction struct {
	DocType        string `json:"docType"`
	ID             string `json:"id"`
	CapitalAccount string `json:"capitalAccount"`
	Type           string `json:"type"`
	Amount         string `json:"amount"`
	Full           bool   `json:"full"`
	Status         string `json:"status"`
	Description    string `json:"description"`
	Date           string `json:"Date"`
	Period         int    `json:"period"`
}

type HighWaterMark struct {
	Amount string `json:"amount"`
	Date   string `json:"date"`
}

type CreateCapitalAccountRequest struct {
	Fund     string `json:"fund" binding:"required"`
	Investor string `json:"investor" binding:"required"`
}

func ValidateCreateCapitalAccountRequest(r *CreateCapitalAccountRequest) bool {
	return true
}

type CreateCapitalAccountActionRequest struct {
	CapitalAccount string `json:"capitalAccount" binding:"required"`
	Type           string `json:"type" binding:"required"`
	Amount         string `json:"amount" binding:"required"`
	Full           bool   `json:"full" binding:"isdefault|required"`
	Date           string `json:"date" binding:"required"`
	Period         int    `json:"period" binding:"isdefault|required"`
}

func ValidateCreateCapitalAccountActionRequest(r *CreateCapitalAccountActionRequest) bool {
	return true
}
