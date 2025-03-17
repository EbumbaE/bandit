package main

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

// Arm (вариант) представляет один из вариантов правила.
type Arm struct {
	Alpha float64 // Количество успехов
	Beta  float64 // Количество неудач
	Mutex sync.Mutex
}

// Bandit реализует многорукий бандит с Thompson Sampling.
type Bandit struct {
	Arms  map[string]*Arm // Key: armID (идентификатор варианта)
	Mutex sync.RWMutex
}

func NewBandit() *Bandit {
	return &Bandit{
		Arms: make(map[string]*Arm),
	}
}

// AddArm добавляет новый вариант (с начальными параметрами Alpha=1, Beta=1).
func (b *Bandit) AddArm(armID string) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	b.Arms[armID] = &Arm{Alpha: 1, Beta: 1}
}

// SelectArm выбирает вариант с максимальным сэмплированным значением.
func (b *Bandit) SelectArm() (string, float64) {
	b.Mutex.RLock()
	defer b.Mutex.RUnlock()

	var selectedArmID string
	maxSample := -1.0

	for armID, arm := range b.Arms {
		arm.Mutex.Lock()
		// Сэмплируем значение из Beta-распределения
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

// UpdateArm обновляет параметры Alpha/Beta для выбранного варианта.
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
		arm.Alpha += reward // Для непрерывных наград (например, доход)
	} else {
		arm.Beta += 1 // Для бинарных наград (0/1)
	}
}

// GetScores возвращает текущие score всех вариантов.
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

// SelectArmRandom выбирает вариант случайно, с вероятностью, пропорциональной его score.
func (b *Bandit) SelectArmRandom() (string, float64) {
	b.Mutex.RLock()
	defer b.Mutex.RUnlock()

	// Собираем все scores
	scores := make(map[string]float64)
	totalScore := 0.0
	for armID, arm := range b.Arms {
		arm.Mutex.Lock()
		score := arm.Alpha / (arm.Alpha + arm.Beta)
		arm.Mutex.Unlock()
		scores[armID] = score
		totalScore += score
	}

	// Нормализуем scores (чтобы их сумма была равна 1)
	normalizedScores := make(map[string]float64)
	for armID, score := range scores {
		normalizedScores[armID] = score / totalScore
	}

	// Взвешенный случайный выбор
	randomValue := rand.Float64()
	cumulativeProbability := 0.0

	for armID, prob := range normalizedScores {
		cumulativeProbability += prob
		if randomValue <= cumulativeProbability {
			return armID, scores[armID]
		}
	}

	// Если что-то пошло не так, возвращаем первый вариант
	for armID := range scores {
		return armID, scores[armID]
	}

	return "", 0.0
}

func main() {
	// Инициализация Bandit
	bandit := NewBandit()
	bandit.AddArm("arm1")
	bandit.AddArm("arm2")

	// Инициализация генератора случайных чисел
	rand.Seed(uint64(time.Now().UnixNano()))

	// Симуляция 1000 итераций
	for i := 0; i < 1000; i++ {
		// Выбор варианта (случайно с учетом score)
		selectedArmID, score := bandit.SelectArmRandom()
		fmt.Printf("Selected Arm: %s (Score: %.2f)\n", selectedArmID, score)

		// Симуляция награды (например, клик или покупка)
		reward := 0.0
		if selectedArmID == "arm1" {
			reward = 0.7 // arm1 имеет более высокую вероятность успеха
		} else {
			reward = 0.3
		}

		// Обновление параметров
		bandit.UpdateArm(selectedArmID, reward)
	}

	// Получение итоговых score
	scores := bandit.GetScores()
	fmt.Println("Final Scores:", scores)
	// Пример вывода: Final Scores: map[arm1:0.85 arm2:0.45]
}
