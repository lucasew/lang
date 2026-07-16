package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SynthesizeFunc synthesizes surface forms for lemma+postag (pluggable synthesizer).
// Returns empty when synthesis fails; caller may fall back to bare lemma.
type SynthesizeFunc func(lemma, postag string) []string

// AbstractSimpleReplaceLemmasRule ports org.languagetool.rules.ca.AbstractSimpleReplaceLemmasRule.
// Matches wrong lemmas and suggests synthesized replacements.
type AbstractSimpleReplaceLemmasRule struct {
	ID          string
	Description string
	// WrongLemmas maps wrong lemma → replacement lemmas.
	WrongLemmas map[string][]string
	// Synthesize optional; when nil, bare replacement lemmas are suggested.
	Synthesize SynthesizeFunc
}

func (r *AbstractSimpleReplaceLemmasRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "CA_LEMMA_REPLACE"
}

func (r *AbstractSimpleReplaceLemmasRule) GetWrongWords() map[string][]string {
	return r.WrongLemmas
}

func (r *AbstractSimpleReplaceLemmasRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil || r.WrongLemmas == nil {
		return nil
	}
	var out []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		var replacementLemmas []string
		var replacePOSTag string
		var originalLemma string
		matched := false
		for _, at := range tokens[i].GetReadings() {
			lem := at.GetLemma()
			if lem == nil {
				continue
			}
			if repl, ok := r.WrongLemmas[*lem]; ok {
				replacementLemmas = repl
				if at.GetPOSTag() != nil {
					replacePOSTag = *at.GetPOSTag()
				}
				originalLemma = *lem
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		var possible []string
		if replacementLemmas != nil && replacePOSTag != "" {
			for _, rl := range replacementLemmas {
				synthesized := r.synth(rl, replacePOSTag)
				if len(synthesized) == 0 {
					// try gender-neutral postag like Java
					tag2 := relaxGender(replacePOSTag)
					if tag2 != replacePOSTag {
						synthesized = r.synth(rl, tag2)
					}
				}
				if len(synthesized) == 0 && len(rl) > 1 {
					possible = append(possible, rl)
				} else {
					possible = append(possible, synthesized...)
				}
			}
		}
		m := rules.NewRuleMatch(r, sentence, tokens[i].GetStartPos(), tokens[i].GetEndPos(),
			"Possible lemma replacement for "+originalLemma)
		if len(possible) > 0 {
			m.SetSuggestedReplacements(possible)
		}
		out = append(out, m)
	}
	return out
}

func (r *AbstractSimpleReplaceLemmasRule) synth(lemma, postag string) []string {
	if r.Synthesize == nil {
		return nil
	}
	return r.Synthesize(lemma, postag)
}

func relaxGender(postag string) string {
	// Java: replaceAll("[MFC]S",".S").replaceAll("[MFC]P",".P")
	s := postag
	s = replaceAllRunes(s, []string{"MS", "FS", "CS"}, ".S")
	s = replaceAllRunes(s, []string{"MP", "FP", "CP"}, ".P")
	return s
}

func replaceAllRunes(s string, olds []string, new string) string {
	for _, o := range olds {
		s = strings.ReplaceAll(s, o, new)
	}
	return s
}
