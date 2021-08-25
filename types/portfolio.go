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
	Fund        string   `json:"fund"`
	Portfolio   string   `json:"portfolio"`
	Security    Security `json:"security"`
	Type        string   `json:"type"`
	Date        string   `json:"date"`
	Period      int      `json:"period"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
}
