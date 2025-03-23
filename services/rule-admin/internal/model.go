package internal

type Rule struct {
	Id          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	State       StateType `db:"state"`
	Variants    []Variant
}

type StateType string

var (
	StateTypeEnable  StateType = "enable"
	StateTypeDisable StateType = "disable"
)

type Variant struct {
	Id    string    `db:"id"`
	Name  string    `db:"name"`
	Data  []byte    `db:"data"`
	State StateType `db:"state"`
}
