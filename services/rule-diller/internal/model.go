package internal

type Variant struct {
	Key   string
	Data  []byte
	Score float64
	Count uint64
}

type PayloadAnalitic struct {
	Service       string `json:"service"`
	Context       string `json:"context"`
	VariantID     string `json:"variant_id"`
	BanditVersion uint64 `json:"bandit_version"`
}

type Rule struct {
	Service  string
	Context  string
	Variants []Variant
	Version  uint64
}
