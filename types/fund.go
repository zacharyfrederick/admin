package types

type Fund struct {
	DocType            string `json:"docType"`
	ID                 string `json:"id"`
	Name               string `json:"name"`
	CurrentPeriod      int    `json:"currentPeriod"`
	InceptionDate      string `json:"inceptionDate"`
	PeriodClosingValue string `json:"periodClosingValue"`
	PeriodOpeningValue string `json:"periodOpeningValue"`
	AggregateFixedFees string `json:"aggregateFixedFees"`
	AggregateDeposits  string `json:"aggregateDeposits"`
	NextInvestorNumber int    `json:"nextInvestorNumber"`
	PeriodUpdated      bool   `json:"periodUpdated"`
}
