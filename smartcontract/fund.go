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
	//initialize all values for the current period to be 0
	//the deposits are later aggregated and used to define the fund opening value when it is bootstrapped
	closingValues := make(map[string]string)
	closingValues["0"] = decimal.Zero.String()
	openingValues := make(map[string]string)
	openingValues["0"] = decimal.Zero.String()
	fixedFees := make(map[string]string)
	fixedFees["0"] = decimal.Zero.String()
	deposits := make(map[string]string)
	deposits["0"] = decimal.Zero.String()
	fund := types.Fund{
		DocType:            types.DOCTYPE_FUND,
		ID:                 fundId,
		Name:               name,
		CurrentPeriod:      0,
		InceptionDate:      inceptionDate,
		NextInvestorNumber: 0,
		ClosingValues:      closingValues,
		OpeningValues:      openingValues,
		FixedFees:          fixedFees,
		Deposits:           deposits,
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

func (s *AdminContract) BootstrapFund(ctx contractapi.TransactionContextInterface, fundId string) error {
	fund, err := s.QueryFundById(ctx, fundId)
	if err != nil {
		return nil
	}
	if fund == nil {
		return errors.New("a fund with that id does not exists")
	}

	updatedAccountValues, err := s.BootstrapCapitalAccountsForFund(ctx, fundId)
	if err != nil {
		return err
	}

	currentPeriod := fmt.Sprintf("%d", fund.CurrentPeriod)
	fund.Deposits[currentPeriod] = updatedAccountValues.TotalDeposits
	fund.OpeningValues[currentPeriod] = updatedAccountValues.OpeningFundValue
	fund.CurrentPeriod += 1

	fundJson, err := json.Marshal(fund)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(fund.ID, fundJson)
	if err != nil {
		return err
	}
	return nil
}

func (s *AdminContract) BootstrapCapitalAccountsForFund(ctx contractapi.TransactionContextInterface, fundId string) (*BootstrappedFundValues, error) {
	accounts, err := queryCapitalAccountsByFund(ctx, fundId)
	if err != nil {
		return &BootstrappedFundValues{}, err
	}
	//loop over aggregate deposits and track closing fund value
	openingFundValue := decimal.Zero
	totalDeposits := decimal.Zero
	for _, account := range accounts {
		err := s.BootstrapCapitalAccount(ctx, account)
		if err != nil {
			return &BootstrappedFundValues{}, errors.New("err from bootstrap capital account")
		}
		currentPeriod := fmt.Sprintf("%d", account.CurrentPeriod-1) //we updated the current period already
		deposit, err := decimal.NewFromString(account.Deposits[currentPeriod])
		if err != nil {
			return &BootstrappedFundValues{}, err
		}
		totalDeposits = totalDeposits.Add(deposit)
		openingFundValue = openingFundValue.Add(deposit)
	}
	//update the ownership percentage for each account based on the closing fund value
	for _, account := range accounts {
		err := updateCapitalAccountOwnership(ctx, account, openingFundValue)
		if err != nil {
			return &BootstrappedFundValues{}, err
		}
		accountJson, err := json.Marshal(account)
		if err != nil {
			return &BootstrappedFundValues{}, err
		}
		ctx.GetStub().PutState(account.ID, accountJson)
	}
	retValue := &BootstrappedFundValues{
		OpeningFundValue: openingFundValue.String(),
		TotalDeposits:    totalDeposits.String(),
	}
	return retValue, nil
}

func updateCapitalAccountOwnership(ctx contractapi.TransactionContextInterface, account *types.CapitalAccount, openingFundValue decimal.Decimal) error {
	currentPeriod := fmt.Sprintf("%d", account.CurrentPeriod-1) //we have already updated current period for this account
	openingAccountValue, err := decimal.NewFromString(account.OpeningValue[currentPeriod])
	if err != nil {
		return err
	}
	ownership := openingAccountValue.Div(openingFundValue)
	account.OwnershipPercentage[currentPeriod] = ownership.String()
	return nil
}

func (s *AdminContract) BootstrapCapitalAccount(ctx contractapi.TransactionContextInterface, account *types.CapitalAccount) error {
	_, err := s.StepCapitalAccount(ctx, account)
	if err != nil {
		return err
	}
	return nil
}

func (s *AdminContract) StepCapitalAccount(ctx contractapi.TransactionContextInterface, account *types.CapitalAccount) (*types.CapitalAccount, error) {
	deposits, err := QueryDepositsByFundAccountPeriod(ctx, account.ID, account.CurrentPeriod)
	if err != nil {
		return account, err
	}
	withdrawals, err := QueryWithdrawalsByFundAccountPeriod(ctx, account.ID, account.CurrentPeriod)
	if err != nil {
		return account, err
	}
	total := decimal.Zero
	for _, deposit := range deposits {
		amount, err := decimal.NewFromString(deposit.Amount)
		if err != nil {
			return account, err
		}
		total = total.Add(amount)
	}
	for _, withdrawal := range withdrawals {
		amount, err := decimal.NewFromString(withdrawal.Amount)
		if err != nil {
			return account, err
		}
		total = total.Sub(amount)
	}
	currentPeriod := fmt.Sprintf("%d", account.CurrentPeriod)
	closingValue, err := decimal.NewFromString(account.ClosingValue[currentPeriod])
	if err != nil {
		return account, err
	}
	openingValue := closingValue.Add(total)
	if openingValue.Sign() == -1 {
		return account, errors.New("the actions resulted in a negative capital account balance")
	}
	//update the deposits, opening value, and current period of the account
	account.Deposits[currentPeriod] = openingValue.String()
	account.OpeningValue[currentPeriod] = openingValue.String()
	account.CurrentPeriod = account.CurrentPeriod + 1
	return account, nil
}

//BootstrappedFundValues is a struct that tracks the total deposits and closing fund value
//slightly redundant right now but it may be useful to separate these two later on?
type BootstrappedFundValues struct {
	OpeningFundValue string
	TotalDeposits    string
}
