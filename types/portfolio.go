package types

type Portfolio struct {
	DocType    string     `json:"docType"`
	ID         string     `json:"id"`
	Fund       string     `json:"fund"`
	Name       string     `json:"name"`
	Securities []Security `json:"securities"`
}

type Security struct {
	Name     string `json:"name"`
	CUSIP    string `json:"cusip"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type PortfolioAction struct {
	DocType     string   `json:"docType"`
	ID          string   `json:"id"`
	Portfolio   string   `json:"portfolio"`
	Security    Security `json:"security"`
	Type        string   `json:"type"`
	Date        string   `json:"date"`
	Period      int      `json:"period"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
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
