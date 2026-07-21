package pt

import (
	"bufio"
	"embed"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"strings"
	"sync"
)

//go:embed data/brazilian_municipalities/*.tsv
var municipalitiesFS embed.FS

var brStates = []string{
	"AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO", "MA",
	"MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI", "RJ", "RN",
	"RS", "RO", "RR", "SC", "SP", "SE", "TO",
}

// BrazilianToponymMap ports org.languagetool.rules.pt.BrazilianToponymMap.
type BrazilianToponymMap struct {
	// state → lowercased municipality names (hyphen→space)
	byState map[string][]string
	// any valid municipality name (for quick membership)
	all map[string]struct{}
}

var (
	brToponymOnce sync.Once
	brToponymMap  *BrazilianToponymMap
)

// LoadBrazilianToponymMap loads embedded state TSV lists once.
func LoadBrazilianToponymMap() *BrazilianToponymMap {
	brToponymOnce.Do(func() {
		m := &BrazilianToponymMap{
			byState: map[string][]string{},
			all:     map[string]struct{}{},
		}
		for _, state := range brStates {
			path := "data/brazilian_municipalities/" + state + ".tsv"
			f, err := municipalitiesFS.Open(path)
			if err != nil {
				continue
			}
			var list []string
			sc := bufio.NewScanner(f)
			for sc.Scan() {
				line := tools.JavaStringTrim(sc.Text())
				if line == "" {
					continue
				}
				norm := strings.ToLower(strings.ReplaceAll(line, "-", " "))
				list = append(list, norm)
				m.all[norm] = struct{}{}
			}
			_ = f.Close()
			m.byState[state] = list
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
