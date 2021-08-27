package types

type Investor struct {
	DocType string `json:"docType"`
	ID      string `json:"id"`
	Name    string `json:"name"`
}

type CreateInvestorRequest struct {
	Name string `json:"name" binding:"required"`
}

func ValidateCreateInvestorRequest(r *CreateInvestorRequest) bool {
	return true
}
