package smartcontract

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/zacharyfrederick/admin/types"
	"github.com/zacharyfrederick/admin/types/doctypes"
	"github.com/zacharyfrederick/admin/utils"
)

func (s *AdminContract) CreateInvestor(ctx contractapi.TransactionContextInterface, investorId string, name string) error {
	objExists, err := utils.AssetExists(ctx, investorId)
	if err != nil {
		return err
	}
	if objExists {
		return fmt.Errorf("an object already exists with that ID")
	}
	investor := types.Investor{
		DocType: doctypes.DOCTYPE_INVESTOR,
		Name:    name,
		ID:      investorId,
	}
	investorJson, err := json.Marshal(investor)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(investorId, investorJson)
}

func (s *AdminContract) QueryInvestorById(ctx contractapi.TransactionContextInterface, investorId string) (*types.Investor, error) {
	data, err := ctx.GetStub().GetState(investorId)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	var investor types.Investor
	err = json.Unmarshal(data, &investor)
	if err != nil {
		return nil, err
	}
	return &investor, nil
}
