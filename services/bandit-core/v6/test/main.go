package main

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"sync"
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	bandit "github.com/EbumbaE/bandit/services/bandit-core/v6"
)

type ArmHistory struct {
	Mu            []float64
	SigmaSq       []float64
	Count         []uint64
	Probabilities []float64
	Steps         []int
}

func main() {
	gb := bandit.NewDefaultGaussianBandit()

	memoryStorage := &MemoryStorage{
		data:          make(map[string]map[string][]byte),
		armSerializer: bandit.NewArmSerializer(),
	}

	armNames := []string{"arm1", "arm2", "arm3"}
	initialParams := bandit.DefaultArmParams()
	initialParams.Version = 1

	for _, armName := range armNames[:2] {
		memoryStorage.Save("rule1", armName, initialParams)
	}

	rand.Seed(uint64(time.Now().UnixNano()))
	history := make(map[string]*ArmHistory)

	rememberProbs, rememberCount := float64(0), uint64(0)

	for i := 0; i < 20_000; i++ {
		switch i {
		case 5_000:
			memoryStorage.Save("rule1", armNames[2], initialParams)
		case 10_000:
			probs := gb.CalculateProbabilities(memoryStorage.GetAll("rule1"))
			rememberProbs, rememberCount = probs[armNames[2]].Score, probs[armNames[2]].Count

			memoryStorage.Delete("rule1", armNames[2])
		}

		currentArms := memoryStorage.GetAll("rule1")

		selectedArmID := bandit.SelectByProbabilities(gb.CalculateProbabilities(currentArms), bandit.DefaultExplorationFactor)

		var reward float64
		switch selectedArmID {
		case "arm1":
			reward = rand.NormFloat64()*math.Sqrt(2.0) + 5.0
		case "arm2":
			reward = rand.NormFloat64()*math.Sqrt(2.0) + 3.0
		case "arm3":
			reward = rand.NormFloat64()*math.Sqrt(2.0) + 8.0
		}

		oldParams := currentArms[selectedArmID]
		oldParams.Version = gb.Version

		if i >= 10_000 && i <= 15_000 {
			oldParams.Version--
		}

		newParams := gb.Calculate(oldParams, reward)

		memoryStorage.Save("rule1", selectedArmID, newParams)

		for armID, params := range currentArms {
			if _, exists := history[armID]; !exists {
				history[armID] = &ArmHistory{}
			}

			history[armID].Mu = append(history[armID].Mu, params.Mu)
			history[armID].SigmaSq = append(history[armID].SigmaSq, params.SigmaSq)
			history[armID].Count = append(history[armID].Count, params.Count)
			history[armID].Steps = append(history[armID].Steps, i)
		}

		probs := gb.CalculateProbabilities(currentArms)
		for armID := range currentArms {
			history[armID].Probabilities = append(history[armID].Probabilities, probs[armID].Score)
		}
	}

	visualizeResults(history)

	finalArms := memoryStorage.GetAll("rule1")
	probs := gb.CalculateProbabilities(finalArms)

	for armID, prob := range probs {
		fmt.Printf("Arm: %s | Probability: %.4f | Count: %d\n",
			armID, prob.Score, finalArms[armID].Count)
	}
	fmt.Printf("Arm: %s | Probability: %.4f | Count: %d\n", armNames[2], rememberProbs, rememberCount)

	fmt.Println("gb version:", gb.Version)
}

type MemoryStorage struct {
	data          map[string]map[string][]byte
	armSerializer bandit.ArmSerializer
	mu            sync.RWMutex
}

func (m *MemoryStorage) Save(ruleID, armID string, params *bandit.ArmParams) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[ruleID]; !exists {
		m.data[ruleID] = make(map[string][]byte)
	}

	var err error
	m.data[ruleID][armID], err = m.armSerializer.Serialize(params)
	if err != nil {
		panic(err)
	}
}

func (m *MemoryStorage) GetAll(ruleID string) map[string]*bandit.ArmParams {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var err error
	arms := make(map[string]*bandit.ArmParams)
	if ruleArms, exists := m.data[ruleID]; exists {
		for id, params := range ruleArms {
			arms[id], err = m.armSerializer.Deserialize(params)
			if err != nil {
				panic(err)
			}
		}
	}
	return arms
}

func (m *MemoryStorage) Delete(ruleID, armID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ruleArms, exists := m.data[ruleID]; exists {
		delete(ruleArms, armID)
	}
}

func visualizeResults(history map[string]*ArmHistory) {
	err := plotHistory("mu.png", "μ с течением времени", "Итерация", "μ", history, func(h *ArmHistory, i int) float64 { return h.Mu[i] })
	if err != nil {
		panic(err)
	}
	err = plotHistory("sigma_sq.png", "σ^2 с течением времени", "Итерация", "σ^2", history, func(h *ArmHistory, i int) float64 { return h.SigmaSq[i] })
	if err != nil {
		panic(err)
	}
	err = plotHistory("count.png", "Количество выборов ручки со временем", "Итерация", "Количество", history, func(h *ArmHistory, i int) float64 { return float64(h.Count[i]) })
	if err != nil {
		panic(err)
	}

	err = plotHistory("probabilities.png", "Вероятность выбора с течением времени", "Итерация", "Вероятность выбора", history, func(h *ArmHistory, i int) float64 { return h.Probabilities[i] })
	if err != nil {
		panic(err)
	}
}

func plotHistory(filename, title, xLabel, yLabel string, history map[string]*ArmHistory, extract func(*ArmHistory, int) float64) error {
	p := plot.New()
	p.Title.Text = title
	p.X.Label.Text = xLabel
	p.Y.Label.Text = yLabel

	colors := []color.Color{
		color.RGBA{R: 0, G: 255, B: 0, A: 255},
		color.RGBA{R: 0, G: 0, B: 255, A: 255},
		color.RGBA{R: 255, G: 0, B: 0, A: 255},
	}

	armIDs := make([]string, 0, len(history))
	for armID := range history {
		armIDs = append(armIDs, armID)
	}
	sort.Strings(armIDs)

	for idx, armID := range armIDs {
		h := history[armID]

		if len(h.Steps) == 0 {
			continue
		}
		pts := make(plotter.XYs, len(h.Steps))
		for step, val := range h.Steps {
			pts[step].X = float64(val)
			pts[step].Y = extract(h, step)
		}

		line, err := plotter.NewLine(pts)
		if err != nil {
			return err
		}

		line.Color = colors[idx%len(colors)]

		p.Add(line)
		p.Legend.Add(armID, line)
	}

	if err := p.Save(10*vg.Inch, 6*vg.Inch, filename); err != nil {
		return err
	}

	return nil
}
