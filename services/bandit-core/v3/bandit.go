package bandit

import (
	"encoding/json"
	"errors"
	"math"

	"gonum.org/v1/gonum/stat/distuv"
)

type BanditAlgorithm interface {
	Calculate(params *ArmParams, reward float64) *ArmParams
	Select(arms map[string]*ArmParams) string
}

type ArmParams struct {
	Mu        float64 `json:"mu"`
	SigmaSq   float64 `json:"sigma_sq"`
	N         int     `json:"n"`
	TimeCount int     `json:"time_count"`
}

type GaussianBandit struct {
	BaseSigma float64
	MinSigma  float64
}

func NewGaussianBandit(baseSigma, minSigma float64) *GaussianBandit {
	return &GaussianBandit{
		BaseSigma: baseSigma,
		MinSigma:  minSigma,
	}
}

func DefaultArmParams() *ArmParams {
	return &ArmParams{
		Mu:      0,
		SigmaSq: 1,
		N:       0,
	}
}

func (gb *GaussianBandit) Calculate(params *ArmParams, reward float64) *ArmParams {
	newParams := *params
	n := float64(newParams.N)
	newParams.N++

	newParams.Mu = (n*newParams.Mu + reward) / float64(newParams.N)

	if params.N >= 1 {
		newParams.SigmaSq = (n*newParams.SigmaSq +
			(reward-params.Mu)*(reward-newParams.Mu)) / float64(newParams.N)
	}

	newParams.TimeCount++
	return &newParams
}

func (gb *GaussianBandit) Select(arms map[string]*ArmParams) string {
	maxSample := -math.MaxFloat64
	selected := ""

	for id, params := range arms {
		sigma := math.Sqrt(params.SigmaSq/float64(params.N)) +
			gb.BaseSigma*math.Log(float64(params.TimeCount)+2)
		sigma = math.Max(sigma, gb.MinSigma)

		sample := distuv.Normal{Mu: params.Mu, Sigma: sigma}.Rand()
		if sample > maxSample {
			maxSample = sample
			selected = id
		}
	}
	return selected
}

func (gb *GaussianBandit) CalculateProbabilities(arms map[string]*ArmParams) map[string]float64 {
	total := 0.0
	samples := make(map[string]float64)

	for armID, params := range arms {
		var sample float64
		if params.N == 0 {
			sample = distuv.Normal{Mu: 0, Sigma: 1}.Rand()
		} else {
			sigma := math.Sqrt(params.SigmaSq / float64(params.N))
			sample = distuv.Normal{Mu: params.Mu, Sigma: sigma}.Rand()
		}
		samples[armID] = sample
		total += math.Exp(sample)
	}

	probs := make(map[string]float64)
	for armID, sample := range samples {
		probs[armID] = math.Exp(sample) / total
	}

	return probs
}

type ArmSerializer interface {
	Serialize(params *ArmParams) ([]byte, error)
	Deserialize(data []byte) (*ArmParams, error)
}

func NewArmSerializer() ArmSerializer {
	return &armSerializer{}
}

type armSerializer struct{}

func (s *armSerializer) Serialize(params *ArmParams) ([]byte, error) {
	if params == nil {
		return nil, errors.New("nil ArmParams provided")
	}

	data, err := json.Marshal(params)
	if err != nil {
		return nil, errors.New("marshal params")
	}
	return data, nil
}

func (s *armSerializer) Deserialize(data []byte) (*ArmParams, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data provided")
	}

	var params ArmParams
	if err := json.Unmarshal(data, &params); err != nil {
		return nil, errors.New("unmarshal params")
	}

	if params.N < 0 {
		return nil, errors.New("invalid N value in deserialized data")
	}
	if params.SigmaSq < 0 {
		return nil, errors.New("invalid SigmaSq value in deserialized data")
	}

	return &params, nil
}
