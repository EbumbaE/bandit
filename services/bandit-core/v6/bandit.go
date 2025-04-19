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

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
}

type ArmParams struct {
	Count     uint64  `json:"count"`
	Mu        float64 `json:"mu"`
	SigmaSq   float64 `json:"sigma_sq"`
	Alpha     float64 `json:"alpha"`
	Beta      float64 `json:"beta"`
	TimeCount uint64  `json:"time_count"`
	Version   uint64  `json:"version"`
}

type GaussianBandit struct {
	BaseSigma         float64
	MinSigma          float64
	MinAlpha          float64
	DecayFactor       float64
	SigmaSmoothFactor float64
	Version           uint64
}

func NewDefaultGaussianBandit() *GaussianBandit {
	return &GaussianBandit{
		BaseSigma:         0.5,
		MinSigma:          0.1,
		MinAlpha:          1.0 + 1e-8,
		DecayFactor:       0.6,
		SigmaSmoothFactor: 0.3,
		Version:           1,
	}
}

func DefaultArmParams() *ArmParams {
	return &ArmParams{
		Mu:        0.0,
		SigmaSq:   1.0,
		Count:     0,
		TimeCount: 0,
		Alpha:     2.0,
		Beta:      1.0,
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

	oldCount := params.Count
	newCount := oldCount + 1
	newParams.Count = newCount
	newParams.Mu = (float64(oldCount)*params.Mu + decayedReward) / float64(newCount)

	delta := decayedReward - newParams.Mu
	deltaSq := delta * delta

	newParams.Alpha = max(gb.MinAlpha, params.Alpha+decayWeight/2)
	newParams.Beta = params.Beta + (deltaSq*decayWeight)/2

	newParams.SigmaSq = newParams.Beta / (newParams.Alpha - 1.0)
	newParams.SigmaSq = params.SigmaSq + gb.SigmaSmoothFactor*(newParams.SigmaSq-params.SigmaSq)
	newParams.SigmaSq = math.Max(newParams.SigmaSq, gb.MinSigma*gb.MinSigma)

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

		if params.Count > 0 {
			var sigmaSq float64
			params.Alpha = max(gb.MinAlpha, params.Alpha)
			sigmaSq = distuv.InverseGamma{Alpha: params.Alpha, Beta: params.Beta}.Rand()

			sigma = math.Sqrt(sigmaSq/float64(params.Count)) + gb.BaseSigma*math.Log(float64(params.TimeCount)+2)
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
	probs := make(map[string]Probability, len(arms))

	samples := make(map[string]float64)
	maxSample := -math.MaxFloat64

	for armID, params := range arms {
		var sample float64
		if params.Count == 0 {
			sample = distuv.Normal{Mu: 0, Sigma: gb.MinSigma}.Rand()
		} else {
			sigma := math.Sqrt(params.SigmaSq / float64(params.Count))
			sample = distuv.Normal{Mu: params.Mu, Sigma: sigma}.Rand()
		}

		samples[armID] = sample
		maxSample = max(maxSample, sample)
	}

	totalExpSample := float64(0.0)
	for armID, sample := range samples {
		expSample := math.Exp(sample - maxSample)
		probs[armID] = Probability{Score: expSample, Count: uint64(arms[armID].Count)}
		totalExpSample += expSample
	}

	for armID, prob := range probs {
		prob.Score /= totalExpSample
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

	if params.Count < 0 {
		return nil, errors.New("invalid N value in deserialized data")
	}
	if params.SigmaSq < 0 {
		return nil, errors.New("invalid SigmaSq value in deserialized data")
	}

	return &params, nil
}
