package smartcontract

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"

	"github.com/zacharyfrederick/admin/types"
)

func (s *AdminContract) CreateFund(ctx contractapi.TransactionContextInterface, fundId string, name string, inceptionDate string) error {
	obj, err := ctx.GetStub().GetState(fundId)
	if err != nil {
		return fmt.Errorf(("error retrieving the world state"))
	}

	if obj != nil {
		return fmt.Errorf("an object already exists with that id")
	}

	fund := types.Fund{
		DocType:            types.DOCTYPE_FUND,
		ID:                 fundId,
		Name:               name,
		CurrentPeriod:      0,
		InceptionDate:      inceptionDate,
		PeriodClosingValue: "0",
		PeriodOpeningValue: "0",
		AggregateFixedFees: "0",
		AggregateDeposits:  "0",
		NextInvestorNumber: 0,
		PeriodUpdated:      false,
	}

	fundJson, err := json.Marshal(fund)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(fundId, fundJson)
}

func (s *AdminContract) QueryFundByName(ctx contractapi.TransactionContextInterface, name string) (*types.Fund, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"fund", "name": "%s"}}`, name)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var fund types.Fund
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var fund types.Fund
		err = json.Unmarshal(queryResult.Value, &fund)
		if err != nil {
			return nil, err
		}

		if true {
			break
		}
	}

	return &fund, nil
}

func (s *AdminContract) QueryFundById(ctx contractapi.TransactionContextInterface, fundId string) (*types.Fund, error) {
	data, err := ctx.GetStub().GetState(fundId)

	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var fund types.Fund
	err = json.Unmarshal(data, &fund)
	if err != nil {
		return nil, err
	}

	return &fund, nil
}
