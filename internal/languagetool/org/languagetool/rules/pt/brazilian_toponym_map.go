package pt

import (
	"strings"
	"sync"
)

// BrazilianToponymMap ports org.languagetool.rules.pt.BrazilianToponymMap.
type BrazilianToponymMap struct {
	// Java: private final Map<String, List<String>> map;
	// state → municipality names as stored by the loader (hyphen→space, lower).
	map_ map[string][]string
	// all is a derived index of every municipality string across map_.values()
	// for O(1) anyMatch(list.contains) in isValidToponym — same membership.
	all map[string]struct{}
}

var (
	brToponymOnce sync.Once
	brToponymMap  *BrazilianToponymMap
)

// NewBrazilianToponymMap ports package-private BrazilianToponymMap():
// map = new BrazilianToponymMapLoader().buildMap()
func NewBrazilianToponymMap() *BrazilianToponymMap {
	return newBrazilianToponymMap(NewBrazilianToponymMapLoader().BuildMap())
}

// LoadBrazilianToponymMap returns a process-wide singleton (filter ctor uses one map).
// Construction path matches Java ctor via BrazilianToponymMapLoader.BuildMap.
func LoadBrazilianToponymMap() *BrazilianToponymMap {
	brToponymOnce.Do(func() {
		brToponymMap = NewBrazilianToponymMap()
	})
	return brToponymMap
}

func newBrazilianToponymMap(raw map[string][]string) *BrazilianToponymMap {
	m := &BrazilianToponymMap{
		map_: raw,
		all:  map[string]struct{}{},
	}
	for _, list := range raw {
		for _, n := range list {
			m.all[n] = struct{}{}
		}
	}
	return m
}

// toponymIter ports private <T> T toponymIter(toponym, processor, defaultValue).
// Hungry left-strip: normalise, split on space, join parts[i:], apply processor.
// Only isValidToponym uses this; processor "null" is modeled as ok==false.
func (m *BrazilianToponymMap) toponymIter(toponym string, ok func(string) bool) bool {
	// Java: toponym.replace('-', ' ').toLowerCase()
	normalised := strings.ToLower(strings.ReplaceAll(toponym, "-", " "))
	// Java: normalisedToponym.split(" ") — limit 0, trailing empties discarded
	parts := javaStringSplitSpace(normalised)
	for i := 0; i < len(parts); i++ {
		// Java: String.join(" ", Arrays.copyOfRange(toponymParts, i, toponymLength))
		check := strings.Join(parts[i:], " ")
		if ok(check) {
			return true
		}
	}
	return false
}

// javaStringSplitSpace mirrors Java String.split(" ") (limit 0).
// Trailing empty strings discarded; empty input yields [""] (no match → {input}).
func javaStringSplitSpace(s string) []string {
	if s == "" {
		return []string{""}
	}
	parts := strings.Split(s, " ")
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// IsValidToponym ports public boolean isValidToponym(String toponym).
// Java: toponymIter(..., check -> map.values().stream().anyMatch(list -> list.contains(check)) ? true : null, false)
func (m *BrazilianToponymMap) IsValidToponym(toponym string) bool {
	if m == nil {
		return false
	}
	return m.toponymIter(toponym, func(check string) bool {
		_, found := m.all[check]
		return found
	})
}

// StatesWithMunicipality ports public List<String> getStatesWithMunicipality(String toponym).
// Java uses municipalities.contains(toponym) with the argument AS-IS — no normalise.
func (m *BrazilianToponymMap) StatesWithMunicipality(toponym string) []string {
	if m == nil {
		return nil
	}
	var states []string
	for state, municipalities := range m.map_ {
		if listContains(municipalities, toponym) {
			states = append(states, state)
		}
	}
	return states
}

// IsToponymInState ports public boolean isToponymInState(String toponym, String state).
// Java: municipalities = map.get(state); return municipalities != null && municipalities.contains(toponym)
// Argument AS-IS — no normalise.
func (m *BrazilianToponymMap) IsToponymInState(toponym, state string) bool {
	if m == nil {
		return false
	}
	municipalities, ok := m.map_[state]
	return ok && listContains(municipalities, toponym)
}

// listContains ports List.contains (reference equality not needed; strings use equals).
func listContains(list []string, s string) bool {
	for _, n := range list {
		if n == s {
			return true
		}
	}
	return false
}
