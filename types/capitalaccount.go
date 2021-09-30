package types

import (
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/shopspring/decimal"
	"github.com/zacharyfrederick/admin/types/doctypes"
)

type CapitalAccount struct {
	DocType             string         `json:"docType"`
	ID                  string         `json:"id"`
	Fund                string         `json:"fund"`
	Investor            string         `json:"investor"`
	Number              int            `json:"number"`
	CurrentPeriod       int            `json:"currentPeriod"`
	ClosingValue        map[int]string `json:"periodClosingValue"`
	OpeningValue        map[int]string `json:"periodOpeningValue"`
	FixedFees           map[int]string `json:"fixedFees"`
	Deposits            map[int]string `json:"deposits"`
	OwnershipPercentage map[int]string `json:"ownershipPercentage"`
	PerformanceFees     map[int]string `json:"performanceFees"`
	HighWaterMark       HighWaterMark  `json:"highWaterMark"`
	PeriodUpdated       bool           `json:"periodUpdated"`
	FixedFee            string         `json:"fixedFee"`
	HasPerformanceFees  bool           `json:"hasPerformanceFees"`
	PerformanceFeeRate  string         `json:"performanceFeeRate"`
}

func (c *CapitalAccount) UpdateClosingValue(fundClosingValue decimal.Decimal) {
	ownershipPercentage := decimal.RequireFromString(c.OwnershipPercentage[c.PreviousPeriod()])
	c.SetClosingValue(ownershipPercentage.Mul(fundClosingValue).String())
}

func (c *CapitalAccount) SetClosingValue(closingValue string) {
	c.ClosingValue[c.CurrentPeriod] = closingValue
}

func (c *CapitalAccount) UpdateOpeningValue(openingValue string) {
	c.OpeningValue[c.CurrentPeriod] = openingValue
}

func (f *CapitalAccount) GetID() string {
	return f.ID
}

func (f *CapitalAccount) ToJSON() ([]byte, error) {
	capitalAccountJSON, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return capitalAccountJSON, nil
}

func (f *CapitalAccount) FromJSON(data []byte) error {
	err := json.Unmarshal(data, f)
	if err != nil {
		return err
	}
	return nil
}

func (c *CapitalAccount) SaveState(ctx contractapi.TransactionContextInterface) error {
	capitalAccountJSON, err := c.ToJSON()
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(c.ID, capitalAccountJSON)
}

func (c *CapitalAccount) PreviousPeriod() int {
	return c.CurrentPeriod - 1
}

func (c *CapitalAccount) BootstrapAccountValues(openingValue string) {
	c.Deposits[c.CurrentPeriod] = openingValue
	c.OpeningValue[c.CurrentPeriod] = openingValue
	c.CurrentPeriod += 1
}

func (c *CapitalAccount) IncrementCurrentPeriod() {
	c.CurrentPeriod += 1
}

func CreateDefaultCapitalAccount(
	nextInvestorNumber int,
	currentPeriod int,
	accountId string,
	fundId string,
	investorId string,
	hasPerformanceFees bool,
	performanceFeeRate string,
) CapitalAccount {
	capitalAccount := CapitalAccount{
		DocType:             doctypes.DOCTYPE_CAPITALACCOUNT,
		ID:                  accountId,
		Fund:                fundId,
		Investor:            investorId,
		Number:              nextInvestorNumber,
		CurrentPeriod:       currentPeriod,
		ClosingValue:        map[int]string{0: "0"},
		OpeningValue:        map[int]string{0: "0"},
		FixedFees:           map[int]string{0: "0"},
		PerformanceFees:     map[int]string{0: "0"},
		Deposits:            map[int]string{0: "0"},
		OwnershipPercentage: map[int]string{0: "0"},
		HighWaterMark:       HighWaterMark{Amount: decimal.Zero.String(), Date: 0},
		PeriodUpdated:       false,
		FixedFee:            "0.02",
		HasPerformanceFees:  hasPerformanceFees,
		PerformanceFeeRate:  performanceFeeRate,
	}

	//if the capital account is created after the inception period of the fund we need to initialize
	//all of the tracking values to 0
	if currentPeriod > 0 {
		for i := 1; i <= currentPeriod; i++ {
			capitalAccount.ClosingValue[i] = "0"
			capitalAccount.OpeningValue[i] = "0"
			capitalAccount.FixedFees[i] = "0"
			capitalAccount.PerformanceFees[i] = "0"
			capitalAccount.Deposits[i] = "0"
			capitalAccount.OwnershipPercentage[i] = "0"
		}
	}

	return capitalAccount
}

func (c *CapitalAccountAction) SaveState(ctx contractapi.TransactionContextInterface) error {
	capitalAccountActionJSON, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(c.ID, capitalAccountActionJSON)
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

func CreateDefaultCapitalAccountAction(
	transactionId string,
	capitalAccountId string,
	type_ string,
	amount string,
	full bool,
	date string,
	period int,
) CapitalAccountAction {
	capitalAccountAction := CapitalAccountAction{
		DocType:        doctypes.DOCTYPE_CAPITALACCOUNTACTION,
		ID:             transactionId,
		CapitalAccount: capitalAccountId,
		Type:           type_,
		Amount:         amount,
		Full:           full,
		Status:         TX_STATUS_SUBMITTED,
		Description:    "",
		Date:           date,
		Period:         period,
	}
	return capitalAccountAction
}

type HighWaterMark struct {
	Amount string `json:"amount"`
	Date   int    `json:"date"`
}

type CreateCapitalAccountRequest struct {
	Fund               string `json:"fund"     binding:"required"`
	Investor           string `json:"investor" binding:"required"`
	HasPerformanceFees bool   `json:"hasPerformanceFees" binding:"isdefault|required"`
	PerformanceRate    string `json:"performanceRate" binding:"required"`
}

func ValidateCreateCapitalAccountRequest(r *CreateCapitalAccountRequest) bool {
	return true
}

type CreateCapitalAccountActionRequest struct {
	CapitalAccount string `json:"capitalAccount" binding:"required"`
	Type           string `json:"type"           binding:"required"`
	Amount         string `json:"amount"         binding:"required"`
	Full           bool   `json:"full"           binding:"isdefault|required"`
	Date           string `json:"date"           binding:"required"`
	Period         int    `json:"period"         binding:"isdefault|required"`
}

func ValidateCreateCapitalAccountActionRequest(r *CreateCapitalAccountActionRequest) bool {
	return true
}
