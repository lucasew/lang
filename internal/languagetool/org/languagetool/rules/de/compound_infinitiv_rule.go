package de

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// CompoundInfinitivRule is a surface stand-in for CompoundInfinitivRule.
// Without ZUS/VER:INF tags and a speller, particles that commonly form
// erweiterte Infinitive are matched before "zu" + infinitive-looking token.
// Separable prefixes (an/auf/vor/…) are excluded to limit false positives.
type CompoundInfinitivRule struct {
	Messages  map[string]string
	particles map[string]struct{}
}

func NewCompoundInfinitivRule(messages map[string]string) *CompoundInfinitivRule {
	list := []string{"vorbei", "sauber", "weiterhin"}
	// "sicher" often compounds (sicherzugehen) but has anti-patterns; include carefully
	list = append(list, "sicher")
	m := map[string]struct{}{}
	for _, w := range list {
		m[w] = struct{}{}
	}
	return &CompoundInfinitivRule{Messages: messages, particles: m}
}

func (r *CompoundInfinitivRule) GetID() string { return "COMPOUND_INFINITIV_RULE" }

func looksLikeInfinitiveDE(w string) bool {
	if len(w) < 4 {
		return false
	}
	r0 := []rune(w)[0]
	if unicode.IsUpper(r0) {
		return false
	}
	lc := strings.ToLower(w)
	return strings.HasSuffix(lc, "en") || strings.HasSuffix(lc, "eln") || strings.HasSuffix(lc, "ern")
}

func (r *CompoundInfinitivRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 2; i < len(tokens)-1; i++ {
		if tokens[i].GetToken() != "zu" {
			continue
		}
		prev := tokens[i-1].GetToken()
		next := tokens[i+1].GetToken()
		lcPrev := strings.ToLower(prev)
		if _, ok := r.particles[lcPrev]; !ok {
			continue
		}
		if !looksLikeInfinitiveDE(next) {
			continue
		}
		// anti: "auf Nummer sicher zu …"
		if i >= 3 && strings.EqualFold(tokens[i-2].GetToken(), "Nummer") && lcPrev == "sicher" {
			continue
		}
		// anti: "ganz schön zu tun"
		if i >= 3 && strings.EqualFold(tokens[i-2].GetToken(), "schön") {
			continue
		}
		msg := "Wenn der erweiterte Infinitiv von dem Verb '" + prev + next +
			"' abgeleitet ist, sollte er zusammengeschrieben werden."
		rm := rules.NewRuleMatch(r, sentence, tokens[i-1].GetStartPos(), tokens[i+1].GetEndPos(), msg)
		rm.ShortMessage = "erweiterter Infinitiv"
		rm.SetSuggestedReplacement(prev + "zu" + next)
		matches = append(matches, rm)
	}
	return matches
}
