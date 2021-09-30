package smartcontract_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/zacharyfrederick/admin/smartcontract"
	"github.com/zacharyfrederick/admin/types"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	"github.com/stretchr/testify/assert"
	"github.com/zacharyfrederick/admin/smartcontract/mocks"
	smartcontracterrors "github.com/zacharyfrederick/admin/types/errors"
)

//go:generate counterfeiter -o mocks/transaction.go -fake-name TransactionContext . transactionContext
type transactionContext interface {
	contractapi.TransactionContextInterface
}

//go:generate counterfeiter -o mocks/chaincodestub.go -fake-name ChaincodeStub . chaincodeStub
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate counterfeiter -o mocks/statequeryiterator.go -fake-name StateQueryIterator . stateQueryIterator
type stateQueryIterator interface {
	shim.StateQueryIteratorInterface
}

func prepareTest() (*mocks.ChaincodeStub, *mocks.TransactionContext) {
	chaincodeStub := &mocks.ChaincodeStub{}
	transactionContext := &mocks.TransactionContext{}
	transactionContext.GetStubReturns(chaincodeStub)
	return chaincodeStub, transactionContext
}
func TestCreateFund(t *testing.T) {
	_, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	err := admin.CreateFund(transactionContext, "test_id", "test_fund", "12-27-1996")
	assert.Nil(t, err)
}

func TestCreateFundExistingId(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	chaincodeStub.GetStateReturns([]byte("fake_object"), nil)
	admin := smartcontract.AdminContract{}
	err := admin.CreateFund(transactionContext, "test_id", "test_fund", "12-27-1996")
	assert.Equal(t, err, smartcontracterrors.IdAlreadyInUseError)
}

func TestCreateFundErrorReadingWorldState(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	chaincodeStub.GetStateReturns(nil, smartcontracterrors.ReadingWorldStateError)
	admin := smartcontract.AdminContract{}
	err := admin.CreateFund(transactionContext, "test_id", "test_fund", "12-27-1996")
	assert.Equal(t, err, smartcontracterrors.ReadingWorldStateError)
}

func TestCreateInvestor(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	chaincodeStub.GetStateReturns(nil, nil)
	admin := smartcontract.AdminContract{}
	err := admin.CreateInvestor(transactionContext, "test_id", "test_name")
	assert.Nil(t, err)
}

func TestCreateInvestorErrorReadingWorldState(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	chaincodeStub.GetStateReturns(nil, errors.New("fake error reading world state"))
	admin := smartcontract.AdminContract{}
	err := admin.CreateInvestor(transactionContext, "test_id", "test_name")
	assert.Equal(t, err, smartcontracterrors.ReadingWorldStateError)
}

func TestCreateInvestorExistingId(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	chaincodeStub.GetStateReturns([]byte("fake object"), nil)
	admin := smartcontract.AdminContract{}
	err := admin.CreateInvestor(transactionContext, "test_id", "test_name")
	assert.Equal(t, err, smartcontracterrors.IdAlreadyInUseError)
}

func TestCreateCapitalAccount(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	// create the test fund
	fund := types.CreateDefaultFund("testFundId", "testFund", "testDate")
	fundJSON, err := fund.ToJSON()
	assert.Nil(t, err)
	//create the test investor
	investor := types.CreateDefaultInvestor("testInvestorId", "testInvestor")
	investorJSON, err := investor.ToJSON()
	assert.Nil(t, err)
	//ensure that the stub returns the right items in order
	chaincodeStub.GetStateReturnsOnCall(0, nil, nil)          //test if the id is in use
	chaincodeStub.GetStateReturnsOnCall(1, fundJSON, nil)     //QueryFundById
	chaincodeStub.GetStateReturnsOnCall(2, investorJSON, nil) //QueryInvestorById
	chaincodeStub.PutStateReturnsOnCall(0, nil)               //fund.SaveState(ctx)
	chaincodeStub.PutStateReturnsOnCall(1, nil)               //capitalAccount.SaveState(ctx)
	//run the test
	err = admin.CreateCapitalAccount(
		transactionContext,
		"testAccountId",
		"testFundId",
		"testInvestorId",
		false,
		"0",
	)
	assert.Nil(t, err)
}

