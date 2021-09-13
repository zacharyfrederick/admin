package errors

import "errors"

var ReadingWorldStateError = errors.New("error retrieving the world state")
var IdAlreadyInUseError = errors.New("an object already exists with that id")
var FundNotFoundError = errors.New("a fund with that id does not exist")
var PortfolioNotFoundError = errors.New("a portfolio with that id does not exist")
var InvestorNotFoundError = errors.New("an investor with that id does not exist")
var CapitalAccountNotFoundError = errors.New("a capital account with that id does not exist")
var InvalidPortfolioActionTypeError = errors.New("invalid portfolio action type")
var WritingWorldStateError = errors.New("error writing the world state")
var InvalidCapitalAccountActionTypeError = errors.New("invalid capital account action type")
var SaveStateError = errors.New("error saving the state to the blockchain")
var LoadStateError = errors.New("error loading the state from JSON")
var CannotBootstrapCapitalAccountError = errors.New("this capital account cannot be bootstrapped")
var NegativeCapitalAccountBalanceError = errors.New("the actions resulted in a negative capital account balance")
var CannotBootstrapFundError = errors.New("this fund cannot be bootstrapped")
var CannotStepFundError = errors.New("this fund cannot be stepped")
var NoPortfoliosFoundError = errors.New("no portfolios found for this fund")
var NoMostRecentDateForPortfolioError = errors.New("this portfolio does not have a most recent date")
var NoValuationsFoundForDateError = errors.New("no valuations found for date")
var DecimalConversionError = errors.New("error converting decimal")
