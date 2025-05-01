package internal

type Bandit struct {
	RuleId    string    `db:"rule_id"`
	Version   uint64    `db:"version"`
	Config    []byte    `db:"config"`
	BanditKey string    `db:"bandit_key"`
	State     StateType `db:"state"`

	Arms []Arm
}

type Arm struct {
	VariantId string    `db:"variant_id"`
	Count     uint64    `db:"count"`
	Config    []byte    `db:"config"`
	State     StateType `db:"state"`

	Score float64
}

type StateType string

var (
	StateTypeEnable  StateType = "enabled"
	StateTypeDisable StateType = "disabled"
)
