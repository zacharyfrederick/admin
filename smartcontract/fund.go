package smartcontract

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/shopspring/decimal"

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

func (s *AdminContract) BootstrapFundById(ctx contractapi.TransactionContextInterface, fundId string) (*types.Fund, error) {
	fund, err := s.QueryFundById(ctx, fundId)
	if err != nil {
		return nil, err
	}
	if fund == nil {
		errorString := errors.New("A fund with that id does not exist")
		return nil, errorString
	}
	if fund.CurrentPeriod != 0 {
		errorString := errors.New("This fund cannot be bootstrapped, it is not in period 0")
		return nil, errorString
	}

	totalDeposits, err := CalculateFundDeposits(ctx, fund)
	if err != nil {
		return fund, err
	}

	fund.AggregateDeposits = totalDeposits
	fund.PeriodClosingValue = totalDeposits
	fund.CurrentPeriod = fund.CurrentPeriod + 1

	fundJson, err := json.Marshal(fund)
	if err != nil {
		return fund, err
	}

	err = ctx.GetStub().PutState(fundId, fundJson)
	if err != nil {
		return fund, err
	}

	return fund, nil
}

func CalculateFundClosingValue(ctx contractapi.TransactionContextInterface, fund *types.Fund) (string, error) {
	if fund.CurrentPeriod == 0 {
		return "0.0", nil
	}
	return "", errors.New("unimplemented") // TODO calculate the funds closing value based on market valuatiion
}

func CalculateFundFixedFees(ctx contractapi.TransactionContextInterface, fund *types.Fund) (string, error) {
	if fund.CurrentPeriod == 0 {
		return "0.0", nil
	}
	return "", errors.New("unimplemented") // TODO calculate the funds closing value based on market valuatiion
}

func CalculateFundDeposits(ctx contractapi.TransactionContextInterface, fund *types.Fund) (string, error) {
	capitalAccountActions, err := QueryCapitalAccountActionsByFundPeriod(ctx, fund.ID, fund.CurrentPeriod)
	if err != nil {
		return "", err
	}
	totalDeposits, err := AggregateDepositsFromActions(capitalAccountActions)
	if err != nil {
		return "", err
	}
	return totalDeposits, nil
}

func AggregateDepositsFromActions(actions []*types.CapitalAccountAction) (string, error) {
	total, _ := decimal.NewFromString("0.0")

	for _, action := range actions {
		if action.Type == "withdrawal" {
			continue
		}
		deposit, err := decimal.NewFromString(action.Amount)
		if err != nil {
			return "", err
		}
		total = total.Add(deposit)
	}
	return total.String(), nil
}
