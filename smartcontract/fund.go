package smartcontract

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/shopspring/decimal"

	"github.com/zacharyfrederick/admin/types"
	smartcontracterrors "github.com/zacharyfrederick/admin/types/errors"
)

func (s *AdminContract) CreateFund(ctx contractapi.TransactionContextInterface, fundId string, name string, inceptionDate string) error {
	obj, err := ctx.GetStub().GetState(fundId)
	if err != nil {
		return smartcontracterrors.ReadingWorldStateError
	}
	if obj != nil {
		return smartcontracterrors.IdAlreadyInUseError
	}
	fund := types.CreateDefaultFund(fundId, name, inceptionDate)
	return SaveState(ctx, &fund)
}

func (s *AdminContract) QueryFundById(ctx contractapi.TransactionContextInterface, fundId string) (*types.Fund, error) {
	fundJSON, err := ctx.GetStub().GetState(fundId)
	if err != nil {
		return nil, err
	}
	if fundJSON == nil {
		return nil, nil
	}
	var fund types.Fund
	err = LoadState(fundJSON, &fund)
	if err != nil {
		return nil, err
	}
	return &fund, err
}

func (s *AdminContract) StepFund(ctx contractapi.TransactionContextInterface, fundId string) (*types.Fund, error) {
	fund, err := s.QueryFundById(ctx, fundId)
	if err != nil {
		return nil, err
	}
	if fund == nil {
		return nil, smartcontracterrors.FundNotFoundError
	}
	if fund.CurrentPeriod == 0 {
		return nil, smartcontracterrors.CannotStepFundError
	}
	fundClosingVaue, err := s.CalculateFundClosingValue(ctx, fund)
	if err != nil {
		return nil, err
	}
	fmt.Println(fundClosingVaue)
	return fund, nil
}

func (s *AdminContract) CalculateFundClosingValue(ctx contractapi.TransactionContextInterface, fund *types.Fund) (string, error) {
	portfolios, err := s.QueryPortfoliosByFund(ctx, fund.ID)
	if err != nil {
		return "", err
	}
	if portfolios == nil {
		return "", smartcontracterrors.NoPortfoliosFoundError
	}
	NAV := decimal.Zero
	for _, portfolio := range portfolios {
		fmt.Println(portfolio)
		if portfolio.MostRecentDate == "" {
			return "", smartcontracterrors.NoMostRecentDateForPortfolioError
		}
		valuationDate := portfolio.MostRecentDate
		valuations, ok := portfolio.Valuations[valuationDate]
		if !ok {
			return "", smartcontracterrors.NoValuationsFoundForDateError
		}
		portfolioTotal, err := calculatePortfolioNAV(valuations)
		if err != nil {
			return "", err
		}
		NAV = NAV.Add(portfolioTotal)
	}
	return NAV.String(), nil
}

func calculatePortfolioNAV(valuations types.ValuedAssetMap) (decimal.Decimal, error) {
	portfolioTotal := decimal.Zero
	for _, valuedAsset := range valuations {
		amount, err := decimal.NewFromString(valuedAsset.Amount)
		if err != nil {
			return decimal.Zero, smartcontracterrors.DecimalConversionError
		}
		price, err := decimal.NewFromString(valuedAsset.Price)
		if err != nil {
			return decimal.Zero, smartcontracterrors.DecimalConversionError
		}
		subtotal := amount.Mul(price)
		portfolioTotal = portfolioTotal.Add(subtotal)
	}
	return portfolioTotal, nil
}

func (s *AdminContract) BootstrapFund(ctx contractapi.TransactionContextInterface, fundId string) (*types.Fund, error) {
	fund, err := s.QueryFundById(ctx, fundId)
	if err != nil {
		return nil, err
	}
	if fund == nil {
		return nil, smartcontracterrors.FundNotFoundError
	}
	if fund.CurrentPeriod != 0 {
		return nil, smartcontracterrors.CannotBootstrapFundError
	}
	bootstrappedFundValues, err := s.BootstrapCapitalAccountsForFund(ctx, fundId)
	if err != nil {
		return nil, err
	}
	fund.BootstrapFundValues(bootstrappedFundValues.TotalDeposits, bootstrappedFundValues.OpeningFundValue)
	err = SaveState(ctx, fund)
	if err != nil {
		return nil, err
	}
	return fund, nil
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
		err := updateCapitalAccountOwnership(account, openingFundValue)
		if err != nil {
			return &BootstrappedFundValues{}, err
		}
		SaveState(ctx, account)
	}
	retValue := &BootstrappedFundValues{
		OpeningFundValue: openingFundValue.String(),
		TotalDeposits:    totalDeposits.String(),
	}
	return retValue, nil
}

func updateCapitalAccountOwnership(account *types.CapitalAccount, openingFundValue decimal.Decimal) error {
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
		return smartcontracterrors.CannotBootstrapCapitalAccountError
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
		return smartcontracterrors.NegativeCapitalAccountBalanceError
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
