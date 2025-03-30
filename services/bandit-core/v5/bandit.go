package bandit

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

type ArmParams struct {
	Mu        float64 `json:"mu"`
	SigmaSq   float64 `json:"sigma_sq"`
	N         int     `json:"n"`
	TimeCount int     `json:"time_count"`
	Version   uint64  `json:"version"`
}

type GaussianBandit struct {
	BaseSigma         float64
	MinSigma          float64
	DecayFactor       float64
	SigmaSmoothFactor float64
	Version           uint64
}

func NewDefaultGaussianBandit() *GaussianBandit {
	return &GaussianBandit{
		BaseSigma:         0.5,
		MinSigma:          0.1,
		DecayFactor:       0.6,
		SigmaSmoothFactor: 0.3,
		Version:           1,
	}
}

func DefaultArmParams() *ArmParams {
	return &ArmParams{
		Mu:        0.0,
		SigmaSq:   1.0,
		N:         0.0,
		TimeCount: 0,
	}
}

func (gb *GaussianBandit) Calculate(params *ArmParams, reward float64) *ArmParams {
	newParams := *params

	if gb.Version < params.Version {
		fmt.Printf("Calculate: gb.Version < params.Version: %v, %v \n", gb, params)
		return params
	}

	versionDiff := gb.Version - params.Version
	decayWeight := math.Pow(gb.DecayFactor, float64(versionDiff))

	decayedReward := reward * decayWeight

	varianceWeight := math.Min(1.0, 1.0/(1.0+0.1*float64(versionDiff)))
	adjustedReward := reward * varianceWeight

	n := float64(newParams.N)
	newParams.N++
	newParams.Mu = (n*newParams.Mu + decayedReward) / float64(newParams.N)

	if params.N >= 1 {
		delta := adjustedReward - params.Mu
		deltaSq := delta * delta

		newSigmaSq := (n*params.SigmaSq + deltaSq*varianceWeight) / (n + varianceWeight)

		newParams.SigmaSq = params.SigmaSq + gb.SigmaSmoothFactor*(newSigmaSq-params.SigmaSq)
	}

	if params.Version == gb.Version {
		gb.Version++
	}

	newParams.TimeCount++
	newParams.Version = gb.Version
	return &newParams
}

func (gb *GaussianBandit) Select(arms map[string]*ArmParams) string {
	maxSample := -math.MaxFloat64
	selected := ""

	for id, params := range arms {
		sigma := gb.MinSigma

		if params.N > 0 {
			sigma = math.Sqrt(params.SigmaSq/float64(params.N)) + gb.BaseSigma*math.Log(float64(params.TimeCount)+2)
			sigma = math.Max(sigma, gb.MinSigma)
		}

		sample := distuv.Normal{Mu: params.Mu, Sigma: sigma}.Rand()
		if sample > maxSample || selected == "" {
			maxSample = sample
			selected = id
		}
	}
	return selected
}

const DefaultExplorationFactor = 0.1

type Probability struct {
	Score float64
	Count uint64
}

func SelectByProbabilities(options map[string]Probability, explorationFactor float64) string {
	if len(options) == 0 {
		return ""
	}

	rand.Seed(uint64(time.Now().UnixNano()))

	var totalCount uint64
	for _, opt := range options {
		totalCount += opt.Count
	}

	sumAdjusted := 0.0
	for key, opt := range options {
		opt.Score = opt.Score + explorationFactor*math.Sqrt(math.Log(float64(totalCount+1))/(float64(opt.Count)+1))

		sumAdjusted += opt.Score

		options[key] = opt
	}

	r := rand.Float64() * sumAdjusted
	cumulativeProb := 0.0

	var lastKey string
	for key, opt := range options {
		lastKey = key

		cumulativeProb += opt.Score / sumAdjusted

		if r <= cumulativeProb {
			opt.Count++
			options[key] = opt
			return key
		}
	}

	return lastKey
}

func (gb *GaussianBandit) CalculateProbabilities(arms map[string]*ArmParams) map[string]Probability {
	total := 0.0
	probs := make(map[string]Probability, len(arms))

	for armID, params := range arms {
		var sample float64
		if params.N == 0 {
			sample = distuv.Normal{Mu: 0, Sigma: 1}.Rand()
		} else {
			sigma := math.Sqrt(params.SigmaSq / float64(params.N))
			sample = distuv.Normal{Mu: params.Mu, Sigma: sigma}.Rand()
		}
		probs[armID] = Probability{Score: sample}
		total += math.Exp(sample)
	}

	for armID, prob := range probs {

		prob.Score = math.Exp(prob.Score) / total
		prob.Count = uint64(arms[armID].N)
		probs[armID] = prob
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