func TestCreateCapitalAccountExistingId(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturnsOnCall(0, []byte("fake data"), nil)
	err := admin.CreateCapitalAccount(
		transactionContext,
		"testAccountId",
		"testFundId",
		"testInvestorId",
		false,
		"0",
	)
	assert.Equal(t, err, smartcontracterrors.IdAlreadyInUseError)
}

func TestCreateCapitalAccountMissingFund(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturnsOnCall(0, nil, nil)
	chaincodeStub.GetStateReturnsOnCall(1, nil, nil)
	err := admin.CreateCapitalAccount(
		transactionContext,
		"testAccountId",
		"testFundId",
		"testInvestorId",
		false,
		"0",
	)
	assert.Equal(t, err, smartcontracterrors.FundNotFoundError)
}

func TestCreateCapitalAccountMissingInvestor(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	fund := types.CreateDefaultFund("testFundId", "testFund", "testDate")
	fundJSON, err := fund.ToJSON()
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, nil, nil)
	chaincodeStub.GetStateReturnsOnCall(1, fundJSON, nil)
	chaincodeStub.GetStateReturnsOnCall(2, nil, nil)
	err = admin.CreateCapitalAccount(
		transactionContext,
		"testAccountId",
		"testFundId",
		"testInvestorId",
		false,
		"0",
	)
	assert.Equal(t, err, smartcontracterrors.InvestorNotFoundError)
}

func TestCreateCapitalAccountAction(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	capitalAccount := types.CreateDefaultCapitalAccount(
		0,
		0,
		"testAccountId",
		"testFundId",
		"testInvestorId",
		false,
		"0",
	)
	capitalAccountJSON, err := capitalAccount.ToJSON()
	assert.Nil(t, err)
	chaincodeStub.GetStateReturns(capitalAccountJSON, nil)
	err = admin.CreateCapitalAccountAction(
		transactionContext,
		"testTransactionId",
		"testAccountId",
		"deposit",
		"100",
		false,
		"12-27-1996",
		0,
	)
	assert.Nil(t, err)
}

func TestCreateCapitalAccountActionMissingAccount(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturns(nil, nil)
	err := admin.CreateCapitalAccountAction(
		transactionContext,
		"testTransactionId",
		"testAccountId",
		"deposit",
		"100",
		false,
		"12-27-1996",
		0,
	)
	assert.Equal(t, err, smartcontracterrors.CapitalAccountNotFoundError)
}

func TestCreateCapitalAccountActionInvalidType(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	capitalAccount := types.CreateDefaultCapitalAccount(
		0,
		0,
		"testAccountId",
		"testFundId",
		"testInvestorId",
		false,
		"0",
	)
	capitalAccountJSON, err := capitalAccount.ToJSON()
	assert.Nil(t, err)
	chaincodeStub.GetStateReturns(capitalAccountJSON, nil)
	err = admin.CreateCapitalAccountAction(
		transactionContext,
		"testTransactionId",
		"testAccountId",
		"fake action",
		"100",
		false,
		"12-27-1996",
		0,
	)
	assert.Equal(t, err, smartcontracterrors.InvalidCapitalAccountActionTypeError)
}

func TestCreatePortfolio(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	fund := types.CreateDefaultFund("testFundId", "testFund", "testDate")
	fundJSON, err := fund.ToJSON()
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, nil, nil)
	chaincodeStub.GetStateReturnsOnCall(1, fundJSON, nil)
	err = admin.CreatePortfolio(
		transactionContext,
		"testPortfolioId",
		"testFundId",
		"testPortfolio",
	)
	assert.Nil(t, err)
}

func TestCreatePortfolioMissingFund(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturnsOnCall(0, nil, nil)
	chaincodeStub.GetStateReturnsOnCall(1, nil, nil)
	err := admin.CreatePortfolio(
		transactionContext,
		"testPortfolioId",
		"testFundId",
		"testPortfolio",
	)
	assert.Equal(t, err, smartcontracterrors.FundNotFoundError)
}

