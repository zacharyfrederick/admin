package smartcontract

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
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
		DocType: types.DOCTYPE_PORTFOLIO,
		Name:    name,
		ID:      portfolioId,
		Fund:    fundId,
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
	security := types.Asset{
		Name:     name,
		CUSIP:    cusip,
		Amount:   amount,
		Currency: currency,
	}
	portfolioAction := types.PortfolioAction{
		DocType:   types.DOCTYPE_PORTFOLIOACTION,
		Portfolio: portfolioId,
		Type:      type_,
		Date:      date,
		ID:        actionId,
		Security:  security,
		Period:    period,
	}
	portfolioActionJson, err := json.Marshal(portfolioAction)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(actionId, portfolioActionJson)
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
