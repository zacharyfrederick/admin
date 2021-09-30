package types

type Benchmark struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Values map[string]string `json:"values"`
}
