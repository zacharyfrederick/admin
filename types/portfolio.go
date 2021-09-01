package types

//Map of assets using their name as a unique key
type AssetMap map[string]Asset

//Map of valued assets using their name as a unique key
type ValuedAssetMap map[string]ValuedAsset

//Map of dates to asset maps, represents a collection of portfolios snapshotted at a certain datea
type DateAssetMap map[string]AssetMap

//Map of dates to valued asset maps, represents a collection of valued portfolios snapshotted at a certain date
type DateValuedAssetMap map[string]ValuedAssetMap

type Portfolio struct {
	DocType    string             `json:"docType"`
	ID         string             `json:"id"`
	Fund       string             `json:"fund"`
	Name       string             `json:"name"`
	Securities []Asset            `json:"securities"`
	Assets     DateAssetMap       `json:"assets"`
	Valuations DateValuedAssetMap `json:"valuations"`
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
	Security    Asset  `json:"security"`
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

func ValidateCreatePortfolioActionRequest(r *CreatePortfolioActionRequest) bool {
	return true
}
