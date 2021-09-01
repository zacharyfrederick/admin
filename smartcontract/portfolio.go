package smartcontract

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/shopspring/decimal"
	"github.com/zacharyfrederick/admin/types"
	"github.com/zacharyfrederick/admin/utils"
)

func (s *AdminContract) CreatePortfolio(ctx contractapi.TransactionContextInterface, portfolioId string, fundId string, name string) error {
	idInUse, err := utils.AssetExists(ctx, portfolioId)
	if err != nil {
		return err
	}
	if idInUse {
		return fmt.Errorf("an object with that id already exists")
	}
	fund, err := s.QueryFundById(ctx, fundId)
	if err != nil {
		return err
	}
	if fund == nil {
		return fmt.Errorf("a fund with the specified id does not exist")
	}
	Portfolio := types.Portfolio{
		DocType:        types.DOCTYPE_PORTFOLIO,
		Name:           name,
		ID:             portfolioId,
		Fund:           fundId,
		MostRecentDate: "",
	}
	portfolioJson, err := json.Marshal(Portfolio)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(portfolioId, portfolioJson)
}

func (s *AdminContract) CreatePortfolioAction(ctx contractapi.TransactionContextInterface, actionId string, portfolioId string, type_ string, date string, period int, name string, cusip string, amount string, currency string) error {
	if type_ != "buy" && type_ != "sell" {
		return fmt.Errorf("the specified action is invalid for a portfolio: '%s'", type_)
	}
	portfolio, err := s.QueryPortfolioById(ctx, portfolioId)
	if err != nil {
		return err
	}
	if portfolio == nil {
		return errors.New("a portfolio with that ID does not exist")
	}
	security := types.Asset{
		Name:     name,
		CUSIP:    cusip,
		Amount:   amount,
		Currency: currency,
	}
	portfolioAction := &types.PortfolioAction{
		DocType:   types.DOCTYPE_PORTFOLIOACTION,
		Portfolio: portfolioId,
		Type:      type_,
		Date:      date,
		ID:        actionId,
		Security:  security,
		Period:    period,
		Status:    types.TX_STATUS_SUBMITTED,
	}
	portfolioActionJson, err := json.Marshal(portfolioAction)
	if err != nil {
		return err
	}
	err = executePortfolioAction(ctx, portfolio, portfolioAction)
	if err != nil {
		return err
	}
	portfolioJson, err := json.Marshal(portfolio)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(portfolioId, portfolioJson)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(actionId, portfolioActionJson)
}

func executePortfolioAction(ctx contractapi.TransactionContextInterface, portfolio *types.Portfolio, action *types.PortfolioAction) error {
	switch action.Type {
	case "buy":
		return buySecurityForPortfolio(ctx, portfolio, action)
	case "sell":
		return sellSecurityForPortfolio(ctx, portfolio, action)
	default:
		return errors.New("unrecognized action type")
	}
}

func buySecurityForPortfolio(ctx contractapi.TransactionContextInterface, portfolio *types.Portfolio, action *types.PortfolioAction) error {
	transactionDate := action.Date
	assetName := action.Security.Name
	if portfolio.MostRecentDate == "" {
		portfolio.Assets[transactionDate][assetName] = action.Security
		portfolio.MostRecentDate = action.Date
		return nil
	}
	currentAssets, err := getMostRecentAssetsForPortfolio(portfolio, transactionDate)
	if err != nil {
		return err
	}
	err = addAsset(currentAssets, action.Security)
	if err != nil {
		return err
	}
	portfolio.Assets[transactionDate] = currentAssets
	portfolio.MostRecentDate = action.Date
	return nil
}

func getMostRecentAssetsForPortfolio(portfolio *types.Portfolio, date string) (types.AssetMap, error) {
	currentAssets, ok := portfolio.Assets[date]
	if ok {
		return currentAssets, nil
	} else {
		currentAssets, ok := portfolio.Assets[portfolio.MostRecentDate]
		if ok {
			return currentAssets, nil
		} else {
			return nil, errors.New("no portfolio found for the most recent date")
		}
	}
}

func addAsset(assets types.AssetMap, assetToAdd types.Asset) error {
	currentAsset, ok := assets[assetToAdd.Name]
	if ok {
		currentAmount, err := decimal.NewFromString(currentAsset.Amount)
		if err != nil {
			return err
		}
		newAmount, err := decimal.NewFromString(assetToAdd.Amount)
		if err != nil {
			return err
		}
		totalAmount := currentAmount.Add(newAmount)
		currentAsset.Amount = totalAmount.String()
		assets[assetToAdd.Name] = currentAsset
	} else {
		assets[assetToAdd.Name] = assetToAdd
	}
	return nil
}

func removeAsset(assets types.AssetMap, assetToAdd types.Asset) error {
	currentAsset, ok := assets[assetToAdd.Name]
	if ok {
		currentAmount, err := decimal.NewFromString(currentAsset.Amount)
		if err != nil {
			return err
		}
		newAmount, err := decimal.NewFromString(assetToAdd.Amount)
		if err != nil {
			return err
		}
		totalAmount := currentAmount.Sub(newAmount)
		if totalAmount.Sign() == -1 {
			return errors.New("cannot have a negative security amount")
		}
		currentAsset.Amount = totalAmount.String()
		assets[assetToAdd.Name] = currentAsset
	} else {
		return errors.New("cannot have a negative security amount")
	}
	return nil
}

func sellSecurityForPortfolio(ctx contractapi.TransactionContextInterface, portfolio *types.Portfolio, action *types.PortfolioAction) error {
	transactionDate := action.Date
	if portfolio.MostRecentDate == "" {
		return errors.New("cannot sell a security from an empty portfolio")
	}
	currentAssets, err := getMostRecentAssetsForPortfolio(portfolio, transactionDate)
	if err != nil {
		return err
	}
	err = removeAsset(currentAssets, action.Security)
	if err != nil {
		return err
	}
	portfolio.Assets[transactionDate] = currentAssets
	portfolio.MostRecentDate = action.Date
	return nil
}

func (s *AdminContract) QueryPortfoliosByFund(ctx contractapi.TransactionContextInterface, fundId string) ([]*types.Portfolio, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType": "portfolio", "fund": "%s"}}`, fundId)
	return executePortfolioQuery(ctx, queryString)
}

func (s *AdminContract) QueryPortfolioById(ctx contractapi.TransactionContextInterface, capitalAccountId string) (*types.Portfolio, error) {
	data, err := ctx.GetStub().GetState(capitalAccountId)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	var portfolio types.Portfolio
	err = json.Unmarshal(data, &portfolio)
	if err != nil {
		return nil, err
	}
	return &portfolio, nil
}

func (s *AdminContract) QueryPortfolioActionById(ctx contractapi.TransactionContextInterface, capitalAccountId string) (*types.PortfolioAction, error) {
	data, err := ctx.GetStub().GetState(capitalAccountId)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	var portfolioAction types.PortfolioAction
	err = json.Unmarshal(data, &portfolioAction)
	if err != nil {
		return nil, err
	}
	return &portfolioAction, nil
}

func executePortfolioQuery(ctx contractapi.TransactionContextInterface, queryString string) ([]*types.Portfolio, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var portfolios []*types.Portfolio
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var portfolio types.Portfolio
		err = json.Unmarshal(queryResult.Value, &portfolio)
		if err != nil {
			return nil, err
		}
		portfolios = append(portfolios, &portfolio)
	}
	return portfolios, nil
}
