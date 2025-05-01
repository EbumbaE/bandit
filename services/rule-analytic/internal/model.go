package internal

type PayloadAnalitic struct {
	Service     string `json:"service"`
	Context     string `json:"context"`
	RuleID      string `json:"rule_id"`
	VariantID   string `json:"variant_id"`
	RuleVersion uint64 `json:"rule_version"`
}

type HistoryEvent struct {
	Payload PayloadAnalitic
	Action  ActionType
	Amount  float64
}

type BanditEvent struct {
	RuleID      string  `json:"rule_id"`
	VariantID   string  `json:"variant_id"`
	Reward      float64 `json:"reward"`
	Count       uint64  `json:"count"`
	RuleVersion uint64  `json:"rule_version"`
}

type ActionType string

var (
	ClickActionType    = ActionType("click")
	ViewActionType     = ActionType("view")
	CartActionType     = ActionType("add_to_cart")
	PurchaseActionType = ActionType("purchase")
)

func (a ActionType) String() string {
	return string(a)
}
