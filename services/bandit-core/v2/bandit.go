package bandit

import (
	"math"
	"sync"

	"gonum.org/v1/gonum/stat/distuv"
)

type Arm struct {
	Mu      float64
	SigmaSq float64
	N       int
	Mutex   sync.Mutex
}

type GaussianBandit struct {
	Arms        map[string]*Arm
	Mutex       sync.RWMutex
	BaseSigma   float64
	MinSigma    float64
	TimeCounter int
}

func NewGaussianBandit() *GaussianBandit {
	return &GaussianBandit{
		Arms:      make(map[string]*Arm),
		BaseSigma: 0.5,
		MinSigma:  0.1,
	}
}

func (gb *GaussianBandit) Key() string {
	return "gaussian"
}

func (gb *GaussianBandit) AddArm(armID string) {
	gb.Mutex.Lock()
	defer gb.Mutex.Unlock()
	gb.Arms[armID] = &Arm{
		Mu:      0,
		SigmaSq: 1,
		N:       0,
	}
}

func (gb *GaussianBandit) RemoveArm(armID string) {
	gb.Mutex.Lock()
	defer gb.Mutex.Unlock()
	delete(gb.Arms, armID)
}

func (gb *GaussianBandit) SelectArm() string {
	gb.Mutex.RLock()
	defer gb.Mutex.RUnlock()

	gb.TimeCounter++

	var selectedArmID string
	maxSample := -math.MaxFloat64

	for armID, arm := range gb.Arms {
		arm.Mutex.Lock()
		if arm.N == 0 {
			sample := distuv.Normal{Mu: 0, Sigma: 1}.Rand()
			arm.Mutex.Unlock()
			if sample > maxSample {
				maxSample = sample
				selectedArmID = armID
			}
			continue
		}

		baseSigma := gb.BaseSigma * math.Log(float64(gb.TimeCounter)+2)
		sigma := math.Sqrt(arm.SigmaSq/float64(arm.N)) + baseSigma
		sigma = math.Max(sigma, gb.MinSigma)

		sample := distuv.Normal{Mu: arm.Mu, Sigma: sigma}.Rand()
		arm.Mutex.Unlock()

		if sample > maxSample {
			maxSample = sample
			selectedArmID = armID
		}
	}

	return selectedArmID
}

func (gb *GaussianBandit) UpdateArm(armID string, reward float64) {
	gb.Mutex.RLock()
	defer gb.Mutex.RUnlock()

	arm, exists := gb.Arms[armID]
	if !exists {
		return
	}

	arm.Mutex.Lock()
	defer arm.Mutex.Unlock()

	n := float64(arm.N)
	newN := n + 1
	newMu := (arm.Mu*n + reward) / newN

	if arm.N >= 1 {
		newSigmaSq := (n*arm.SigmaSq + (reward-arm.Mu)*(reward-newMu)) / newN
		arm.SigmaSq = newSigmaSq
	}

	arm.Mu = newMu
	arm.N = int(newN)
}

func (gb *GaussianBandit) GetSelectionProbabilities() map[string]float64 {
	gb.Mutex.RLock()
	defer gb.Mutex.RUnlock()

	samples := make(map[string]float64)
	for armID, arm := range gb.Arms {
		arm.Mutex.Lock()
		if arm.N == 0 {
			samples[armID] = distuv.Normal{Mu: 0, Sigma: 1}.Rand()
		} else {
			sigma := math.Sqrt(arm.SigmaSq / float64(arm.N))
			samples[armID] = distuv.Normal{Mu: arm.Mu, Sigma: sigma}.Rand()
		}
		arm.Mutex.Unlock()
	}

	probabilities := make(map[string]float64)
	total := 0.0
	for _, sample := range samples {
		total += math.Exp(sample)
	}

	for armID, sample := range samples {
		probabilities[armID] = math.Exp(sample) / total
	}

	return probabilities
}
