package types

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/shopspring/decimal"
	"github.com/zacharyfrederick/admin/types/doctypes"
)

type CapitalAccount struct {
	DocType             string            `json:"docType"`
	ID                  string            `json:"id"`
	Fund                string            `json:"fund"`
	Investor            string            `json:"investor"`
	Number              int               `json:"number"`
	CurrentPeriod       int               `json:"currentPeriod"`
	ClosingValue        map[string]string `json:"periodClosingValue"`
	OpeningValue        map[string]string `json:"periodOpeningValue"`
	FixedFees           map[string]string `json:"fixedFees"`
	Deposits            map[string]string `json:"deposits"`
	OwnershipPercentage map[string]string `json:"ownershipPercentage"`
	HighWaterMark       HighWaterMark     `json:"highWaterMark"`
	PeriodUpdated       bool              `json:"periodUpdated"`
	FixedFee            string            `json:"fixedFee"`
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

func (c *CapitalAccount) CurrentPeriodAsString() string {
	return fmt.Sprintf("%d", c.CurrentPeriod)
}

func (c *CapitalAccount) PreviousPeriodAsString() string {
	return fmt.Sprintf("%d", c.CurrentPeriod-1)
}

func (c *CapitalAccount) BootstrapAccountValues(openingValue string) {
	currentPeriod := c.CurrentPeriodAsString()
	c.Deposits[currentPeriod] = openingValue
	c.OpeningValue[currentPeriod] = openingValue
	c.CurrentPeriod += 1
}

func (c *CapitalAccount) IncrementCurrentPeriod() {
	c.CurrentPeriod += 1
}

func CreateDefaultCapitalAccount(nextInvestorNumber int, currentPeriod int, accountId string, fundId string, investorId string) CapitalAccount {
	closingValueMap := make(map[string]string)
	closingValueMap["0"] = decimal.Zero.String()
	openingValueMap := make(map[string]string)
	openingValueMap["0"] = decimal.Zero.String()
	fixedFeesMap := make(map[string]string)
	fixedFeesMap["0"] = decimal.Zero.String()
	depositMap := make(map[string]string)
	depositMap["0"] = decimal.Zero.String()
	ownershipPercentageMap := make(map[string]string)
	ownershipPercentageMap["0"] = decimal.Zero.String()
	capitalAccount := CapitalAccount{
		DocType:             doctypes.DOCTYPE_CAPITALACCOUNT,
		ID:                  accountId,
		Fund:                fundId,
		Investor:            investorId,
		Number:              nextInvestorNumber,
		CurrentPeriod:       currentPeriod,
		ClosingValue:        closingValueMap,
		OpeningValue:        openingValueMap,
		FixedFees:           fixedFeesMap,
		Deposits:            depositMap,
		OwnershipPercentage: ownershipPercentageMap,
		HighWaterMark:       HighWaterMark{Amount: "0.0", Date: "None"},
		PeriodUpdated:       false,
		FixedFee:            "0.02",
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

func CreateDefaultCapitalAccountAction(transactionId string, capitalAccountId string, type_ string, amount string, full bool, date string, period int) CapitalAccountAction {
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
