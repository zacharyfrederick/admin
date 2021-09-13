package types

import (
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/zacharyfrederick/admin/types/doctypes"
)

type Investor struct {
	DocType string `json:"docType"`
	ID      string `json:"id"`
	Name    string `json:"name"`
}

func (f *Investor) GetID() string {
	return f.ID
}

func (f *Investor) ToJSON() ([]byte, error) {
	investorJSON, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return investorJSON, nil
}

func (f *Investor) FromJSON(data []byte) error {
	err := json.Unmarshal(data, f)
	if err != nil {
		return err
	}
	return nil
}

func (t *Investor) SaveState(ctx contractapi.TransactionContextInterface) error {
	investorJSON, err := t.ToJSON()
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(t.ID, investorJSON)
}

type CreateInvestorRequest struct {
	Name string `json:"name" binding:"required"`
}

func ValidateCreateInvestorRequest(r *CreateInvestorRequest) bool {
	return true
}

func CreateDefaultInvestor(investorId string, name string) Investor {
	investor := Investor{
		DocType: doctypes.DOCTYPE_INVESTOR,
		Name:    name,
		ID:      investorId,
	}
	return investor
}

func CreateInvestorFromJSON(data []byte) (*Investor, error) {
	var investor Investor
	err := json.Unmarshal(data, &investor)
	if err != nil {
		return nil, err
	}
	return &investor, nil
}
