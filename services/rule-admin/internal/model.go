package internal

type Rule struct {
	Id          string
	Name        string
	Description string
	State       StateType
	Variants    []Variant
}

type StateType string

var (
	StateTypeEnable  StateType = "enable"
	StateTypeDisable StateType = "disable"
)

type Variant struct {
	Id    string
	Data  []byte
	State StateType
}
