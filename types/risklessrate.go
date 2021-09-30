package types

type RisklessRate struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Values map[string]string `json:"values"`
}