func TestCreatePortfolioAction(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	portfolio := types.CreateDefaultPortfolio("testPortfolioId", "testFundId", "testPortfolio")
	portfolioJSON, err := portfolio.ToJSON()
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, portfolioJSON, nil)
	err = admin.CreatePortfolioAction(
		transactionContext,
		"testActionId",
		"testPortfolioId",
		"buy",
		"12-27-1996",
		0,
		"AMZN",
		"testCusip",
		"100",
		"USD",
	)
	assert.Nil(t, err)
}

func TestCreatePortfolioActionMissingPortfolio(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturns(nil, nil)
	err := admin.CreatePortfolioAction(
		transactionContext,
		"testActionId",
		"testPortfolioId",
		"buy",
		"12-27-1996",
		0,
		"AMZN",
		"testCusip",
		"100",
		"USD",
	)
	assert.Equal(t, err, smartcontracterrors.PortfolioNotFoundError)
}

func TestCreatePortfolioActionInvalidAction(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	portfolio := types.CreateDefaultPortfolio("testPortfolioId", "testFundId", "testPortfolio")
	portfolioJSON, err := portfolio.ToJSON()
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, portfolioJSON, nil)
	err = admin.CreatePortfolioAction(
		transactionContext,
		"testActionId",
		"testPortfolioId",
		"fake action",
		"12-27-1996",
		0,
		"AMZN",
		"testCusip",
		"100",
		"USD",
	)
	assert.Equal(t, err, smartcontracterrors.InvalidPortfolioActionTypeError)
}

func TestBootstrapCapitalAccount(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}

	firstDeposit := types.CreateDefaultCapitalAccountAction(
		"testDeposit1",
		"testAccountId",
		"buy",
		"10000",
		false,
		"12-27-1996",
		0,
	)
	firstDepositJSON, err := json.Marshal(firstDeposit)
	assert.Nil(t, err)

	secondDeposit := types.CreateDefaultCapitalAccountAction(
		"testDeposit2",
		"testAccountId",
		"buy",
		"3000",
		false,
		"12-27-1996",
		0,
	)
	secondDepositJSON, err := json.Marshal(secondDeposit)
	assert.Nil(t, err)

	depositIterator := mocks.StateQueryIterator{}
	depositIterator.HasNextReturnsOnCall(0, true)
	depositIterator.HasNextReturnsOnCall(1, true)
	depositIterator.HasNextReturnsOnCall(2, false)
	depositIterator.NextReturnsOnCall(0, &queryresult.KV{
		Value: firstDepositJSON,
	}, nil)
	depositIterator.NextReturnsOnCall(1, &queryresult.KV{
		Value: secondDepositJSON,
	}, nil)

	withdrawalIterator := mocks.StateQueryIterator{}
	withdrawalIterator.HasNextReturnsOnCall(0, false)
	withdrawalIterator.CloseReturns(nil)

	chaincodeStub.GetQueryResultReturnsOnCall(0, &depositIterator, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(1, &withdrawalIterator, nil)

	capitalAccount := types.CreateDefaultCapitalAccount(
		0,
		0,
		"testAccountId",
		"testFundId",
		"testInvestorId",
		false,
		"0",
	)

	err = admin.BootstrapCapitalAccount(transactionContext, &capitalAccount)

	currentPeriod := capitalAccount.CurrentPeriod
	deposits := capitalAccount.Deposits[0]
	openingValue := capitalAccount.OpeningValue[0]

	assert.Equal(t, currentPeriod, 1)
	assert.Equal(t, deposits, "13000")
	assert.Equal(t, openingValue, "13000")
}

func TestBootstrapCapitalAccountInvalidPeriod(t *testing.T) {
	_, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	capitalAccount := types.CreateDefaultCapitalAccount(
		0,
		1,
		"testAccountId",
		"testFundId",
		"testInvestorId",
		false,
		"0",
	)
	err := admin.BootstrapCapitalAccount(transactionContext, &capitalAccount)
	assert.Equal(t, err, smartcontracterrors.CannotBootstrapCapitalAccountError)
}

