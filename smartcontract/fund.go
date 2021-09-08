package smartcontract

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/shopspring/decimal"

	"github.com/zacharyfrederick/admin/types"
	"github.com/zacharyfrederick/admin/types/doctypes"
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
	closingValues["0"] = "0"
	openingValues := make(map[string]string)
	openingValues["0"] = "0"
	fixedFees := make(map[string]string)
	fixedFees["0"] = "0"
	deposits := make(map[string]string)
	deposits["0"] = "0"
	fund := types.Fund{
		DocType:            doctypes.DOCTYPE_FUND,
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
	return fund.SaveState(ctx)
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
	bootstrappedFundValues, err := s.BootstrapCapitalAccountsForFund(ctx, fundId)
	if err != nil {
		return err
	}
	fund.BootstrapFundValues(bootstrappedFundValues.TotalDeposits, bootstrappedFundValues.OpeningFundValue)
	return fund.SaveState(ctx)
}

func (s *AdminContract) BootstrapCapitalAccountsForFund(ctx contractapi.TransactionContextInterface, fundId string) (*BootstrappedFundValues, error) {
	accounts, err := queryCapitalAccountsByFund(ctx, fundId)
	if err != nil {
		return &BootstrappedFundValues{}, err
	}
	//loop over accounts, aggregate deposits and track closing fund value
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
		account.SaveState(ctx)
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
	if account.CurrentPeriod != 0 {
		return errors.New("this capital account cannot be bootstrapped")
	}
	deposits, err := QueryDepositsByFundAccountPeriod(ctx, account.ID, account.CurrentPeriod)
	if err != nil {
		return err
	}
	withdrawals, err := QueryWithdrawalsByFundAccountPeriod(ctx, account.ID, account.CurrentPeriod)
	if err != nil {
		return err
	}
	total, err := aggregateActions(deposits, withdrawals)
	if err != nil {
		return err
	}
	currentPeriod := account.CurrentPeriodAsString()
	closingValue, err := decimal.NewFromString(account.ClosingValue[currentPeriod])
	if err != nil {
		return err
	}
	openingValue := closingValue.Add(total)
	if openingValue.Sign() == -1 {
		return errors.New("the actions resulted in a negative capital account balance")
	}
	account.BootstrapAccountValues(openingValue.String())
	return nil
}

func aggregateActions(deposits []*types.CapitalAccountAction, withdrawals []*types.CapitalAccountAction) (decimal.Decimal, error) {
	totalDeposis, err := aggregateDeposits(deposits)
	if err != nil {
		return decimal.Zero, err
	}
	totalWithdrawals, err := aggregateWithdrawals(withdrawals)
	if err != nil {
		return decimal.Zero, err
	}
	total := totalDeposis.Add(totalWithdrawals)
	return total, nil
}

func aggregateDeposits(deposits []*types.CapitalAccountAction) (decimal.Decimal, error) {
	total := decimal.Zero
	for _, deposit := range deposits {
		amount, err := decimal.NewFromString(deposit.Amount)
		if err != nil {
			return decimal.Zero, err
		}
		total = total.Add(amount)
	}
	return total, nil
}

func aggregateWithdrawals(withdrawals []*types.CapitalAccountAction) (decimal.Decimal, error) {
	total := decimal.Zero
	for _, deposit := range withdrawals {
		amount, err := decimal.NewFromString(deposit.Amount)
		if err != nil {
			return decimal.Zero, err
		}
		total = total.Sub(amount)
	}
	return total, nil
}

//BootstrappedFundValues is a struct that tracks the total deposits and closing fund value
//slightly redundant right now but it may be useful to separate these two later on?
type BootstrappedFundValues struct {
	OpeningFundValue string
	TotalDeposits    string
}
