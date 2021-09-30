package types

import (
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/zacharyfrederick/admin/types/doctypes"
)

type Fund struct {
	DocType              string         `json:"docType"`
	ID                   string         `json:"id"`
	Name                 string         `json:"name"`
	CurrentPeriod        int            `json:"currentPeriod"`
	InceptionDate        string         `json:"inceptionDate"`
	ClosingValues        map[int]string `json:"periodClosingValue"`
	OpeningValues        map[int]string `json:"periodOpeningValue"`
	FixedFees            map[int]string `json:"aggregateFixedFees"`
	Deposits             map[int]string `json:"aggregateDeposits"`
	PerformanceFees      map[int]string `json:"performanceFees"`
	NextInvestorNumber   int            `json:"nextInvestorNumber"`
	PeriodUpdated        bool           `json:"periodUpdated"`
	HasPerformanceFees   bool           `json:"hasPerformanceFees"`
	PerformanceFeePeriod int            `json:"performanceFeePeriod"`
	MidYearDeposits      []string       `json:"midYearDeposits"`
	MidYearWithdrawals   []string       `json:"midYearWithdrawals"`
}

func (f *Fund) IsPerformanceFeePeriod() bool {
	return f.CurrentPeriod%f.PerformanceFeePeriod == 0
}
func (f *Fund) IncrementInvestorNumber() {
	f.NextInvestorNumber += 1
}

func (f *Fund) IncrementCurrentPeriod() {
	f.CurrentPeriod += 1
}

func (f *Fund) PreviousPeriod() int {
	return f.CurrentPeriod - 1
}

func (f *Fund) GetID() string {
	return f.ID
}

func (f *Fund) ToJSON() ([]byte, error) {
	fundJSON, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return fundJSON, nil
}

func (f *Fund) FromJSON(data []byte) error {
	err := json.Unmarshal(data, f)
	if err != nil {
		return err
	}
	return nil
}

func (f *Fund) SaveState(ctx contractapi.TransactionContextInterface) error {
	fundJson, err := f.ToJSON()
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(f.ID, fundJson)
}

func (f *Fund) BootstrapFundValues(totalDeposits string, openingFundValue string) {
	f.Deposits[f.CurrentPeriod] = totalDeposits
	f.OpeningValues[f.CurrentPeriod] = openingFundValue
	f.CurrentPeriod += 1
}

func CreateDefaultFund(fundId string, name string, inceptionDate string) Fund {
	fund := Fund{
		DocType:              doctypes.DOCTYPE_FUND,
		ID:                   fundId,
		Name:                 name,
		CurrentPeriod:        0,
		InceptionDate:        inceptionDate,
		NextInvestorNumber:   0,
		ClosingValues:        map[int]string{0: "0"},
		OpeningValues:        map[int]string{0: "0"},
		FixedFees:            map[int]string{0: "0"},
		Deposits:             map[int]string{0: "0"},
		PerformanceFees:      map[int]string{0: "0"},
		PeriodUpdated:        false,
		HasPerformanceFees:   true,
		PerformanceFeePeriod: 12,
		MidYearDeposits:      make([]string, 0),
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

type FundAndCapitalAccounts struct {
	Fund     *Fund             `json:"fund"`
	Accounts []*CapitalAccount `json:"accounts"`
}
