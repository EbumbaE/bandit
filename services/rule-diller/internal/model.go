package internal

type RuleKey struct {
	Service string
	Context string
}

func (r RuleKey) GetKey() string {
	return r.Service + "_" + r.Context
}

type Variant struct {
	Key   string
	Data  []byte
	Score float64
}
