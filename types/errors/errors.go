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
