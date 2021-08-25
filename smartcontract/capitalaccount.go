package smartcontract

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/zacharyfrederick/admin/types"
	"github.com/zacharyfrederick/admin/utils"
)

func (s *AdminContract) CreateCapitalAccount(ctx contractapi.TransactionContextInterface, capitalAccountId string, fundId string, investorId string) error {
	idInUse, err := utils.AssetExists(ctx, capitalAccountId)

	if err != nil {
		return err
	}

	if idInUse {
		return fmt.Errorf("an object with the specified Id already exists")
	}

	fund, err := s.QueryFundById(ctx, fundId)
	if err != nil {
		return err
	}

	if fund == nil {
		return fmt.Errorf("a fund with id '%s' does not exist", fundId)
	}

	investor, err := s.QueryInvestorById(ctx, investorId)
	if err != nil {
		return err
	}
	if investor == nil {
		return fmt.Errorf("an investor with id '%s' does not exist", investorId)
	}

	capitalAccount := types.CapitalAccount{
		DocType:             types.DOCTYPE_CAPITALACCOUNT,
		ID:                  capitalAccountId,
		Fund:                fundId,
		Investor:            investorId,
		Number:              fund.NextInvestorNumber,
		CurrentPeriod:       fund.CurrentPeriod,
		PeriodClosingValue:  "0.0",
		PeriodOpeningValue:  "0.0",
		FixedFees:           "0.0",
		Deposits:            "0.0",
		OwnershipPercentage: "0.0",
		HighWaterMark:       types.HighWaterMark{Amount: "0.0", Date: "None"},
	}

	capitalAccountJson, err := json.Marshal(capitalAccount)
	if err != nil {
		return err
	}

	fund.NextInvestorNumber += 1
	fundJson, err := json.Marshal(fund)
	if err != nil {
		return err
	}

	ctx.GetStub().PutState(fund.ID, fundJson)
	return ctx.GetStub().PutState(capitalAccountId, capitalAccountJson)
}

func (s *AdminContract) CreateCapitalAccountAction(ctx contractapi.TransactionContextInterface, transactionId string, capitalAccountId string, type_ string, amount string, full bool, date string, period int) error {
	if type_ != "deposit" && type_ != "withdrawal" {
		return fmt.Errorf("the specified type of '%s' is invalid for a CapitalAccountAction", type_)
	}

	capitalAccount, err := s.QueryCapitalAccountById(ctx, capitalAccountId)
	if err != nil {
		return err
	}

	if capitalAccount == nil {
		return fmt.Errorf("a capital account with that id does not exist")
	}

	capitalAccountAction := types.CapitalAccountAction{
		DocType:        types.DOCTYPE_CAPITALACCOUNTACTION,
		ID:             transactionId,
		CapitalAccount: capitalAccountId,
		Type:           type_,
		Amount:         amount,
		Full:           full,
		Status:         types.TX_STATUS_SUBMITTED,
		Description:    "",
		Date:           date,
		Period:         period,
	}

	capitalAccountActionJson, err := json.Marshal(capitalAccountAction)

	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(transactionId, capitalAccountActionJson)
}

func (s *AdminContract) QueryCapitalAccountById(ctx contractapi.TransactionContextInterface, capitalAccountId string) (*types.CapitalAccount, error) {
	data, err := ctx.GetStub().GetState(capitalAccountId)

	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var capitalAccount types.CapitalAccount
	err = json.Unmarshal(data, &capitalAccount)
	if err != nil {
		return nil, err
	}

	return &capitalAccount, nil

}

func (s *AdminContract) QueryCapitalAccountsByInvestor(ctx contractapi.TransactionContextInterface, fundId string, investorId string) ([]*types.CapitalAccount, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccount", "fund": "%s", "investor": "%s"}}`, fundId, investorId)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var capitalAccounts []*types.CapitalAccount
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var capitalAccount types.CapitalAccount
		err = json.Unmarshal(queryResult.Value, &capitalAccount)
		if err != nil {
			return nil, err
		}
		capitalAccounts = append(capitalAccounts, &capitalAccount)
	}

	return capitalAccounts, nil
}

func (s *AdminContract) QueryCapitalAccountsByFund(ctx contractapi.TransactionContextInterface, fundId string) ([]*types.CapitalAccount, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccount", "fund": "%s"}}`, fundId)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var capitalAccounts []*types.CapitalAccount
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var capitalAccount types.CapitalAccount
		err = json.Unmarshal(queryResult.Value, &capitalAccount)
		if err != nil {
			return nil, err
		}
		capitalAccounts = append(capitalAccounts, &capitalAccount)
	}

	return capitalAccounts, nil
}

func (s *AdminContract) QueryCapitalAccountActionsByFund(ctx contractapi.TransactionContextInterface, fundId string) ([]*types.CapitalAccountAction, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccount", "fund": "%s"}}`, fundId)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var capitalAccountActions []*types.CapitalAccountAction
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var capitalAccountAction types.CapitalAccountAction
		err = json.Unmarshal(queryResult.Value, &capitalAccountAction)
		if err != nil {
			return nil, err
		}
		capitalAccountActions = append(capitalAccountActions, &capitalAccountAction)
	}

	return capitalAccountActions, nil
}

func (s *AdminContract) QueryCapitalAccountActionsByAccountPeriod(ctx contractapi.TransactionContextInterface, fundId string, capitalAccountId string, period int) ([]*types.CapitalAccountAction, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccount", "fund": "%s", "capitalAccount": "%s", "period": "%d"}}`, fundId, capitalAccountId, period)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var capitalAccountActions []*types.CapitalAccountAction
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var capitalAccountAction types.CapitalAccountAction
		err = json.Unmarshal(queryResult.Value, &capitalAccountAction)
		if err != nil {
			return nil, err
		}
		capitalAccountActions = append(capitalAccountActions, &capitalAccountAction)
	}

	return capitalAccountActions, nil
}
