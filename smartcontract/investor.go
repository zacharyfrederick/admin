package smartcontract

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/zacharyfrederick/admin/types"
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
		DocType: types.DOCTYPE_INVESTOR,
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

func (s *AdminContract) QueryInvestorByName(ctx contractapi.TransactionContextInterface, name string) (*types.Investor, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"investor", "name": "%s"}}`, name)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var investor types.Investor
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(queryResult.Value, &investor)
		if err != nil {
			return nil, err
		}

		if true {
			break
		}
	}

	return &investor, nil
}
