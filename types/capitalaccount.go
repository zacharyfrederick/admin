package types

type CapitalAccount struct {
	DocType             string        `json:"docType"`
	ID                  string        `json:"id"`
	Fund                string        `json:"fund"`
	Investor            string        `json:"name"`
	Number              int           `json:"number"`
	CurrentPeriod       int           `json:"currentPeriod"`
	PeriodClosingValue  string        `json:"periodClosingValue"`
	PeriodOpeningValue  string        `json:"periodOpeningValue"`
	FixedFees           string        `json:"fixedFees"`
	Deposits            string        `json:"deposits"`
	OwnershipPercentage string        `json:"ownershipPercentage"`
	HighWaterMark       HighWaterMark `json:"highWaterMark"`
}

type CapitalAccountAction struct {
	DocType     string `json:"docType"`
	ID          string `json:"id"`
	Type        string `json:"type"`
	Amount      string `json:"amount"`
	Full        bool   `json:"full"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Date        string `json:"Date"`
	Period      int    `json:"period"`
}

type HighWaterMark struct {
	Amount string `json:"amount"`
	Date   string `json:"date"`
}
