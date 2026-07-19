package de

import "strings"

// Past-tense and participle suggestion paths from
// GermanSpellerRule.getPastTenseVerbSuggestion / getParticipleSuggestion.
// Require Synthesize + TagPOS/LemmaOf; fail-closed when hooks missing.

// baseForThirdPersonSingularVerb ports baseForThirdPersonSingularVerb:
// tag word; if VER:3:SIN… return lemma via LemmaOf.
func (r *GermanSpellerRule) baseForThirdPersonSingularVerb(word string) string {
	if r == nil || word == "" || r.TagPOS == nil {
		return ""
	}
	has := false
	for _, t := range r.TagPOS(word) {
		if strings.HasPrefix(t, "VER:3:SIN") {
			has = true
			break
		}
	}
	if !has {
		return ""
	}
	if r.LemmaOf == nil {
		return ""
	}
	return r.LemmaOf(word)
}

// pastTenseVerbSuggestion ports getPastTenseVerbSuggestion:
// words ending in "e" (e.g. greifte) → stem → lemma → synth VER:3:SIN:PRT:.*
func (r *GermanSpellerRule) pastTenseVerbSuggestion(word string) []string {
	if r == nil || !strings.HasSuffix(word, "e") || r.Synthesize == nil {
		return nil
	}
	if len(word) < 2 {
		return nil
	}
	stem := word[:len(word)-1]
	lemma := r.baseForThirdPersonSingularVerb(stem)
	if lemma == "" {
		return nil
	}
	forms := r.Synthesize(lemma, `VER:3:SIN:PRT:.*`)
	if len(forms) == 0 {
		return nil
	}
	return []string{forms[0]}
}

// participleSuggestion ports getParticipleSuggestion:
// ge…t (e.g. geschwimmt) → strip ge + t→en base → synth VER:PA2:.* if dict accepts.
func (r *GermanSpellerRule) participleSuggestion(word string) []string {
	if r == nil || r.Synthesize == nil {
		return nil
	}
	if !strings.HasPrefix(word, "ge") || !strings.HasSuffix(word, "t") {
		return nil
	}
	if len(word) < 4 {
		return nil
	}
	baseform := word[2:len(word)-1] + "en"
	forms := r.Synthesize(baseform, `VER:PA2:.*`)
	if len(forms) == 0 {
		return nil
	}
	if !dictAccepts(forms[0]) {
		return nil
	}
	return []string{forms[0]}
}
