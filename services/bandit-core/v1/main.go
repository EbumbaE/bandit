package main

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

type Arm struct {
	Alpha float64
	Beta  float64
	Mutex sync.Mutex
}

type Bandit struct {
	Arms  map[string]*Arm
	Mutex sync.RWMutex
}

func NewBandit() *Bandit {
	return &Bandit{
		Arms: make(map[string]*Arm),
	}
}

func (b *Bandit) AddArm(armID string) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	b.Arms[armID] = &Arm{Alpha: 1, Beta: 1}
}

func (b *Bandit) SelectArm() (string, float64) {
	b.Mutex.RLock()
	defer b.Mutex.RUnlock()

	var selectedArmID string
	maxSample := -1.0

	for armID, arm := range b.Arms {
		arm.Mutex.Lock()
		betaDist := distuv.Beta{Alpha: arm.Alpha, Beta: arm.Beta}
		sample := betaDist.Rand()
		arm.Mutex.Unlock()

		if sample > maxSample {
			maxSample = sample
			selectedArmID = armID
		}
	}

	return selectedArmID, maxSample
}

func (b *Bandit) UpdateArm(armID string, reward float64) {
	b.Mutex.RLock()
	defer b.Mutex.RUnlock()

	arm, exists := b.Arms[armID]
	if !exists {
		return
	}

	arm.Mutex.Lock()
	defer arm.Mutex.Unlock()

	if reward >= 0 {
		arm.Alpha += reward
	} else {
		arm.Beta += 1
	}
}

func (b *Bandit) GetScores() map[string]float64 {
	b.Mutex.RLock()
	defer b.Mutex.RUnlock()

	scores := make(map[string]float64)
	for armID, arm := range b.Arms {
		arm.Mutex.Lock()
		score := arm.Alpha / (arm.Alpha + arm.Beta)
		arm.Mutex.Unlock()
		scores[armID] = score
	}

	return scores
}

func (b *Bandit) SelectArmRandom() (string, float64) {
	b.Mutex.RLock()
	defer b.Mutex.RUnlock()

	scores := make(map[string]float64)
	totalScore := 0.0
	for armID, arm := range b.Arms {
		arm.Mutex.Lock()
		score := arm.Alpha / (arm.Alpha + arm.Beta)
		arm.Mutex.Unlock()
		scores[armID] = score
		totalScore += score
	}

	normalizedScores := make(map[string]float64)
	for armID, score := range scores {
		normalizedScores[armID] = score / totalScore
	}

	randomValue := rand.Float64()
	cumulativeProbability := 0.0

	for armID, prob := range normalizedScores {
		cumulativeProbability += prob
		if randomValue <= cumulativeProbability {
			return armID, scores[armID]
		}
	}

	for armID := range scores {
		return armID, scores[armID]
	}

	return "", 0.0
}

func main() {
	bandit := NewBandit()
	bandit.AddArm("arm1")
	bandit.AddArm("arm2")

	rand.Seed(uint64(time.Now().UnixNano()))

	for i := 0; i < 1000; i++ {
		selectedArmID, score := bandit.SelectArmRandom()
		fmt.Printf("Selected Arm: %s (Score: %.2f)\n", selectedArmID, score)

		reward := 0.0
		if selectedArmID == "arm1" {
			reward = 0.7
		} else {
			reward = 0.3
		}

		bandit.UpdateArm(selectedArmID, reward)
	}

	scores := bandit.GetScores()
	fmt.Println("Final Scores:", scores)
}
