package types

type Fund struct {
	DocType            string            `json:"docType"`
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	CurrentPeriod      int               `json:"currentPeriod"`
	InceptionDate      string            `json:"inceptionDate"`
	ClosingValues      map[string]string `json:"periodClosingValue"`
	OpeningValues      map[string]string `json:"periodOpeningValue"`
	FixedFees          map[string]string `json:"aggregateFixedFees"`
	Deposits           map[string]string `json:"aggregateDeposits"`
	NextInvestorNumber int               `json:"nextInvestorNumber"`
	PeriodUpdated      bool              `json:"periodUpdated"`
}

type CreateFundRequest struct {
	Name          string `json:"name" binding:"required"`
	InceptionDate string `json:"inceptionDate" binding:"required"`
}

func ValidateCreateFundRequest(r *CreateFundRequest) bool {
	return true
}
