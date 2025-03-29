package ruler

import (
	"sync"

	"github.com/google/uuid"
)

type Storage struct {
	mtx sync.RWMutex

	rules          map[string]Rule
	variantsByRule map[string][]Variant
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) GetRuleByID(id string) Rule {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.rules[id]
}

func (s *Storage) AddRule(r Rule) string {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	id := uuid.New().String()

	r.ID = id
	s.rules[id] = r

	return id
}

func (s *Storage) GetVariantsByRuleID(ruleID string) []Variant {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.variantsByRule[ruleID]
}

func (s *Storage) AddVariant(v Variant) string {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	id := uuid.New().String()

	v.ID = id
	s.variantsByRule[v.RuleID] = append(s.variantsByRule[v.RuleID], v)

	return id
}
