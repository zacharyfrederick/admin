package smartcontract

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/zacharyfrederick/admin/types"
	smartcontracterrors "github.com/zacharyfrederick/admin/types/errors"
	"github.com/zacharyfrederick/admin/utils"
)

func (s *AdminContract) CreateCapitalAccount(ctx contractapi.TransactionContextInterface, capitalAccountId string, fundId string, investorId string) error {
	idInUse, err := utils.AssetExists(ctx, capitalAccountId)
	if err != nil {
		return smartcontracterrors.ReadingWorldStateError
	}
	if idInUse {
		return smartcontracterrors.IdAlreadyInUseError
	}
	fund, err := s.QueryFundById(ctx, fundId)
	if err != nil {
		return err
	}
	if fund == nil {
		return smartcontracterrors.FundNotFoundError
	}
	fund.IncrementInvestorNumber()
	investor, err := s.QueryInvestorById(ctx, investorId)
	if err != nil {
		return smartcontracterrors.ReadingWorldStateError
	}
	if investor == nil {
		return smartcontracterrors.InvestorNotFoundError
	}
	err = SaveState(ctx, fund)
	if err != nil {
		return err
	}
	capitalAccount := types.CreateDefaultCapitalAccount(fund.NextInvestorNumber-1, fund.CurrentPeriod, capitalAccountId, fundId, investorId)
	return SaveState(ctx, &capitalAccount)
}

func (s *AdminContract) CreateCapitalAccountAction(ctx contractapi.TransactionContextInterface, transactionId string, capitalAccountId string, type_ string, amount string, full bool, date string, period int) error {
	if type_ != "deposit" && type_ != "withdrawal" {
		return smartcontracterrors.InvalidCapitalAccountActionTypeError
	}
	capitalAccount, err := s.QueryCapitalAccountById(ctx, capitalAccountId)
	if err != nil {
		return err
	}
	if capitalAccount == nil {
		return smartcontracterrors.CapitalAccountNotFoundError
	}
	capitalAccountAction := types.CreateDefaultCapitalAccountAction(transactionId, capitalAccountId, type_, amount, full, date, period)
	return capitalAccountAction.SaveState(ctx)
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
	return executeCapitalAccountQuery(ctx, queryString)
}

func (s *AdminContract) QueryCapitalAccountsByFund(ctx contractapi.TransactionContextInterface, fundId string) ([]*types.CapitalAccount, error) {
	return queryCapitalAccountsByFund(ctx, fundId)
}

func queryCapitalAccountsByFund(ctx contractapi.TransactionContextInterface, fundId string) ([]*types.CapitalAccount, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccount", "fund": "%s"}}`, fundId)
	return executeCapitalAccountQuery(ctx, queryString)
}

func (s *AdminContract) QueryCapitalAccountActionsByFund(ctx contractapi.TransactionContextInterface, fundId string) ([]*types.CapitalAccountAction, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccountAction", "fund": "%s"}}`, fundId)
	return executeCapitalAccountActionQuery(ctx, queryString)
}

func (s *AdminContract) QueryCapitalAccountActionsByFundPeriod(ctx contractapi.TransactionContextInterface, fundId string, period int) ([]*types.CapitalAccountAction, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccountAction", "fund": "%s", "period": "%d"}}`, fundId, period)
	return executeCapitalAccountActionQuery(ctx, queryString)
}

func (s *AdminContract) QueryCapitalAccountActionsByAccountPeriod(ctx contractapi.TransactionContextInterface, fundId string, capitalAccountId string, period int) ([]*types.CapitalAccountAction, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccountAction", "fund": "%s", "capitalAccount": "%s", "period": "%d"}}`, fundId, capitalAccountId, period)
	return executeCapitalAccountActionQuery(ctx, queryString)
}

func (s *AdminContract) QueryCapitalAccountActionById(ctx contractapi.TransactionContextInterface, capitalAccountId string) (*types.CapitalAccountAction, error) {
	data, err := ctx.GetStub().GetState(capitalAccountId)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	var capitalAccountAction types.CapitalAccountAction
	err = json.Unmarshal(data, &capitalAccountAction)
	if err != nil {
		return nil, err
	}
	return &capitalAccountAction, nil
}

func QueryCapitalAccountActionsByFundPeriod(ctx contractapi.TransactionContextInterface, fundId string, period int) ([]*types.CapitalAccountAction, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccountAction", "fund": "%s", "period": "%d"}}`, fundId, period)
	return executeCapitalAccountActionQuery(ctx, queryString)
}

func QueryDepositsByFundPeriod(ctx contractapi.TransactionContextInterface, fundId string, period int) ([]*types.CapitalAccountAction, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccountAction", "fund": "%s", "period": "%d", "type": "deposit"}}`, fundId, period)
	return executeCapitalAccountActionQuery(ctx, queryString)
}

func QueryWithdrawalsByFundPeriod(ctx contractapi.TransactionContextInterface, fundId string, period int) ([]*types.CapitalAccountAction, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccountAction", "fund": "%s", "period": "%d", "type": "withdrawal"}}`, fundId, period)
	return executeCapitalAccountActionQuery(ctx, queryString)
}

func QueryDepositsByFundAccountPeriod(ctx contractapi.TransactionContextInterface, capitalAccountId string, period int) ([]*types.CapitalAccountAction, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccountAction", "period": %d, "type": "deposit", "capitalAccount": "%s"}}`, period, capitalAccountId)
	return executeCapitalAccountActionQuery(ctx, queryString)
}

func QueryWithdrawalsByFundAccountPeriod(ctx contractapi.TransactionContextInterface, capitalAccountId string, period int) ([]*types.CapitalAccountAction, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccountAction", "period": %d, "type": "withdrawal", "capitalAccount": "%s"}}`, period, capitalAccountId)
	return executeCapitalAccountActionQuery(ctx, queryString)
}

func executeCapitalAccountQuery(ctx contractapi.TransactionContextInterface, queryString string) ([]*types.CapitalAccount, error) {
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

func executeCapitalAccountActionQuery(ctx contractapi.TransactionContextInterface, queryString string) ([]*types.CapitalAccountAction, error) {
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
