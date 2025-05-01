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

const (
	GaussianBanditKey        = "gaussian"
	DefaultExplorationFactor = 0.1
)

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
}

type GaussianArm struct {
	Count     uint64  `json:"count"`
	Mu        float64 `json:"mu"`
	SigmaSq   float64 `json:"sigma_sq"`
	Alpha     float64 `json:"alpha"`
	Beta      float64 `json:"beta"`
	TimeCount uint64  `json:"time_count"`
	Version   uint64  `json:"version"`
}

type GaussianBandit struct {
	BaseSigma         float64 `json:"base_sigma"`
	MinSigma          float64 `json:"min_sigma"`
	MinAlpha          float64 `json:"min_alpha"`
	DecayFactor       float64 `json:"decay_factor"`
	SigmaSmoothFactor float64 `json:"sigma_smooth_factor"`
	Version           uint64  `json:"version"`
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

func NewDefaultGaussianArm() *GaussianArm {
	return &GaussianArm{
		Mu:        0.0,
		SigmaSq:   1.0,
		Count:     0,
		TimeCount: 0,
		Alpha:     2.0,
		Beta:      1.0,
	}
}

func (gb *GaussianBandit) Calculate(params *GaussianArm, reward float64, count uint64) *GaussianArm {
	newParams := *params

	if gb.Version < params.Version {
		fmt.Printf("Calculate: gb.Version < params.Version: %v, %v \n", gb, params)
		return params
	}

	versionDiff := gb.Version - params.Version
	decayWeight := math.Pow(gb.DecayFactor, float64(versionDiff))
	decayedReward := reward * decayWeight

	oldCount := params.Count
	newCount := oldCount + count
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

func (gb *GaussianBandit) Select(arms map[string]*GaussianArm) string {
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

type Probability struct {
	Score float64
	Count uint64
}

func (gb *GaussianBandit) CalculateProbabilities(arms map[string]*GaussianArm) map[string]Probability {
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

func (ga *GaussianArm) Serialize() ([]byte, error) {
	if ga == nil {
		return nil, errors.New("nil GaussianArm provided")
	}

	data, err := json.Marshal(ga)
	if err != nil {
		return nil, errors.New("marshal params")
	}
	return data, nil
}

func (ga *GaussianArm) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty data provided")
	}

	if err := json.Unmarshal(data, ga); err != nil {
		return errors.New("unmarshal params")
	}

	return nil
}

func (gb *GaussianBandit) Serialize() ([]byte, error) {
	if gb == nil {
		return nil, errors.New("nil GaussianBandit provided")
	}

	data, err := json.Marshal(gb)
	if err != nil {
		return nil, errors.New("marshal params")
	}
	return data, nil
}

func (gb *GaussianBandit) Deserialize(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty data provided")
	}

	if err := json.Unmarshal(data, gb); err != nil {
		return errors.New("unmarshal params")
	}

	return nil
}
