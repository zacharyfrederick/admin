package types

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

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

func (f *Fund) CurrentPeriodAsString() string {
	return fmt.Sprintf("%d", f.CurrentPeriod)
}

func (f *Fund) PreviousPeriodAsString() string {
	return fmt.Sprintf("%d", f.CurrentPeriod-1)
}

func (f *Fund) SaveState(ctx contractapi.TransactionContextInterface) error {
	fundJson, err := json.Marshal(f)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(f.ID, fundJson)
}

func (f *Fund) BootstrapFundValues(totalDeposits string, openingFundValue string) {
	currentPeriod := f.CurrentPeriodAsString()
	f.Deposits[currentPeriod] = totalDeposits
	f.OpeningValues[currentPeriod] = openingFundValue
	f.CurrentPeriod += 1
}

type CreateFundRequest struct {
	Name          string `json:"name" binding:"required"`
	InceptionDate string `json:"inceptionDate" binding:"required"`
}

func ValidateCreateFundRequest(r *CreateFundRequest) bool {
	return true
}
