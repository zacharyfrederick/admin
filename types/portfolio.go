package types

import (
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/zacharyfrederick/admin/types/doctypes"
)

//Map of assets using their name as a unique key
type AssetMap map[string]Asset

//Map of valued assets using their name as a unique key
type ValuedAssetMap map[string]ValuedAsset

//Map of dates to asset maps, represents a collection of portfolios snapshotted at a certain datea
type DateAssetMap map[string]AssetMap

//Map of dates to valued asset maps, represents a collection of valued portfolios snapshotted at a certain date
type DateValuedAssetMap map[string]ValuedAssetMap

type Portfolio struct {
	DocType        string             `json:"docType"`
	ID             string             `json:"id"`
	Fund           string             `json:"fund"`
	Name           string             `json:"name"`
	Securities     []Asset            `json:"securities"`
	Assets         DateAssetMap       `json:"assets"`
	Valuations     DateValuedAssetMap `json:"valuations"`
	MostRecentDate string             `json:"mostRecentDate"`
}

type Asset struct {
	Name     string `json:"name"`
	CUSIP    string `json:"cusip"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type ValuedAsset struct {
	Name     string `json:"name"`
	CUSIP    string `json:"cusip"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Price    string `json:"price"`
}

type PortfolioAction struct {
	DocType     string `json:"docType"`
	ID          string `json:"id"`
	Portfolio   string `json:"portfolio"`
	Asset       Asset  `json:"asset"`
	Type        string `json:"type"`
	Date        string `json:"date"`
	Period      int    `json:"period"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

type CreatePortfolioRequest struct {
	Fund string `json:"fund" binding:"required"`
	Name string `json:"name" binding:"required"`
}

func ValidateCreatePortfolioRequest(r *CreatePortfolioRequest) bool {
	return true
}

type CreatePortfolioActionRequest struct {
	Portfolio string `json:"portfolio" binding:"required"`
	Type      string `json:"type" binding:"required"`
	Date      string `json:"date" binding:"required"`
	Period    int    `json:"period" binding:"isdefault|required"`
	Name      string `json:"name" binding:"required"`
	CUSIP     string `json:"cusip" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
	Currency  string `json:"currency" binding:"required"`
}

type ValuePortfolioRequest struct {
	Portfolio string `json:"portfolio" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Date      string `json:"date" binding:"required"`
	Price     string `json:"price" binding:"isdefault|required"`
}

func ValidateCreatePortfolioActionRequest(r *CreatePortfolioActionRequest) bool {
	return true
}

func ValidateValuePortfolioRequest(r *ValuePortfolioRequest) bool {
	return true
}

func CreateDefaultPortfolio(portfolioId string, fundId string, name string) Portfolio {
	portfolio := Portfolio{
		DocType:        doctypes.DOCTYPE_PORTFOLIO,
		Name:           name,
		ID:             portfolioId,
		Fund:           fundId,
		MostRecentDate: "",
	}
	return portfolio
}

func (p *Portfolio) ToJSON() ([]byte, error) {
	portfolioJSON, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return portfolioJSON, nil
}
func (p *Portfolio) SaveState(ctx contractapi.TransactionContextInterface) error {
	portfolioJSON, err := p.ToJSON()
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(p.ID, portfolioJSON)
}

func (p *PortfolioAction) SaveState(ctx contractapi.TransactionContextInterface) error {
	portfolioActionJSON, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(p.ID, portfolioActionJSON)
}

func CreateAsset(name string, cusip string, amount string, currency string) Asset {
	security := Asset{
		Name:     name,
		CUSIP:    cusip,
		Amount:   amount,
		Currency: currency,
	}
	return security
}

func CreateDefaultPortfolioAction(portfolioId string, type_ string, date string, id string, asset Asset, period int) PortfolioAction {
	portfolioAction := PortfolioAction{
		DocType:   doctypes.DOCTYPE_PORTFOLIOACTION,
		Portfolio: portfolioId,
		Type:      type_,
		Date:      date,
		ID:        id,
		Asset:     asset,
		Period:    period,
		Status:    TX_STATUS_SUBMITTED,
	}
	return portfolioAction
}
