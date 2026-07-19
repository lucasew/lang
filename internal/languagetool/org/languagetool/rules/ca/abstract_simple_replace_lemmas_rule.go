package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// SynthesizeFunc synthesizes surface forms for lemma+postag (pluggable synthesizer).
// Ports CatalanSynthesizer.synthesize(AnalyzedToken, postag). Empty when synthesis fails.
type SynthesizeFunc func(lemma, postag string) []string

// AbstractSimpleReplaceLemmasRule ports org.languagetool.rules.ca.AbstractSimpleReplaceLemmasRule.
// Matches wrong lemmas (from tagger readings) and suggests synthesized replacements.
// Without lemma readings, fail closed (no surface invent of forms).
type AbstractSimpleReplaceLemmasRule struct {
	ID          string
	Description string
	ShortMsg    string
	// MessageFn optional; default "Possible lemma replacement for {lemma}".
	MessageFn func(tokenStr string, replacements []string) string
	// WrongLemmas maps wrong lemma → replacement lemmas.
	WrongLemmas map[string][]string
	// Synthesize optional; when nil/empty, bare replacement lemmas are suggested (Java).
	Synthesize SynthesizeFunc
}

func (r *AbstractSimpleReplaceLemmasRule) GetID() string {
	if r != nil && r.ID != "" {
		return r.ID
	}
	return "CA_LEMMA_REPLACE"
}

func (r *AbstractSimpleReplaceLemmasRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	return ""
}

func (r *AbstractSimpleReplaceLemmasRule) GetWrongWords() map[string][]string {
	if r == nil {
		return nil
	}
	return r.WrongLemmas
}

// Match ports AbstractSimpleReplaceLemmasRule.match.
func (r *AbstractSimpleReplaceLemmasRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil || r.WrongLemmas == nil {
		return nil
	}
	var out []*rules.RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil || tok.IsImmunized() {
			continue
		}
		var replacementLemmas []string
		var replacePOSTag string
		var originalLemma string
		matched := false
		for _, at := range tok.GetReadings() {
			if at == nil {
				continue
			}
			lem := at.GetLemma()
			if lem == nil || *lem == "" {
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
		// find suggestions (Java: only when replacePOSTag != null)
		var possible []string
		if replacementLemmas != nil && replacePOSTag != "" {
			for _, rl := range replacementLemmas {
				synthesized := r.synth(rl, replacePOSTag)
				if len(synthesized) == 0 {
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
		msg := "Possible lemma replacement for " + originalLemma
		if r.MessageFn != nil {
			msg = r.MessageFn(tok.GetToken(), possible)
		}
		m := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
		m.ShortMessage = r.ShortMsg
		if len(possible) > 0 {
			m.SetSuggestedReplacements(possible)
		}
		out = append(out, m)
	}
	return out
}

func (r *AbstractSimpleReplaceLemmasRule) synth(lemma, postag string) []string {
	if r == nil || r.Synthesize == nil {
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
