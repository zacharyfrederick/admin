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

func (s *AdminContract) CreatePortfolioAction(ctx contractapi.TransactionContextInterface, actionId string, fundId string, portfolioId string, type_ string, date string, period int, name string, cusip string, amount string, currency string) error {
	if type_ != "buy" && type_ != "sell" {
		return fmt.Errorf("the specified action is invalid for a portfolio: '%s'", type_)
	}

	security := types.Security{
		Name:     name,
		CUSIP:    cusip,
		Amount:   amount,
		Currency: currency,
	}

	portfolioAction := types.PortfolioAction{
		DocType:   types.DOCTYPE_PORTFOLIOACTION,
		Portfolio: portfolioId,
		Fund:      fundId,
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

func (s *AdminContract) QueryPortfolioByName(ctx contractapi.TransactionContextInterface, fundId string, name string) (*types.Portfolio, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType": "portfolio", "fund": "%s", "name": "%s"}}`, fundId, name)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var portfolios types.Portfolio
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(queryResult.Value, &portfolios)
		if err != nil {
			return nil, err
		}

		if true {
			break
		}
	}

	return &portfolios, nil
}

func (s *AdminContract) QueryPortfolioByFund(ctx contractapi.TransactionContextInterface, fundId string) (*types.Portfolio, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType": "portfolio", "fund": "%s"}}`, fundId)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var portfolios types.Portfolio
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(queryResult.Value, &portfolios)
		if err != nil {
			return nil, err
		}

		if true {
			break
		}
	}

	return &portfolios, nil
}
