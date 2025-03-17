package domain

type Proof struct {
	Leaf     string   `json:"leaf"`
	Root     string   `json:"root"`
	Siblings []string `json:"siblings"`
	Position int      `json:"positions"`
}
