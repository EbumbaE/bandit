package internal

type Bandit struct {
	Id      string `db:"id"`
	Service string `db:"service"`
	Context string `db:"context"`
	Version uint64 `db:"version"`
	Arms    []Arm

	RuleId    string    `db:"rule_id"`
	Config    []byte    `db:"config"`
	BanditKey string    `db:"bandit_key"`
	State     StateType `db:"state"`
}

type Arm struct {
	Id    string `db:"id"`
	Data  []byte `db:"data"`
	Score float64
	Count uint64 `db:"count"`

	VariantId string    `db:"variant_id"`
	Config    []byte    `db:"config"`
	State     StateType `db:"state"`
}

type StateType string

var (
	StateTypeEnable  StateType = "enable"
	StateTypeDisable StateType = "disable"
)
