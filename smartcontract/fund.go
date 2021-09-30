package smartcontract

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/zacharyfrederick/admin/types"
	pkgErrors "github.com/zacharyfrederick/admin/types/errors"
)

func (s *AdminContract) CreateFund(
	ctx SmartContractContext,
	fundId string,
	name string,
	inceptionDate string,
) error {
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

func (s *AdminContract) QueryFundById(
	ctx SmartContractContext,
	fundId string,
) (*types.Fund, error) {
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

func (s *AdminContract) StepFund(
	ctx SmartContractContext,
	fundId string,
) (*types.FundAndCapitalAccounts, error) {
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
	fundClosingValue, err := calculateFundClosingValue(ctx, fund)
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
	totalDeposits, err := calculateAggregateDeposits(ctx, fund, accounts)
	if err != nil {
		return nil, err
	}
	totalFixedFees, err := calculateAggregateFixedFees(accounts)
	if err != nil {
		return nil, err
	}
	err = transferFixedFeesToGeneralPartner(accounts, totalFixedFees)
	if err != nil {
		return nil, err
	}
	totalDeposits = totalDeposits.Add(
		totalFixedFees,
	) //fees from limited partners become deposits for general partner
	fundOpeningValue, err := calculateCapitalAccountOpeningValues(accounts)
	if err != nil {
		return nil, err
	}
	err = performWealthConservationFunction(
		fund,
		fundClosingValue,
		totalDeposits,
		totalFixedFees,
		fundOpeningValue,
	)
	if err != nil {
		return nil, err
	}
	fund.IncrementCurrentPeriod()
	fund.MidYearDeposits = []string{}
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
	return &types.FundAndCapitalAccounts{Fund: fund, Accounts: accounts}, nil
}

func performWealthConservationFunction(
	fund *types.Fund,
	closingValue decimal.Decimal,
	deposits decimal.Decimal,
	fees decimal.Decimal,
	openingValue decimal.Decimal,
) error {
	testValue := closingValue.Sub(fees).Add(deposits)
	if !testValue.Equal(openingValue) {
		return pkgErrors.WealthConservationFunctionError
	}
	fund.ClosingValues[fund.CurrentPeriod] = closingValue.String()
	fund.Deposits[fund.CurrentPeriod] = deposits.String()
	fund.FixedFees[fund.CurrentPeriod] = fees.String()
	fund.OpeningValues[fund.CurrentPeriod] = openingValue.String()
	return nil
}

func transferFixedFeesToGeneralPartner(
	accounts []*types.CapitalAccount,
	fixedFees decimal.Decimal,
) error {
	for _, account := range accounts {
		if account.Number == 0 {
			existingDeposits, err := decimal.NewFromString(
				account.Deposits[account.CurrentPeriod],
			)
			if err != nil {
				return pkgErrors.DecimalConversionError
			}
			newDeposits := existingDeposits.Add(fixedFees)
			account.Deposits[account.CurrentPeriod] = newDeposits.String()
			return nil
		}
	}
	return pkgErrors.GeneralPartnerNotFoundError
}

func calculateCapitalAccountOpeningValues(
	accounts []*types.CapitalAccount,
) (decimal.Decimal, error) {
	fundOpeningValue := decimal.Zero
	for _, account := range accounts {
		closingValue, err := decimal.NewFromString(
			account.ClosingValue[account.CurrentPeriod],
		)
		if err != nil {
			return decimal.Zero, pkgErrors.DecimalConversionError
		}
		fixedFees, err := decimal.NewFromString(account.FixedFees[account.CurrentPeriod])
		if err != nil {
			return decimal.Zero, pkgErrors.DecimalConversionError
		}
		deposits, err := decimal.NewFromString(account.Deposits[account.CurrentPeriod])
		if err != nil {
			return decimal.Zero, err
		}
		openingValue := closingValue.Sub(fixedFees).Add(deposits)
		if openingValue.Sign() == -1 {
			return decimal.Zero, pkgErrors.NegativeCapitalAccountBalanceError
		}
		account.OpeningValue[account.CurrentPeriod] = openingValue.String()
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
		account.FixedFees[account.CurrentPeriod] = decimal.Zero.String()
		return decimal.Zero, nil
	}
	closingValue, err := decimal.NewFromString(
		account.ClosingValue[account.CurrentPeriod],
	)
	if err != nil {
		return decimal.Zero, pkgErrors.DecimalConversionError
	}
	fixedFeePercentage, err := decimal.NewFromString(account.FixedFee)
	if err != nil {
		return decimal.Zero, pkgErrors.DecimalConversionError
	}
	fixedFee := fixedFeePercentage.Mul(closingValue)
	account.FixedFees[account.CurrentPeriod] = fixedFee.String()
	return fixedFee, nil
}

func calculateAggregateDeposits(
	ctx SmartContractContext,
	fund *types.Fund,
	accounts []*types.CapitalAccount,
) (decimal.Decimal, error) {
	aggregateDeposits := decimal.Zero
	for _, account := range accounts {
		//if the account has performance fees and isn't on the list of mid year deposits we don't want to aggregate deposits yet
		//for the accounts that match this conditional we apply deposits after the application of performance fees
		if account.HasPerformanceFees && !contains(fund.MidYearDeposits, account.ID) {
			account.Deposits[account.CurrentPeriod] = decimal.Zero.String()
			continue
		}
		accountDeposits, err := calculateCapitalAccountDeposits(ctx, account)
		if err != nil {
			return decimal.Zero, err
		}
		aggregateDeposits = aggregateDeposits.Add(accountDeposits)
	}
	return aggregateDeposits, nil
}

func calculateCapitalAccountDeposits(
	ctx SmartContractContext,
	account *types.CapitalAccount,
) (decimal.Decimal, error) {
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
	account.Deposits[account.CurrentPeriod] = total.String()
	return total, nil
}

func contains(accountIds []string, testId string) bool {
	for _, id := range accountIds {
		if id == testId {
			return true
		}
	}
	return false
}

func calculateMidYearDeposits(
	ctx SmartContractContext,
	fund *types.Fund,
	accounts []*types.CapitalAccount,
) (decimal.Decimal, error) {
	aggregateDeposits := decimal.Zero
	for _, account := range accounts {
		if !account.HasPerformanceFees {
			continue
		}
		if !contains(fund.MidYearDeposits, account.ID) {
			continue
		}
		accountDeposits, err := calculateCapitalAccountDeposits(ctx, account)
		if err != nil {
			return decimal.Zero, err
		}
		aggregateDeposits = aggregateDeposits.Add(accountDeposits)
	}
	return aggregateDeposits, nil
}

func calculateCapitalAccountClosingValues(
	accounts []*types.CapitalAccount,
	fundClosingValue decimal.Decimal,
) error {
	for _, account := range accounts {
		account.UpdateClosingValue(fundClosingValue)
	}
	return nil
}

func updateCapitalAccountClosingValue(
	account *types.CapitalAccount,
	fundClosingValue decimal.Decimal,
) error {
	previousOwnershipPercentage, ok := account.OwnershipPercentage[account.PreviousPeriod()]
	if !ok {
		return pkgErrors.PreviousOwnershipPercentageNotFoundError
	}
	ownershipPercentage, err := decimal.NewFromString(previousOwnershipPercentage)
	if err != nil {
		return pkgErrors.DecimalConversionError
	}
	closingValue := ownershipPercentage.Mul(fundClosingValue)
	account.ClosingValue[account.CurrentPeriod] = closingValue.String()
	return nil
}

func (s *AdminContract) CalculateFundClosingValue(
	ctx SmartContractContext,
	fund *types.Fund,
) (string, error) {
	fundClosingValue, err := calculateFundClosingValue(ctx, fund)
	if err != nil {
		return decimal.Zero.String(), err
	}
	return fundClosingValue.String(), nil
}

func calculateFundClosingValue(
	ctx SmartContractContext,
	fund *types.Fund,
) (decimal.Decimal, error) {
	portfolios, err := queryPortfoliosByFund(ctx, fund.ID)
	if err != nil {
		return decimal.Zero, err
	}
	if portfolios == nil {
		return decimal.Zero, pkgErrors.NoPortfoliosFoundError
	}
	NAV := decimal.Zero
	for _, portfolio := range portfolios {
		if portfolio.MostRecentDate == "" {
			return decimal.Zero, pkgErrors.NoMostRecentDateForPortfolioError
		}
		valuationDate := portfolio.MostRecentDate
		valuations, ok := portfolio.Valuations[valuationDate]
		if !ok {
			return decimal.Zero, pkgErrors.NoValuationsFoundForDateError
		}
		portfolioTotal, err := calculatePortfolioNAV(valuations)
		if err != nil {
			return decimal.Zero, err
		}
		NAV = NAV.Add(portfolioTotal)
	}
	return NAV, nil
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

func (s *AdminContract) BootstrapFund(
	ctx SmartContractContext,
	fundId string,
) (*types.Fund, error) {
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
	fund.BootstrapFundValues(
		bootstrappedFundValues.TotalDeposits,
		bootstrappedFundValues.OpeningFundValue,
	)
	err = SaveState(ctx, fund)
	if err != nil {
		return nil, err
	}
	return fund, nil
}

func (s *AdminContract) BootstrapCapitalAccountsForFund(
	ctx SmartContractContext,
	fundId string,
) (*bootstrappedFundValues, error) {
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
		currentPeriod := account.CurrentPeriod - 1
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
		setHighWaterMark(account)
		SaveState(ctx, account)
	}
	retValue := &bootstrappedFundValues{
		OpeningFundValue: openingFundValue.String(),
		TotalDeposits:    totalDeposits.String(),
	}
	return retValue, nil
}

func setHighWaterMark(account *types.CapitalAccount) {
	period := account.PreviousPeriod()
	closingValue := account.ClosingValue[period]
	highWaterMark := types.HighWaterMark{
		Amount: closingValue,
		Date:   period}
	account.HighWaterMark = highWaterMark
}

func updateCapitalAccountOwnership(
	account *types.CapitalAccount,
	openingFundValue decimal.Decimal,
) error {
	currentPeriod := account.PreviousPeriod()
	openingAccountValue, err := decimal.NewFromString(account.OpeningValue[currentPeriod])
	if err != nil {
		return err
	}
	ownership := openingAccountValue.Div(openingFundValue)
	account.OwnershipPercentage[currentPeriod] = ownership.String()
	return nil
}

func (s *AdminContract) BootstrapCapitalAccount(
	ctx SmartContractContext,
	account *types.CapitalAccount,
) error {
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
	currentPeriod := account.CurrentPeriod
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

func aggregateActions(
	deposits []*types.CapitalAccountAction,
	withdrawals []*types.CapitalAccountAction,
) (decimal.Decimal, error) {
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

func (s *AdminContract) StepFundPerfFees(
	ctx SmartContractContext,
	fundId string,
) (*types.FundAndCapitalAccounts, error) {
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
	fundClosingValue, err := calculateFundClosingValue(ctx, fund)
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
	stepResult := createStepFundResult()
	if fund.IsPerformanceFeePeriod() {
		accountsNoPerfFees, accountsPerfFees := splitSubsetsPerfPeriod(accounts)
		subset1Result, err := processSubset1(ctx, fund, fundClosingValue, accountsNoPerfFees)
		if err != nil {
			return nil, err
		}
		subset2Result, err := processSubset2(ctx, fund, fundClosingValue, accountsPerfFees)
		if err != nil {
			return nil, err
		}
		stepResult = aggregateSubsetResults(subset1Result, subset2Result)
	} else {
		accountsNoPerfFees, accountsPerfFees, accountsMidYearDeposits, accountsMidYearWithdrawals := splitSubsetsRegular(fund, accounts)
		subset1Result, err := processSubset1(ctx, fund, fundClosingValue, accountsNoPerfFees)
		if err != nil {
			return nil, err
		}
		subset2Result, err := processSubset2(ctx, fund, fundClosingValue, accountsPerfFees)
		if err != nil {
			return nil, err
		}
		subset3Result, err := processSubset3(ctx, fund, fundClosingValue, accountsMidYearDeposits)
		if err != nil {
			return nil, err
		}
		subset4Result, err := processSubset4(ctx, fund, fundClosingValue, accountsMidYearWithdrawals)
		if err != nil {
			return nil, err
		}
		stepResult = aggregateSubsetResults(subset1Result, subset2Result, subset3Result, subset4Result)
	}
	fmt.Println(stepResult)
	return nil, err
}

func aggregateSubsetResults(results ...*StepFundResult) *StepFundResult {
	stepResult := createStepFundResult()
	for _, subset := range results {
		stepResult.ClosingValue = stepResult.ClosingValue.Add(subset.ClosingValue)
		stepResult.OpeningValue = stepResult.OpeningValue.Add(subset.OpeningValue)
		stepResult.Deposits = stepResult.Deposits.Add(subset.Deposits)
		stepResult.FixedFees = stepResult.FixedFees.Add(subset.FixedFees)
		stepResult.PerfFees = stepResult.PerfFees.Add(subset.PerfFees)
		stepResult.Accounts = append(stepResult.Accounts, subset.Accounts...)
	}
	return stepResult
}

func processSubset1(
	ctx SmartContractContext,
	fund *types.Fund,
	closingValue decimal.Decimal,
	accounts []*types.CapitalAccount,
) (*StepFundResult, error) {
	deposits := decimal.Zero
	fixedFees := decimal.Zero
	openingValue := decimal.Zero
	var gpIndex int = -1
	for i, account := range accounts {
		account.UpdateClosingValue(closingValue)
		accountDeposits := decimal.Zero
		depositList, err := QueryDepositsByFundAccountPeriod(ctx, account.ID, account.CurrentPeriod)
		if err != nil {
			return nil, err
		}
		for _, deposit := range depositList {
			amount := decimal.RequireFromString(deposit.Amount)
			accountDeposits = accountDeposits.Add(amount)
		}
		withdrawalList, err := QueryWithdrawalsByFundAccountPeriod(ctx, account.ID, account.CurrentPeriod)
		if err != nil {
			return nil, err
		}
		for _, deposit := range withdrawalList {
			amount := decimal.RequireFromString(deposit.Amount)
			accountDeposits = accountDeposits.Sub(amount)
		}
		accountClosingValue := decimal.RequireFromString(account.ClosingValue[account.CurrentPeriod])
		resultingAccountBalance := accountClosingValue.Add(accountDeposits)
		if resultingAccountBalance.Sign() == -1 {
			return nil, pkgErrors.NegativeCapitalAccountBalanceError
		}
		account.Deposits[account.CurrentPeriod] = accountDeposits.String()
		deposits = deposits.Add(accountDeposits)

		//calculate fixed fees
		accountFixedFees := decimal.Zero
		if account.Number != 0 {
			fixedFeeRate := decimal.RequireFromString(account.FixedFee)
			accountFixedFees = accountClosingValue.Mul(fixedFeeRate)
		} else {
			gpIndex = i
		}
		account.FixedFees[account.CurrentPeriod] = accountFixedFees.String()
		fixedFees = fixedFees.Add(accountFixedFees)

		//calculate opening value
		accountOpeningValue := accountClosingValue.Sub(accountFixedFees).Add(accountDeposits)
		if accountOpeningValue.Sign() == -1 {
			return nil, pkgErrors.NegativeCapitalAccountBalanceError
		}
		openingValue = openingValue.Add(openingValue)
	}

	//transfer fees to general partner
	deposits = deposits.Add(fixedFees)
	if gpIndex == -1 {
		return nil, pkgErrors.GeneralPartnerNotFoundError
	}
	generalPartner := accounts[gpIndex]
	existingDeposits := decimal.RequireFromString(generalPartner.Deposits[generalPartner.CurrentPeriod])
	newDeposits := existingDeposits.Add(fixedFees)
	generalPartner.Deposits[generalPartner.CurrentPeriod] = newDeposits.String()
	prevOpeningValue := decimal.RequireFromString(generalPartner.OpeningValue[generalPartner.CurrentPeriod])
	newOpeningValue := prevOpeningValue.Add(fixedFees)
	generalPartner.OpeningValue[generalPartner.CurrentPeriod] = newOpeningValue.String()
	openingValue = openingValue.Add(fixedFees)

	stepResult := createStepFundResult()
	stepResult.ClosingValue = closingValue
	stepResult.Deposits = deposits
	stepResult.FixedFees = fixedFees
	stepResult.OpeningValue = openingValue
	stepResult.PerfFees = decimal.Zero
	stepResult.Accounts = accounts
	return stepResult, nil
}

func processSubset2(
	ctx SmartContractContext,
	fund *types.Fund,
	closingValue decimal.Decimal,
	accounts []*types.CapitalAccount,
) (*StepFundResult, error) {
	deposits := decimal.Zero
	fixedFees := decimal.Zero
	perfFees := decimal.Zero
	openingValue := decimal.Zero

	stepResult := createStepFundResult()
	stepResult.ClosingValue = closingValue
	stepResult.Deposits = deposits
	stepResult.FixedFees = fixedFees
	stepResult.OpeningValue = openingValue
	stepResult.PerfFees = perfFees
	stepResult.Accounts = accounts
	return stepResult, nil
}

func processSubset3(
	ctx SmartContractContext,
	fund *types.Fund,
	fundClosingValue decimal.Decimal,
	accounts []*types.CapitalAccount,
) (*StepFundResult, error) {
	stepResult := createStepFundResult()
	return stepResult, nil
}

func processSubset4(
	ctx SmartContractContext,
	fund *types.Fund,
	fundClosingValue decimal.Decimal,
	accounts []*types.CapitalAccount,
) (*StepFundResult, error) {
	stepResult := createStepFundResult()
	return stepResult, nil
}

func splitSubsetsPerfPeriod(
	accounts []*types.CapitalAccount,
) ([]*types.CapitalAccount, []*types.CapitalAccount) {
	accountsNoPerfFees := []*types.CapitalAccount{}
	accountsPerfFees := []*types.CapitalAccount{}
	for _, account := range accounts {
		if account.HasPerformanceFees {
			accountsPerfFees = append(accountsPerfFees, account)
		} else {
			accountsNoPerfFees = append(accountsNoPerfFees, account)
		}
	}
	return accountsNoPerfFees, accountsPerfFees
}

func splitSubsetsRegular(
	fund *types.Fund,
	accounts []*types.CapitalAccount,
) ([]*types.CapitalAccount, []*types.CapitalAccount, []*types.CapitalAccount, []*types.CapitalAccount) {
	accountsNoPerfFees := []*types.CapitalAccount{}
	accountsPerfFees := []*types.CapitalAccount{}
	accountsMidYearDeposits := []*types.CapitalAccount{}
	accountsMidYearWithdrawals := []*types.CapitalAccount{}
	for _, account := range accounts {
		if account.HasPerformanceFees {
			if contains(fund.MidYearDeposits, account.ID) {
				accountsMidYearDeposits = append(accountsMidYearDeposits, account)
				continue
			}
			if contains(fund.MidYearWithdrawals, account.ID) {
				accountsMidYearWithdrawals = append(accountsMidYearWithdrawals, account)
				continue
			}
			accountsPerfFees = append(accountsPerfFees, account)
		} else {
			accountsNoPerfFees = append(accountsNoPerfFees, account)
		}
	}
	return accountsNoPerfFees, accountsPerfFees, accountsMidYearDeposits, accountsMidYearWithdrawals
}

type StepFundResult struct {
	ClosingValue decimal.Decimal
	OpeningValue decimal.Decimal
	Deposits     decimal.Decimal
	FixedFees    decimal.Decimal
	PerfFees     decimal.Decimal
	Accounts     []*types.CapitalAccount
}

func createStepFundResult() *StepFundResult {
	stepResult := &StepFundResult{
		ClosingValue: decimal.Zero,
		OpeningValue: decimal.Zero,
		Deposits:     decimal.Zero,
		FixedFees:    decimal.Zero,
		PerfFees:     decimal.Zero,
		Accounts:     []*types.CapitalAccount{},
	}
	return stepResult
}
