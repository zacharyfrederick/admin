package smartcontract

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/shopspring/decimal"

	"github.com/zacharyfrederick/admin/types"
	pkgErrors "github.com/zacharyfrederick/admin/types/errors"
)

func (s *AdminContract) CreateFund(ctx contractapi.TransactionContextInterface, fundId string, name string, inceptionDate string) error {
	obj, err := ctx.GetStub().GetState(fundId)
	if err != nil {
		return pkgErrors.ReadingWorldStateError
	}
	if obj != nil {
		return pkgErrors.IdAlreadyInUseError
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

func (s *AdminContract) StepFund(ctx contractapi.TransactionContextInterface, fundId string) (*types.FundAndCapitalAccounts, error) {
	fund, err := s.QueryFundById(ctx, fundId)
	if err != nil {
		return nil, err
	}
	if fund == nil {
		return nil, pkgErrors.FundNotFoundError
	}
	if fund.CurrentPeriod == 0 {
		return nil, pkgErrors.CannotStepFundError
	}
	fundClosingValue, err := s.CalculateFundClosingValue(ctx, fund)
	if err != nil {
		return nil, err
	}
	accounts, err := s.QueryCapitalAccountsByFund(ctx, fund.ID)
	if err != nil {
		return nil, err
	}
	if accounts == nil {
		return nil, pkgErrors.NoCapitalAccountsFoundError
	}
	err = calculateCapitalAccountClosingValues(accounts, fundClosingValue)
	if err != nil {
		return nil, err
	}
	totalDeposits, err := calculateAggregateDeposits(ctx, accounts)
	if err != nil {
		return nil, err
	}
	totalFees, err := calculateAggregateFixedFees(accounts)
	if err != nil {
		return nil, err
	}
	err = transferFeesToGeneralPartner(accounts, totalFees)
	if err != nil {
		return nil, err
	}
	totalDeposits = totalDeposits.Add(totalFees) //fees from limited partners become deposits for general partner
	fundOpeningValue, err := calculateCapitalAccountOpeningValues(accounts)
	if err != nil {
		return nil, err
	}
	fundClosingValueDec := decimal.RequireFromString(fundClosingValue) //we can require it here because we already know its a valid decimal
	err = performWealthConservationFunction(fund, fundClosingValueDec, totalDeposits, totalFees, fundOpeningValue)
	if err != nil {
		return nil, err
	}
	if fund.HasPerformanceFees && fund.CurrentPeriod%fund.PerformanceFeePeriod == 0 {
		err := calculatePerformanceFees(fund, accounts) //TODO
		if err != nil {
			return nil, err
		}
	}
	fund.IncrementCurrentPeriod()
	err = SaveState(ctx, fund)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		account.IncrementCurrentPeriod()
		err := updateCapitalAccountOwnership(account, fundOpeningValue)
		if err != nil {
			return nil, err
		}
		err = SaveState(ctx, account)
		if err != nil {
			return nil, err
		}
	}
	result := &types.FundAndCapitalAccounts{
		Fund:     fund,
		Accounts: accounts,
	}
	return result, nil
}

func calculatePerformanceFees(fund *types.Fund, accounts []*types.CapitalAccount) error {
	fmt.Println("performance fees unimplemented")
	return nil
}

func performWealthConservationFunction(fund *types.Fund, closingValue decimal.Decimal, deposits decimal.Decimal, fees decimal.Decimal, openingValue decimal.Decimal) error {
	testValue := closingValue.Sub(fees).Add(deposits)
	if !testValue.Equal(openingValue) {
		return pkgErrors.WealthConservationFunctionError
	}
	fund.ClosingValues[fund.CurrentPeriodAsString()] = closingValue.String()
	fund.Deposits[fund.CurrentPeriodAsString()] = deposits.String()
	fund.FixedFees[fund.CurrentPeriodAsString()] = fees.String()
	fund.OpeningValues[fund.CurrentPeriodAsString()] = openingValue.String()
	return nil
}

func transferFeesToGeneralPartner(accounts []*types.CapitalAccount, fixedFees decimal.Decimal) error {
	for _, account := range accounts {
		if account.Number == 0 {
			existingDeposits, err := decimal.NewFromString(account.Deposits[account.CurrentPeriodAsString()])
			if err != nil {
				return pkgErrors.DecimalConversionError
			}
			newDeposits := existingDeposits.Add(fixedFees)
			account.Deposits[account.CurrentPeriodAsString()] = newDeposits.String()
			return nil
		}
	}
	return pkgErrors.GeneralPartnerNotFoundError
}

func calculateCapitalAccountOpeningValues(accounts []*types.CapitalAccount) (decimal.Decimal, error) {
	fundOpeningValue := decimal.Zero
	for _, account := range accounts {
		closingValue, err := decimal.NewFromString(account.ClosingValue[account.CurrentPeriodAsString()])
		if err != nil {
			return decimal.Zero, pkgErrors.DecimalConversionError
		}
		fixedFees, err := decimal.NewFromString(account.FixedFees[account.CurrentPeriodAsString()])
		if err != nil {
			return decimal.Zero, pkgErrors.DecimalConversionError
		}
		deposits, err := decimal.NewFromString(account.Deposits[account.CurrentPeriodAsString()])
		if err != nil {
			return decimal.Zero, err
		}
		openingValue := closingValue.Sub(fixedFees).Add(deposits)
		if openingValue.Sign() == -1 {
			return decimal.Zero, pkgErrors.NegativeCapitalAccountBalanceError
		}
		account.OpeningValue[account.CurrentPeriodAsString()] = openingValue.String()
		fundOpeningValue = fundOpeningValue.Add(openingValue)
	}
	return fundOpeningValue, nil
}

func calculateAggregateFixedFees(accounts []*types.CapitalAccount) (decimal.Decimal, error) {
	aggregateFixedFees := decimal.Zero
	for _, account := range accounts {
		accountFixedFees, err := calculateCapitalAccountFixedFees(account)
		if err != nil {
			return decimal.Zero, err
		}
		aggregateFixedFees = aggregateFixedFees.Add(accountFixedFees)
	}
	return aggregateFixedFees, nil
}

func calculateCapitalAccountFixedFees(account *types.CapitalAccount) (decimal.Decimal, error) {
	if account.Number == 0 {
		account.FixedFees[account.CurrentPeriodAsString()] = decimal.Zero.String()
		return decimal.Zero, nil
	}
	closingValue, err := decimal.NewFromString(account.ClosingValue[account.CurrentPeriodAsString()])
	if err != nil {
		return decimal.Zero, pkgErrors.DecimalConversionError
	}
	fixedFeePercentage, err := decimal.NewFromString(account.FixedFee)
	if err != nil {
		return decimal.Zero, pkgErrors.DecimalConversionError
	}
	fixedFee := fixedFeePercentage.Mul(closingValue)
	account.FixedFees[account.CurrentPeriodAsString()] = fixedFee.String()
	return fixedFee, nil
}

func calculateAggregateDeposits(ctx contractapi.TransactionContextInterface, accounts []*types.CapitalAccount) (decimal.Decimal, error) {
	aggregateDeposits := decimal.Zero
	for _, account := range accounts {
		accountDeposits, err := calculateCapitalAccountDeposits(ctx, account)
		if err != nil {
			return decimal.Zero, err
		}
		aggregateDeposits = aggregateDeposits.Add(accountDeposits)
	}
	return aggregateDeposits, nil
}

func calculateCapitalAccountDeposits(ctx contractapi.TransactionContextInterface, account *types.CapitalAccount) (decimal.Decimal, error) {
	deposits, err := QueryDepositsByFundAccountPeriod(ctx, account.ID, account.CurrentPeriod)
	if err != nil {
		return decimal.Zero, err
	}
	withdrawals, err := QueryWithdrawalsByFundAccountPeriod(ctx, account.ID, account.CurrentPeriod)
	if err != nil {
		return decimal.Zero, err
	}
	total, err := aggregateActions(deposits, withdrawals)
	if err != nil {
		return decimal.Zero, err
	}
	account.Deposits[account.CurrentPeriodAsString()] = total.String()
	return total, nil
}

func calculateCapitalAccountClosingValues(accounts []*types.CapitalAccount, fundClosingValue string) error {
	closingValue, err := decimal.NewFromString(fundClosingValue)
	if err != nil {
		return pkgErrors.DecimalConversionError
	}
	for _, account := range accounts {
		err := updateCapitalAccountClosingValue(account, closingValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateCapitalAccountClosingValue(account *types.CapitalAccount, fundClosingValue decimal.Decimal) error {
	previousPeriod := account.PreviousPeriodAsString()
	previousOwnershipPercentage, ok := account.OwnershipPercentage[previousPeriod]
	if !ok {
		return pkgErrors.PreviousOwnershipPercentageNotFoundError
	}
	ownershipPercentage, err := decimal.NewFromString(previousOwnershipPercentage)
	if err != nil {
		return pkgErrors.DecimalConversionError
	}
	closingValue := ownershipPercentage.Mul(fundClosingValue)
	account.ClosingValue[account.CurrentPeriodAsString()] = closingValue.String()
	return nil
}

func (s *AdminContract) CalculateFundClosingValue(ctx contractapi.TransactionContextInterface, fund *types.Fund) (string, error) {
	portfolios, err := s.QueryPortfoliosByFund(ctx, fund.ID)
	if err != nil {
		return "", err
	}
	if portfolios == nil {
		return "", pkgErrors.NoPortfoliosFoundError
	}
	NAV := decimal.Zero
	for _, portfolio := range portfolios {
		if portfolio.MostRecentDate == "" {
			return "", pkgErrors.NoMostRecentDateForPortfolioError
		}
		valuationDate := portfolio.MostRecentDate
		valuations, ok := portfolio.Valuations[valuationDate]
		if !ok {
			return "", pkgErrors.NoValuationsFoundForDateError
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
			return decimal.Zero, pkgErrors.DecimalConversionError
		}
		price, err := decimal.NewFromString(valuedAsset.Price)
		if err != nil {
			return decimal.Zero, pkgErrors.DecimalConversionError
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
		return nil, pkgErrors.FundNotFoundError
	}
	if fund.CurrentPeriod != 0 {
		return nil, pkgErrors.CannotBootstrapFundError
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

func (s *AdminContract) BootstrapCapitalAccountsForFund(ctx contractapi.TransactionContextInterface, fundId string) (*bootstrappedFundValues, error) {
	accounts, err := queryCapitalAccountsByFund(ctx, fundId)
	if err != nil {
		return &bootstrappedFundValues{}, err
	}
	//loop over accounts, aggregate deposits and track closing fund value
	openingFundValue := decimal.Zero
	totalDeposits := decimal.Zero
	for _, account := range accounts {
		err := s.BootstrapCapitalAccount(ctx, account)
		if err != nil {
			return &bootstrappedFundValues{}, errors.New("err from bootstrap capital account")
		}
		currentPeriod := fmt.Sprintf("%d", account.CurrentPeriod-1) //we updated the current period already
		deposit, err := decimal.NewFromString(account.Deposits[currentPeriod])
		if err != nil {
			return &bootstrappedFundValues{}, err
		}
		totalDeposits = totalDeposits.Add(deposit)
		openingFundValue = openingFundValue.Add(deposit)
	}
	//update the ownership percentage for each account based on the closing fund value
	for _, account := range accounts {
		err := updateCapitalAccountOwnership(account, openingFundValue)
		if err != nil {
			return &bootstrappedFundValues{}, err
		}
		SaveState(ctx, account)
	}
	retValue := &bootstrappedFundValues{
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
		return pkgErrors.CannotBootstrapCapitalAccountError
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
		return pkgErrors.NegativeCapitalAccountBalanceError
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

type bootstrappedFundValues struct {
	OpeningFundValue string
	TotalDeposits    string
}