func TestBootstrapCapitalAccountWithWithdrawals(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}

	firstDeposit := types.CreateDefaultCapitalAccountAction(
		"testDeposit1",
		"testAccountId",
		"deposit",
		"10000",
		false,
		"12-27-1996",
		0,
	)
	firstDepositJSON, err := json.Marshal(firstDeposit)
	assert.Nil(t, err)

	secondDeposit := types.CreateDefaultCapitalAccountAction(
		"testDeposit2",
		"testAccountId",
		"deposit",
		"3000",
		false,
		"12-27-1996",
		0,
	)
	secondDepositJSON, err := json.Marshal(secondDeposit)
	assert.Nil(t, err)

	depositIterator := mocks.StateQueryIterator{}
	depositIterator.HasNextReturnsOnCall(0, true)
	depositIterator.HasNextReturnsOnCall(1, true)
	depositIterator.HasNextReturnsOnCall(2, false)
	depositIterator.NextReturnsOnCall(0, &queryresult.KV{
		Value: firstDepositJSON,
	}, nil)
	depositIterator.NextReturnsOnCall(1, &queryresult.KV{
		Value: secondDepositJSON,
	}, nil)

	withdrawal := types.CreateDefaultCapitalAccountAction(
		"testDeposit2",
		"testAccountId",
		"withdrawal",
		"3000",
		false,
		"12-27-1996",
		0,
	)
	withdrawalJSON, err := json.Marshal(withdrawal)

	withdrawalIterator := mocks.StateQueryIterator{}
	withdrawalIterator.HasNextReturnsOnCall(0, true)
	withdrawalIterator.NextReturnsOnCall(0, &queryresult.KV{
		Value: withdrawalJSON,
	}, nil)

	chaincodeStub.GetQueryResultReturnsOnCall(0, &depositIterator, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(1, &withdrawalIterator, nil)

	capitalAccount := types.CreateDefaultCapitalAccount(
		0,
		0,
		"testAccountId",
		"testFundId",
		"testInvestorId",
		false,
		"0",
	)

	err = admin.BootstrapCapitalAccount(transactionContext, &capitalAccount)

	currentPeriod := capitalAccount.CurrentPeriod
	deposits := capitalAccount.Deposits[0]
	openingValue := capitalAccount.OpeningValue[0]

	assert.Equal(t, currentPeriod, 1)
	assert.Equal(t, deposits, "10000")
	assert.Equal(t, openingValue, "10000")
}

