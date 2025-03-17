package main

import (
	"fmt"
	"image/color"
	"math"
	"sort"
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	bandit "github.com/EbumbaE/bandit/services/bandit-core/v2"
)

type ArmHistory struct {
	Mu            []float64
	SigmaSq       []float64
	Count         []int
	Probabilities []float64
	Steps         []int
}

func main() {
	arms := []string{"arm1", "arm2", "arm3"}

	bandit := bandit.NewGaussianBandit()
	for i := range 2 {
		bandit.AddArm(arms[i])
	}

	rand.Seed(uint64(time.Now().UnixNano()))

	history := make(map[string]*ArmHistory)
	bandit.Mutex.RLock()
	for _, armID := range arms {
		history[armID] = &ArmHistory{}
	}
	bandit.Mutex.RUnlock()

	for i := 0; i < 20_000; i++ {
		if i == 5000 {
			bandit.AddArm(arms[2])
		}
		if i == 10000 {
			bandit.RemoveArm(arms[2])
		}

		selectedArmID := bandit.SelectArm()

		var reward float64
		switch selectedArmID {
		case "arm1":
			reward = rand.NormFloat64()*math.Sqrt(2.0) + 5.0
		case "arm2":
			reward = rand.NormFloat64()*math.Sqrt(2.0) + 3.0
		case "arm3":
			reward = rand.NormFloat64()*math.Sqrt(2.0) + 8.0
		}

		bandit.UpdateArm(selectedArmID, reward)

		bandit.Mutex.RLock()
		currentArms := make([]string, 0, len(bandit.Arms))
		for armID := range bandit.Arms {
			currentArms = append(currentArms, armID)
		}
		sort.Strings(currentArms)
		for _, armID := range currentArms {
			arm := bandit.Arms[armID]
			arm.Mutex.Lock()
			history[armID].Mu = append(history[armID].Mu, arm.Mu)
			history[armID].SigmaSq = append(history[armID].SigmaSq, arm.SigmaSq)
			history[armID].Count = append(history[armID].Count, arm.N)
			history[armID].Steps = append(history[armID].Steps, i)
			arm.Mutex.Unlock()
		}
		probs := bandit.GetSelectionProbabilities()
		for _, armID := range currentArms {
			history[armID].Probabilities = append(history[armID].Probabilities, probs[armID])
		}
		bandit.Mutex.RUnlock()
	}

	err := plotHistory("mu.png", "Mu over Time", "Step", "Mu", history, func(h *ArmHistory, i int) float64 { return h.Mu[i] })
	if err != nil {
		panic(err)
	}
	err = plotHistory("sigma_sq.png", "Sigma Squared over Time", "Step", "Sigma Squared", history, func(h *ArmHistory, i int) float64 { return h.SigmaSq[i] })
	if err != nil {
		panic(err)
	}
	err = plotHistory("count.png", "Count over Time", "Step", "Count", history, func(h *ArmHistory, i int) float64 { return float64(h.Count[i]) })
	if err != nil {
		panic(err)
	}

	err = plotHistory("probabilities.png", "Selection Probability over Time", "Step", "Probability", history, func(h *ArmHistory, i int) float64 { return h.Probabilities[i] })
	if err != nil {
		panic(err)
	}

	probabilities := bandit.GetSelectionProbabilities()
	for armID, prob := range probabilities {
		fmt.Printf("Arm: %s | Probability: %.4f | Count: %d\n", armID, prob, bandit.Arms[armID].N)
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
	}
	colorIndex := 0

	armIDs := make([]string, 0, len(history))
	for armID := range history {
		armIDs = append(armIDs, armID)
	}
	sort.Strings(armIDs)

	for _, armID := range armIDs {
		h := history[armID]
		if len(h.Mu) == 0 {
			continue
		}
		pts := make(plotter.XYs, len(h.Mu))
		for i := range h.Mu {
			pts[i].X = float64(h.Steps[i])
			pts[i].Y = extract(h, i)
		}
		line, err := plotter.NewLine(pts)
		if err != nil {
			return err
		}
		line.Color = colors[colorIndex%len(colors)]
		colorIndex++
		p.Add(line)
		p.Legend.Add(armID, line)
	}

	if err := p.Save(6*vg.Inch, 4*vg.Inch, filename); err != nil {
		return err
	}
	return nil
}
