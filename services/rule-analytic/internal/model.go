package internal

type PayloadAnalitic struct {
	Service     string `json:"service"`
	Context     string `json:"context"`
	RuleID      string `json:"rule_id"`
	VariantID   string `json:"variant_id"`
	RuleVersion uint64 `json:"rule_version"`
}

type HistoryEvent struct {
	Payload PayloadAnalitic `json:"payload"`
	Action  ActionType      `json:"action"`
	Amount  float64         `json:"amount"`
}

type BanditEvent struct {
	RuleID      string  `json:"rule_id"`
	VariantID   string  `json:"variant_id"`
	Reward      float64 `json:"reward"`
	RuleVersion uint64  `json:"rule_version"`
}

type ActionType string

var (
	ClickActionType    = ActionType("click")
	ViewActionType     = ActionType("view")
	CartActionType     = ActionType("add_to_cart")
	PurchaseActionType = ActionType("purchase")
)