func TestBootstrapCapitalAccountInvalidWithdrawals(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}

	firstDeposit := types.CreateDefaultCapitalAccountAction(
		"testDeposit1",
		"testAccountId",
		"deposit",
		"10000",
		false,
		"12-27-1996",
		0,
	)
	firstDepositJSON, err := json.Marshal(firstDeposit)
	assert.Nil(t, err)

	secondDeposit := types.CreateDefaultCapitalAccountAction(
		"testDeposit2",
		"testAccountId",
		"deposit",
		"3000",
		false,
		"12-27-1996",
		0,
	)
	secondDepositJSON, err := json.Marshal(secondDeposit)
	assert.Nil(t, err)

	depositIterator := mocks.StateQueryIterator{}
	depositIterator.HasNextReturnsOnCall(0, true)
	depositIterator.HasNextReturnsOnCall(1, true)
	depositIterator.HasNextReturnsOnCall(2, false)
	depositIterator.NextReturnsOnCall(0, &queryresult.KV{
		Value: firstDepositJSON,
	}, nil)
	depositIterator.NextReturnsOnCall(1, &queryresult.KV{
		Value: secondDepositJSON,
	}, nil)

	withdrawal := types.CreateDefaultCapitalAccountAction(
		"testDeposit2",
		"testAccountId",
		"withdrawal",
		"14000",
		false,
		"12-27-1996",
		0,
	)
	withdrawalJSON, err := json.Marshal(withdrawal)

	withdrawalIterator := mocks.StateQueryIterator{}
	withdrawalIterator.HasNextReturnsOnCall(0, true)
	withdrawalIterator.NextReturnsOnCall(0, &queryresult.KV{
		Value: withdrawalJSON,
	}, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(0, &depositIterator, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(1, &withdrawalIterator, nil)
	capitalAccount := types.CreateDefaultCapitalAccount(
		0,
		0,
		"testAccountId",
		"testFundId",
		"testInvestorId",
		false,
		"0",
	)
	err = admin.BootstrapCapitalAccount(transactionContext, &capitalAccount)
	assert.Equal(t, err, smartcontracterrors.NegativeCapitalAccountBalanceError)
}

func TestBootstrapFund(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}

	//query fund
	fund := types.CreateDefaultFund("testId", "testFund", "12-27-1996")
	fundJSON, err := json.Marshal(fund)
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, fundJSON, nil)

	//query capital accounts
	capitalAccount1 := types.CreateDefaultCapitalAccount(
		0,
		0,
		"testAccountId1",
		"testFundId",
		"testInvestorId1",
		false,
		"0",
	)
	capitalAccount1JSON, err := json.Marshal(capitalAccount1)
	assert.Nil(t, err)
	capitalAccount2 := types.CreateDefaultCapitalAccount(
		1,
		0,
		"testAccountId2",
		"testFundId",
		"testInvestorId2",
		false,
		"0",
	)
	capitalAccount2JSON, err := json.Marshal(capitalAccount2)
	assert.Nil(t, err)
	capitalAccountIterator := mocks.StateQueryIterator{}
	capitalAccountIterator.HasNextReturnsOnCall(0, true)
	capitalAccountIterator.HasNextReturnsOnCall(1, true)
	capitalAccountIterator.HasNextReturnsOnCall(2, false)
	capitalAccountIterator.NextReturnsOnCall(0, &queryresult.KV{Value: capitalAccount1JSON}, nil)
	capitalAccountIterator.NextReturnsOnCall(1, &queryresult.KV{Value: capitalAccount2JSON}, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(0, &capitalAccountIterator, nil)

	//capital account 1 bootstrap
	firstDeposit := types.CreateDefaultCapitalAccountAction(
		"testDeposit1",
		"testAccountId1",
		"deposit",
		"10000",
		false,
		"12-27-1996",
		0,
	)
	firstDepositJSON, err := json.Marshal(firstDeposit)
	assert.Nil(t, err)
	depositIterator1 := mocks.StateQueryIterator{}
	depositIterator1.HasNextReturnsOnCall(0, true)
	depositIterator1.HasNextReturnsOnCall(1, false)
	depositIterator1.NextReturnsOnCall(0, &queryresult.KV{Value: firstDepositJSON}, nil)
	withdrawalIterator1 := mocks.StateQueryIterator{}
	withdrawalIterator1.HasNextReturnsOnCall(0, false)
	chaincodeStub.GetQueryResultReturnsOnCall(1, &depositIterator1, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(2, &withdrawalIterator1, nil)

	//capital account 2 bootstrap
	secondDeposit := types.CreateDefaultCapitalAccountAction(
		"testDeposit2",
		"testAccountId2",
		"deposit",
		"90000",
		false,
		"12-27-1996",
		0,
	)
	secondDepositJSON, err := json.Marshal(secondDeposit)
	assert.Nil(t, err)
	depositIterator2 := mocks.StateQueryIterator{}
	depositIterator2.HasNextReturnsOnCall(0, true)
	depositIterator2.HasNextReturnsOnCall(1, false)
	depositIterator2.NextReturnsOnCall(0, &queryresult.KV{Value: secondDepositJSON}, nil)
	withdrawalIterator2 := mocks.StateQueryIterator{}
	withdrawalIterator2.HasNextReturnsOnCall(0, false)
	chaincodeStub.GetQueryResultReturnsOnCall(3, &depositIterator2, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(4, &withdrawalIterator2, nil)

	//run the test
	resultFund, err := admin.BootstrapFund(transactionContext, "testId")
	assert.Nil(t, err)

	openingValue := resultFund.OpeningValues[0]
	deposits := resultFund.OpeningValues[0]

	assert.Equal(t, openingValue, "100000")
	assert.Equal(t, deposits, "100000")
}

func TestBootstrapFundInvalidPeriod(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	fund := types.CreateDefaultFund("testId", "testFund", "12-27-1996")
	fund.IncrementCurrentPeriod()
	fundJSON, err := json.Marshal(fund)
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, fundJSON, nil)
	result, err := admin.BootstrapFund(transactionContext, "testId")
	assert.Nil(t, result)
	assert.Equal(t, err, smartcontracterrors.CannotBootstrapFundError)
}

func TestStepFundInvalidPeriod(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	fund := types.CreateDefaultFund("testId", "testFund", "12-27-1996")
	fundJSON, err := json.Marshal(fund)
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, fundJSON, nil)
	_, err = admin.StepFund(transactionContext, "testId")
	assert.Equal(t, err, smartcontracterrors.CannotStepFundError)
}

