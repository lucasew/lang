package pt

import (
	"strings"
	"sync"
)

// BrazilianToponymMap ports org.languagetool.rules.pt.BrazilianToponymMap.
// Construction uses BrazilianToponymMapLoader.buildMap() (Java ctor).
// Full Map method fidelity is a separate checklist line.
type BrazilianToponymMap struct {
	// state → normalised municipality names (loader: hyphen→space, lower)
	byState map[string][]string
	// any valid municipality name (for quick membership; derived from byState)
	all map[string]struct{}
}

var (
	brToponymOnce sync.Once
	brToponymMap  *BrazilianToponymMap
)

// LoadBrazilianToponymMap loads once via BrazilianToponymMapLoader.BuildMap
// (Java: map = new BrazilianToponymMapLoader().buildMap()).
func LoadBrazilianToponymMap() *BrazilianToponymMap {
	brToponymOnce.Do(func() {
		raw := NewBrazilianToponymMapLoader().BuildMap()
		m := &BrazilianToponymMap{
			byState: raw,
			all:     map[string]struct{}{},
		}
		for _, list := range raw {
			for _, n := range list {
				m.all[n] = struct{}{}
			}
		}
		brToponymMap = m
	})
	return brToponymMap
}

// IsValidToponym reports whether any suffix of the toponym matches a municipality.
func (m *BrazilianToponymMap) IsValidToponym(toponym string) bool {
	return m.toponymIter(toponym, func(check string) bool {
		_, ok := m.all[check]
		return ok
	})
}

// StatesWithMunicipality returns states that contain the exact normalised toponym.
func (m *BrazilianToponymMap) StatesWithMunicipality(toponym string) []string {
	norm := strings.ToLower(strings.ReplaceAll(toponym, "-", " "))
	var out []string
	for state, list := range m.byState {
		for _, n := range list {
			if n == norm {
				out = append(out, state)
				break
			}
		}
	}
	return out
}

// IsToponymInState checks membership in a specific state list.
func (m *BrazilianToponymMap) IsToponymInState(toponym, state string) bool {
	norm := strings.ToLower(strings.ReplaceAll(toponym, "-", " "))
	list := m.byState[state]
	for _, n := range list {
		if n == norm {
			return true
		}
	}
	return false
}

func (m *BrazilianToponymMap) toponymIter(toponym string, ok func(string) bool) bool {
	// Java: toponym.replace('-', ' ').toLowerCase().split(" ")
	norm := strings.ToLower(strings.ReplaceAll(toponym, "-", " "))
	parts := strings.Split(norm, " ")
	for i := 0; i < len(parts); i++ {
		check := strings.Join(parts[i:], " ")
		if ok(check) {
			return true
		}
	}
	return false
}
