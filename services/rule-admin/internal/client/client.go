package client

type Client interface{}

type RuleDillerWrapper struct {
	cl Client
}

func NewRuleDillerWrapper(cl Client) *RuleDillerWrapper {
	return &RuleDillerWrapper{
		cl: cl,
	}
}
