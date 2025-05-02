package internal

type Rule struct {
	Id          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	State       StateType `db:"state"`
	BanditKey   string    `db:"bandit_key"`
	Service     string    `db:"service"`
	Context     string    `db:"context"`

	Variants []Variant
}

type StateType string

var (
	StateTypeEnable  StateType = "enabled"
	StateTypeDisable StateType = "disabled"
)

type Variant struct {
	Id    string    `db:"id"`
	Name  string    `db:"name"`
	Data  string    `db:"data"`
	State StateType `db:"state"`
}

type WantedBandit struct {
	BanditKey string `db:"bandit_key"`
	Name      string `db:"name"`
}
