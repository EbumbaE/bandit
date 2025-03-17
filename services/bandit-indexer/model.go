package ruler

type Rule struct {
	ID        string
	Name      string
	BanditKey string
}

type Variant struct {
	ID     string
	RuleID string
	Data   string
}