func TestStepFundCannotReadWorldState(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturnsOnCall(0, nil, smartcontracterrors.ReadingWorldStateError)
	_, err := admin.StepFund(transactionContext, "testId")
	assert.Equal(t, err, smartcontracterrors.ReadingWorldStateError)
}

func TestStepFundInvalidFund(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturnsOnCall(0, nil, nil)
	_, err := admin.StepFund(transactionContext, "testId")
	assert.Equal(t, err, smartcontracterrors.FundNotFoundError)
}

func TestBootstrapFundCannotReadWorldState(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturnsOnCall(0, nil, smartcontracterrors.ReadingWorldStateError)
	_, err := admin.BootstrapFund(transactionContext, "testId")
	assert.Equal(t, err, smartcontracterrors.ReadingWorldStateError)
}

func TestBootstrapFundInvalidFund(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturnsOnCall(0, nil, nil)
	_, err := admin.BootstrapFund(transactionContext, "testId")
	assert.Equal(t, err, smartcontracterrors.FundNotFoundError)
}

func TestStepFund(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}

	fund := types.CreateDefaultFund("testFundId", "testFund", "12-27-1996")
	fund.IncrementCurrentPeriod() //step fund checks that it is not 0 which is the default value
	fundJSON, err := json.Marshal(fund)
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, fundJSON, nil)

	//create the first portfolio
	portfolio1 := types.CreateDefaultPortfolio("testPortfolioId", "testFundId", "testPortfolio1")
	portfolio1.MostRecentDate = "12-27-1996"
	cash := types.ValuedAsset{
		Name:     "cash",
		CUSIP:    "-1",
		Amount:   "100000",
		Currency: "USD",
		Price:    "1",
	}
	AAPL := types.ValuedAsset{
		Name:     "AAPL",
		CUSIP:    "037833100",
		Amount:   "300",
		Currency: "USD",
		Price:    "148.88",
	}
	portfolio1.Valuations = make(types.DateValuedAssetMap)
	portfolio1.Valuations[portfolio1.MostRecentDate] = make(types.ValuedAssetMap)
	portfolio1.Valuations[portfolio1.MostRecentDate]["cash"] = cash
	portfolio1.Valuations[portfolio1.MostRecentDate]["AAPL"] = AAPL
	portfolio1JSON, err := json.Marshal(portfolio1)
	assert.Nil(t, err)

	portfolioIterator := mocks.StateQueryIterator{}
	portfolioIterator.HasNextReturnsOnCall(0, true)
	portfolioIterator.HasNextReturnsOnCall(1, false)
	portfolioIterator.NextReturnsOnCall(0, &queryresult.KV{Value: portfolio1JSON}, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(0, &portfolioIterator, nil)

	//capital accounts
	capitalAccount1 := types.CreateDefaultCapitalAccount(
		0,
		0,
		"testAccountId1",
		"testFundId",
		"testInvestorId1",
		false,
		"0",
	)
	capitalAccount1.IncrementCurrentPeriod()
	capitalAccount1.OwnershipPercentage[0] = "0.1" //set the previous periods ownership
	capitalAccount1JSON, err := json.Marshal(capitalAccount1)
	assert.Nil(t, err)

	capitalAccount2 := types.CreateDefaultCapitalAccount(
		0,
		0,
		"testAccountId2",
		"testFundId",
		"testInvestorId2",
		false,
		"0",
	)
	capitalAccount2.IncrementCurrentPeriod()
	capitalAccount2.OwnershipPercentage[0] = "0.9"
	capitalAccount2.Number = 1
	capitalAccount2JSON, err := json.Marshal(capitalAccount2)
	assert.Nil(t, err)

	capitalAccountIterator := mocks.StateQueryIterator{}
	capitalAccountIterator.HasNextReturnsOnCall(0, true)
	capitalAccountIterator.HasNextReturnsOnCall(1, true)
	capitalAccountIterator.HasNextReturnsOnCall(2, false)
	capitalAccountIterator.NextReturnsOnCall(0, &queryresult.KV{Value: capitalAccount1JSON}, nil)
	capitalAccountIterator.NextReturnsOnCall(1, &queryresult.KV{Value: capitalAccount2JSON}, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(1, &capitalAccountIterator, nil)

	depositsIterator := mocks.StateQueryIterator{}
	depositsIterator.HasNextReturnsOnCall(0, false) //capital account 1
	depositsIterator.HasNextReturnsOnCall(1, true)  //capital account 2
	depositsIterator.HasNextReturnsOnCall(2, false) //capital account 2
	firstDeposit := types.CreateDefaultCapitalAccountAction(
		"testDeposit1",
		"testAccountId1",
		"deposit",
		"10000",
		false,
		"12-27-1996",
		1,
	)
	firstDepositJSON, err := json.Marshal(firstDeposit)
	assert.Nil(t, err)
	depositsIterator.NextReturnsOnCall(0, &queryresult.KV{Value: firstDepositJSON}, nil)
	withdrawalIterator := mocks.StateQueryIterator{}
	withdrawalIterator.HasNextReturnsOnCall(9, false)
	withdrawalIterator.HasNextReturnsOnCall(1, false)
	chaincodeStub.GetQueryResultReturnsOnCall(2, &depositsIterator, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(3, &withdrawalIterator, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(4, &depositsIterator, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(5, &withdrawalIterator, nil)
	result, err := admin.StepFund(transactionContext, "testFundId")
	assert.Nil(t, err)

	resultFund := result.Fund
	resultClosingValue := resultFund.ClosingValues[resultFund.PreviousPeriod()]
	resultOpeningValue := resultFund.OpeningValues[resultFund.PreviousPeriod()]
	resultFixedFees := resultFund.FixedFees[resultFund.PreviousPeriod()]
	resultDeposits := resultFund.Deposits[resultFund.PreviousPeriod()]
	assert.Equal(t, resultClosingValue, "144664")
	assert.Equal(t, resultOpeningValue, "154664")
	assert.Equal(t, resultFixedFees, "2603.952")
	assert.Equal(t, resultDeposits, "12603.952")
}

func TestStepFundNoCapitalAccounts(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}

	fund := types.CreateDefaultFund("testFundId", "testFund", "12-27-1996")
	fund.IncrementCurrentPeriod() //stepFund checks that it is not 0 which is the default value
	fundJSON, err := json.Marshal(fund)
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, fundJSON, nil)

	//create the first portfolio
	portfolio1 := types.CreateDefaultPortfolio("testPortfolioId", "testFundId", "testPortfolio1")
	portfolio1.MostRecentDate = "12-27-1996"
	cash := types.ValuedAsset{
		Name:     "cash",
		CUSIP:    "-1",
		Amount:   "100000",
		Currency: "USD",
		Price:    "1",
	}
	AAPL := types.ValuedAsset{
		Name:     "AAPL",
		CUSIP:    "037833100",
		Amount:   "300",
		Currency: "USD",
		Price:    "148.88",
	}
	portfolio1.Valuations = make(types.DateValuedAssetMap)
	portfolio1.Valuations[portfolio1.MostRecentDate] = make(types.ValuedAssetMap)
	portfolio1.Valuations[portfolio1.MostRecentDate]["cash"] = cash
	portfolio1.Valuations[portfolio1.MostRecentDate]["AAPL"] = AAPL
	portfolio1JSON, err := json.Marshal(portfolio1)
	assert.Nil(t, err)

	portfolioIterator := mocks.StateQueryIterator{}
	portfolioIterator.HasNextReturnsOnCall(0, true)
	portfolioIterator.HasNextReturnsOnCall(1, false)
	portfolioIterator.NextReturnsOnCall(0, &queryresult.KV{Value: portfolio1JSON}, nil)
	chaincodeStub.GetQueryResultReturnsOnCall(0, &portfolioIterator, nil)

	capitalAccountIterator := mocks.StateQueryIterator{}
	capitalAccountIterator.HasNextReturnsOnCall(0, false)
	chaincodeStub.GetQueryResultReturnsOnCall(1, &capitalAccountIterator, nil)

	result, err := admin.StepFund(transactionContext, "testFundId")
	assert.Nil(t, result)
	assert.Equal(t, err, smartcontracterrors.NoCapitalAccountsFoundError)
}

func TestCreateCapitalAccountNonZeroPeriod(t *testing.T) {
	testPeriod := 10
	capitalAccount := types.CreateDefaultCapitalAccount(
		0,
		testPeriod,
		"testId",
		"testFundId",
		"testInvestorId",
		false,
		"0",
	)
	for i := 0; i < testPeriod; i++ {
		assert.Equal(t, capitalAccount.ClosingValue[i], "0")
		assert.Equal(t, capitalAccount.OpeningValue[i], "0")
		assert.Equal(t, capitalAccount.FixedFees[i], "0")
		assert.Equal(t, capitalAccount.PerformanceFees[i], "0")
		assert.Equal(t, capitalAccount.Deposits[i], "0")
		assert.Equal(t, capitalAccount.OwnershipPercentage[i], "0")
	}
}

func TestCreateCapitalAccountActionMidYearDeposit(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	capitalAccount := types.CreateDefaultCapitalAccount(
		0,
		0,
		"testAccountId",
		"testFundId",
		"testInvestorId",
		true,
		"0",
	)
	capitalAccountJSON, err := capitalAccount.ToJSON()
	assert.Nil(t, err)
	fund := types.CreateDefaultFund("testFundId", "testFund", "12-27-1996")
	fundJSON, err := fund.ToJSON()
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, capitalAccountJSON, nil)
	chaincodeStub.GetStateReturnsOnCall(1, fundJSON, nil)
	err = admin.CreateCapitalAccountAction(
		transactionContext,
		"testTransactionId",
		"testAccountId",
		"deposit",
		"100",
		false,
		"12-27-1996",
		3,
	)
	assert.Equal(t, err, smartcontracterrors.MidYearDepositError)
}

func TestCreateCapitalAccountActionEndYearDeposit(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	capitalAccount := types.CreateDefaultCapitalAccount(
		0,
		0,
		"testAccountId",
		"testFundId",
		"testInvestorId",
		true,
		"0",
	)
	capitalAccountJSON, err := capitalAccount.ToJSON()
	assert.Nil(t, err)
	fund := types.CreateDefaultFund("testFundId", "testFund", "12-27-1996")
	fundJSON, err := fund.ToJSON()
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, capitalAccountJSON, nil)
	chaincodeStub.GetStateReturnsOnCall(1, fundJSON, nil)
	err = admin.CreateCapitalAccountAction(
		transactionContext,
		"testTransactionId",
		"testAccountId",
		"deposit",
		"100",
		false,
		"12-27-1996",
		12,
	)
	assert.Nil(t, err)
}
