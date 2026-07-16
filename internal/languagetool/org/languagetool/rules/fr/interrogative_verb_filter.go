package fr

import "strings"

// InterrogativeVerbFilter ports desired-POS selection for French interrogative/imperative verbs.
type InterrogativeVerbFilter struct{}

func NewInterrogativeVerbFilter() *InterrogativeVerbFilter {
	return &InterrogativeVerbFilter{}
}

// DesiredPostagForPronoun maps a subject/object clitic to a French verb POS regex.
// pronoun is the surface form (je, tu, il, nous, vous, ils, …).
func (f *InterrogativeVerbFilter) DesiredPostagForPronoun(pronoun string) string {
	switch strings.ToLower(pronoun) {
	case "tu":
		// imp 2s/2p or ind/cond 2p (vous forms handled separately)
		return "V.* (imp) [23] [sp]|V .*(ind|cond).* 2 p"
	case "vous":
		return "V.* (imp) .*|V .*(ind|cond).* 1 p"
	case "nous":
		return "V.* (imp) .*"
	case "je", "j":
		return "V .*(ind|cond).* 1 s"
	case "il", "elle", "on":
		return "V .*(ind|cond).* 3 s"
	case "ils", "elles":
		return "V .*(ind|cond).* 3 p"
	case "toi":
		// imperative reflexive
		return "V .*(ind|cond).* 2 s"
	default:
		// common plurals
		if strings.EqualFold(pronoun, "nous") {
			return "V .*(ind|cond).* 1 p"
		}
		return ""
	}
}

// FilterByDesiredPOS keeps candidates whose POS matches desiredPostag regex.
// MatchesPOS is pluggable; when nil all candidates are kept.
func (f *InterrogativeVerbFilter) FilterByDesiredPOS(candidates []string, desiredPostag string, matchesPOS func(form, postagRE string) bool) []string {
	if matchesPOS == nil || desiredPostag == "" {
		return candidates
	}
	var out []string
	for _, c := range candidates {
		if matchesPOS(c, desiredPostag) {
			out = append(out, c)
		}
	}
	return out
}
