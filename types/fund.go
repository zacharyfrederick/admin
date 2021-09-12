package types

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/zacharyfrederick/admin/types/doctypes"
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

func (f *Fund) ToJSON() ([]byte, error) {
	fundJSON, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return fundJSON, nil
}

func (f *Fund) SaveState(ctx contractapi.TransactionContextInterface) error {
	fundJson, err := f.ToJSON()
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

func CreateDefaultFund(fundId string, name string, inceptionDate string) Fund {
	closingValues := make(map[string]string)
	closingValues["0"] = "0"
	openingValues := make(map[string]string)
	openingValues["0"] = "0"
	fixedFees := make(map[string]string)
	fixedFees["0"] = "0"
	deposits := make(map[string]string)
	deposits["0"] = "0"

	fund := Fund{
		DocType:            doctypes.DOCTYPE_FUND,
		ID:                 fundId,
		Name:               name,
		CurrentPeriod:      0,
		InceptionDate:      inceptionDate,
		NextInvestorNumber: 0,
		ClosingValues:      closingValues,
		OpeningValues:      openingValues,
		FixedFees:          fixedFees,
		Deposits:           deposits,
		PeriodUpdated:      false,
	}
	return fund
}

func CreateFundFromJSON(data []byte) (*Fund, error) {
	var fund Fund
	err := json.Unmarshal(data, &fund)
	if err != nil {
		return nil, err
	}
	return &fund, nil
}

type CreateFundRequest struct {
	Name          string `json:"name" binding:"required"`
	InceptionDate string `json:"inceptionDate" binding:"required"`
}

func ValidateCreateFundRequest(r *CreateFundRequest) bool {
	return true
}
