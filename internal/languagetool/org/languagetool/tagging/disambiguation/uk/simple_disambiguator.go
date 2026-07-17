package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// particle suffix pattern for -то/-бо etc. (Java SimpleDisambiguator.PATTERN)
var reParticleSuffix = regexp.MustCompile(`.*-(то|от|таки|бо|но)$`)

// MatcherEntry is one lemma+pos pair to remove.
type MatcherEntry struct {
	Lemma string
	POS   string // substring match on POS tag
}

// TokenMatcher holds entries to strip from a token's readings.
type TokenMatcher struct {
	Entries []MatcherEntry
}

func (m *TokenMatcher) Matches(tok *languagetool.AnalyzedToken) bool {
	if m == nil || tok == nil {
		return false
	}
	lemma := ""
	if tok.GetLemma() != nil {
		lemma = *tok.GetLemma()
	}
	pos := ""
	if tok.GetPOSTag() != nil {
		pos = *tok.GetPOSTag()
	}
	for _, e := range m.Entries {
		if e.Lemma != "" && !strings.EqualFold(lemma, e.Lemma) {
			continue
		}
		if e.POS != "" && !strings.Contains(pos, e.POS) {
			continue
		}
		return true
	}
	return false
}

// SimpleDisambiguator ports tagging.disambiguation.uk.SimpleDisambiguator.
// RemoveMap is inject-friendly (full disambig_remove.txt deferred).
type SimpleDisambiguator struct {
	disambiguation.AbstractDisambiguator
	RemoveMap map[string]*TokenMatcher
}

func NewSimpleDisambiguator() *SimpleDisambiguator {
	return &SimpleDisambiguator{RemoveMap: map[string]*TokenMatcher{}}
}

// NewSimpleDisambiguatorWith starts with an explicit remove map.
func NewSimpleDisambiguatorWith(m map[string]*TokenMatcher) *SimpleDisambiguator {
	if m == nil {
		m = map[string]*TokenMatcher{}
	}
	return &SimpleDisambiguator{RemoveMap: m}
}

func (d *SimpleDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	RemoveRareForms(input, d.RemoveMap)
	return input
}

// RemoveRareForms strips readings matching RemoveMap (in-place).
func RemoveRareForms(input *languagetool.AnalyzedSentence, removeMap map[string]*TokenMatcher) {
	if input == nil || len(removeMap) == 0 {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		if tokens[i] == nil {
			continue
		}
		token := tokens[i].GetToken()
		if token == "" {
			continue
		}
		tm := lookupMatcher(token, removeMap)
		if tm == nil {
			continue
		}
		// copy readings to avoid mutation during iteration issues
		readings := append([]*languagetool.AnalyzedToken(nil), tokens[i].GetReadings()...)
		for j := len(readings) - 1; j >= 0; j-- {
			if tm.Matches(readings[j]) {
				tokens[i].RemoveReading(readings[j], "disambig_remove")
			}
		}
	}
}

func lookupMatcher(token string, removeMap map[string]*TokenMatcher) *TokenMatcher {
	if tm := removeMap[token]; tm != nil {
		return tm
	}
	low := strings.ToLower(token)
	if tm := removeMap[low]; tm != nil {
		return tm
	}
	if reParticleSuffix.MatchString(low) {
		if idx := strings.LastIndex(low, "-"); idx > 0 {
			if tm := removeMap[low[:idx]]; tm != nil {
				return tm
			}
		}
	}
	return nil
}

// RemoveVmisReadings drops v_mis when another non-end reading remains (soft green).
func RemoveVmisReadings(atr *languagetool.AnalyzedTokenReadings) {
	if atr == nil || !canRemoveVmis(atr.GetReadings()) {
		return
	}
	readings := append([]*languagetool.AnalyzedToken(nil), atr.GetReadings()...)
	for _, r := range readings {
		if r != nil && r.GetPOSTag() != nil && strings.Contains(*r.GetPOSTag(), "v_mis") {
			atr.RemoveReading(r, "dis_v_mis")
		}
	}
}

func canRemoveVmis(analyzed []*languagetool.AnalyzedToken) bool {
	foundVmis, foundOther := false, false
	for _, token := range analyzed {
		if token == nil {
			continue
		}
		pos := token.GetPOSTag()
		if pos != nil && strings.Contains(*pos, "v_mis") {
			foundVmis = true
		} else if pos != nil && !strings.HasSuffix(*pos, "_END") {
			foundOther = true
		}
		if foundVmis && foundOther {
			return true
		}
	}
	return false
}

var _ disambiguation.Disambiguator = (*SimpleDisambiguator)(nil)
