package smartcontract_test

import (
	"errors"
	"testing"

	"github.com/zacharyfrederick/admin/smartcontract"
	"github.com/zacharyfrederick/admin/types"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
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
	err = admin.CreateCapitalAccount(transactionContext, "testAccountId", "testFundId", "testInvestorId")
	assert.Nil(t, err)
}

func TestCreateCapitalAccountExistingId(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturnsOnCall(0, []byte("fake data"), nil)
	err := admin.CreateCapitalAccount(transactionContext, "testAccountId", "testFundId", "testInvestorId")
	assert.Equal(t, err, smartcontracterrors.IdAlreadyInUseError)
}

func TestCreateCapitalAccountMissingFund(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturnsOnCall(0, nil, nil)
	chaincodeStub.GetStateReturnsOnCall(1, nil, nil)
	err := admin.CreateCapitalAccount(transactionContext, "testAccountId", "testFundId", "testInvestorId")
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
	err = admin.CreateCapitalAccount(transactionContext, "testAccountId", "testFundId", "testInvestorId")
	assert.Equal(t, err, smartcontracterrors.InvestorNotFoundError)
}

func TestCreateCapitalAccountAction(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	capitalAccount := types.CreateDefaultCapitalAccount(0, 0, "testAccountId", "testFundId", "testInvestorId")
	capitalAccountJSON, err := capitalAccount.ToJSON()
	assert.Nil(t, err)
	chaincodeStub.GetStateReturns(capitalAccountJSON, nil)
	err = admin.CreateCapitalAccountAction(transactionContext, "testTransactionId", "testAccountId", "deposit", "100", false, "12-27-1996", 0)
	assert.Nil(t, err)
}

func TestCreateCapitalAccountActionMissingAccount(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturns(nil, nil)
	err := admin.CreateCapitalAccountAction(transactionContext, "testTransactionId", "testAccountId", "deposit", "100", false, "12-27-1996", 0)
	assert.Equal(t, err, smartcontracterrors.CapitalAccountNotFoundError)
}

func TestCreateCapitalAccountActionInvalidType(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	capitalAccount := types.CreateDefaultCapitalAccount(0, 0, "testAccountId", "testFundId", "testInvestorId")
	capitalAccountJSON, err := capitalAccount.ToJSON()
	assert.Nil(t, err)
	chaincodeStub.GetStateReturns(capitalAccountJSON, nil)
	err = admin.CreateCapitalAccountAction(transactionContext, "testTransactionId", "testAccountId", "fake action", "100", false, "12-27-1996", 0)
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
	err = admin.CreatePortfolio(transactionContext, "testPortfolioId", "testFundId", "testPortfolio")
	assert.Nil(t, err)
}

func TestCreatePortfolioMissingFund(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturnsOnCall(0, nil, nil)
	chaincodeStub.GetStateReturnsOnCall(1, nil, nil)
	err := admin.CreatePortfolio(transactionContext, "testPortfolioId", "testFundId", "testPortfolio")
	assert.Equal(t, err, smartcontracterrors.FundNotFoundError)
}

func TestCreatePortfolioAction(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	portfolio := types.CreateDefaultPortfolio("testPortfolioId", "testFundId", "testPortfolio")
	portfolioJSON, err := portfolio.ToJSON()
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, portfolioJSON, nil)
	err = admin.CreatePortfolioAction(transactionContext, "testActionId", "testPortfolioId", "buy", "12-27-1996", 0, "AMZN", "testCusip", "100", "USD")
	assert.Nil(t, err)
}

func TestCreatePortfolioActionMissingPortfolio(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	chaincodeStub.GetStateReturns(nil, nil)
	err := admin.CreatePortfolioAction(transactionContext, "testActionId", "testPortfolioId", "buy", "12-27-1996", 0, "AMZN", "testCusip", "100", "USD")
	assert.Equal(t, err, smartcontracterrors.PortfolioNotFoundError)
}

func TestCreatePortfolioActionInvalidAction(t *testing.T) {
	chaincodeStub, transactionContext := prepareTest()
	admin := smartcontract.AdminContract{}
	portfolio := types.CreateDefaultPortfolio("testPortfolioId", "testFundId", "testPortfolio")
	portfolioJSON, err := portfolio.ToJSON()
	assert.Nil(t, err)
	chaincodeStub.GetStateReturnsOnCall(0, portfolioJSON, nil)
	err = admin.CreatePortfolioAction(transactionContext, "testActionId", "testPortfolioId", "fake action", "12-27-1996", 0, "AMZN", "testCusip", "100", "USD")
	assert.Equal(t, err, smartcontracterrors.InvalidPortfolioActionTypeError)
}
