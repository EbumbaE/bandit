package ruler

import (
	"sync"

	"github.com/google/uuid"
)

type Bandit interface {
	AddArm(armID string)
	RemoveArm(armID string)
	SelectArm() string
	UpdateArm(armID string, reward float64)
}

type Ruler struct {
	storage *Storage

	variants map[string]Variants
	arms     map[string]Bandit
	mtx      sync.Mutex
}

func NewRuler(s *Storage) *Ruler {
	return &Ruler{
		storage: s,
		rules:   make(map[string]Variants),
		arms:    make(map[string]Variant),
	}
}

func (r *Ruler) AddRule(ruleID string, bandit Bandit) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if _, ok := r.rules[ruleID]; !ok {
		r.rules[ruleID] = bandit
	}
}

func (r *Ruler) AddVariants(ruleID string, variants []Variant) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	for _, v := range variants {
		armID := uuid.New()

		r.rules[ruleID].AddArm(armID)
	}
}

// админка:

// id правила, оно публичное
// при создании правила возвращается id
// по нему можно получить лучшую руку
// по id правила, можно добавить какое-то data
