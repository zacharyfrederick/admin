package smartcontract

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/zacharyfrederick/admin/types"
	smartcontracterrors "github.com/zacharyfrederick/admin/types/errors"
	"github.com/zacharyfrederick/admin/utils"
)

func (s *AdminContract) CreatePortfolio(
	ctx SmartContractContext,
	portfolioId string,
	fundId string,
	name string,
) error {
	idInUse, err := utils.AssetExists(ctx, portfolioId)
	if err != nil {
		return smartcontracterrors.ReadingWorldStateError
	}
	if idInUse {
		return smartcontracterrors.IdAlreadyInUseError
	}
	fund, err := s.QueryFundById(ctx, fundId)
	if err != nil {
		return smartcontracterrors.ReadingWorldStateError
	}
	if fund == nil {
		return smartcontracterrors.FundNotFoundError
	}
	portfolio := types.CreateDefaultPortfolio(portfolioId, fundId, name)
	return SaveState(ctx, &portfolio)
}

func (s *AdminContract) CreatePortfolioAction(
	ctx SmartContractContext,
	actionId string,
	portfolioId string,
	type_ string,
	date string,
	period int,
	name string,
	cusip string,
	amount string,
	currency string,
) error {
	if type_ != "buy" && type_ != "sell" {
		return smartcontracterrors.InvalidPortfolioActionTypeError
	}
	portfolio, err := s.QueryPortfolioById(ctx, portfolioId)
	if err != nil {
		return smartcontracterrors.ReadingWorldStateError
	}
	if portfolio == nil {
		return smartcontracterrors.PortfolioNotFoundError
	}
	asset := types.CreateAsset(name, cusip, amount, currency)
	portfolioAction := types.CreateDefaultPortfolioAction(
		portfolioId,
		type_,
		date,
		actionId,
		asset,
		period,
	)
	err = executePortfolioAction(portfolio, &portfolioAction)
	if err != nil {
		return err
	}
	err = SaveState(ctx, portfolio)
	if err != nil {
		return err
	}
	return SaveState(ctx, &portfolioAction)
}

func (s *AdminContract) UpdatePortfolioValuation(
	ctx SmartContractContext,
	portfolioId string,
	date string,
	name string,
	price string,
) error {
	portfolio, err := s.QueryPortfolioById(ctx, portfolioId)
	if err != nil {
		return err
	}
	if portfolio == nil {
		return errors.New("a portfolio with that ID does not exist")
	}
	currentAssets, ok := portfolio.Assets[date]
	if !ok {
		return errors.New("a portfolio snapshop with that date does not exist")
	}
	asset, ok := currentAssets[name]
	if !ok {
		return errors.New("the specified asset is not in the portfolio")
	}
	//initialize the valuations if it hasn't been created yet
	if portfolio.Valuations == nil {
		portfolio.Valuations = make(types.DateValuedAssetMap)
	}
	valuedAsset := createValuedAsset(asset, price)
	valuationForDate, ok := portfolio.Valuations[date]
	if !ok { // no valuations for date
		portfolio.Valuations[date] = make(types.ValuedAssetMap)
		portfolio.Valuations[date][name] = valuedAsset
	} else {
		currentValuation, ok := valuationForDate[name]
		if ok {
			currentValuation.Price = price
			valuationForDate[name] = currentValuation
		} else {
			valuationForDate[name] = valuedAsset
		}
	}
	portfolioJson, err := json.Marshal(portfolio)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(portfolioId, portfolioJson)
}

func createValuedAsset(asset types.Asset, price string) types.ValuedAsset {
	valuedAsset := types.ValuedAsset{
		Name:     asset.Name,
		CUSIP:    asset.CUSIP,
		Amount:   asset.Amount,
		Currency: asset.Currency,
		Price:    price,
	}
	return valuedAsset
}

func executePortfolioAction(portfolio *types.Portfolio, action *types.PortfolioAction) error {
	switch action.Type {
	case "buy":
		return buySecurityForPortfolio(portfolio, action)
	case "sell":
		return sellSecurityForPortfolio(portfolio, action)
	default:
		return smartcontracterrors.InvalidPortfolioActionTypeError
	}
}

func buySecurityForPortfolio(portfolio *types.Portfolio, action *types.PortfolioAction) error {
	transactionDate := action.Date
	assetName := action.Asset.Name
	if portfolio.MostRecentDate == "" {
		portfolio.Assets = make(types.DateAssetMap)
		portfolio.Assets[transactionDate] = make(types.AssetMap)
		portfolio.Assets[transactionDate][assetName] = action.Asset
		portfolio.MostRecentDate = action.Date
		return nil
	}
	currentAssets, err := getMostRecentAssetsForPortfolio(portfolio, transactionDate)
	if err != nil {
		return err
	}
	err = addAsset(currentAssets, action.Asset)
	if err != nil {
		return err
	}
	portfolio.Assets[transactionDate] = currentAssets
	portfolio.MostRecentDate = action.Date
	return nil
}

func getMostRecentAssetsForPortfolio(
	portfolio *types.Portfolio,
	date string,
) (types.AssetMap, error) {
	currentAssets, ok := portfolio.Assets[date]
	if ok {
		return currentAssets, nil
	} else {
		currentAssets, ok := portfolio.Assets[portfolio.MostRecentDate]
		if ok {
			newAssets := copyAssetMap(currentAssets)
			return newAssets, nil
		} else {
			return nil, errors.New("no portfolio found for the most recent date")
		}
	}
}

func copyAssetMap(assets types.AssetMap) types.AssetMap {
	newMap := make(types.AssetMap)
	for k, v := range assets {
		newMap[k] = v
	}
	return newMap
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

func sellSecurityForPortfolio(portfolio *types.Portfolio, action *types.PortfolioAction) error {
	transactionDate := action.Date
	if portfolio.MostRecentDate == "" {
		return errors.New("cannot sell a security from an empty portfolio")
	}
	currentAssets, err := getMostRecentAssetsForPortfolio(portfolio, transactionDate)
	if err != nil {
		return err
	}
	err = removeAsset(currentAssets, action.Asset)
	if err != nil {
		return err
	}
	portfolio.Assets[transactionDate] = currentAssets
	portfolio.MostRecentDate = action.Date
	return nil
}

func (s *AdminContract) QueryPortfoliosByFund(
	ctx SmartContractContext,
	fundId string,
) ([]*types.Portfolio, error) {
	return queryPortfoliosByFund(ctx, fundId)
}

func queryPortfoliosByFund(ctx SmartContractContext, fundId string) ([]*types.Portfolio, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType": "portfolio", "fund": "%s"}}`, fundId)
	return executePortfolioQuery(ctx, queryString)
}

func (s *AdminContract) QueryPortfolioById(
	ctx SmartContractContext,
	capitalAccountId string,
) (*types.Portfolio, error) {
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

func (s *AdminContract) QueryPortfolioActionById(
	ctx SmartContractContext,
	capitalAccountId string,
) (*types.PortfolioAction, error) {
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

func executePortfolioQuery(
	ctx SmartContractContext,
	queryString string,
) ([]*types.Portfolio, error) {
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
