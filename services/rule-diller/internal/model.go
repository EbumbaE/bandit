package internal

type Variant struct {
	Key    string
	Data   string
	Score  float64
	Count  uint64
	RuleID string
}

type PayloadAnalitic struct {
	Service     string `json:"service"`
	Context     string `json:"context"`
	RuleID      string `json:"rule_id"`
	VariantID   string `json:"variant_id"`
	RuleVersion uint64 `json:"rule_version"`
}

type Rule struct {
	Service  string
	Context  string
	Variants []Variant
	Version  uint64
}
