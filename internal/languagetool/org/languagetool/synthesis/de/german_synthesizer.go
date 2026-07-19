package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	GermanSynthResource = "/de/german_synth.dict"
	GermanTagsFile      = "/de/german_tags.txt"
)

// germanSynthRemove ports GermanSynthesizer.REMOVE (bad/old forms still in synth dict).
var germanSynthRemove = map[string]struct{}{
	"unsren": {}, "unsrem": {}, "unsres": {}, "unsre": {}, "unsern": {}, "unserm": {}, "unsrer": {},
	// old spellings still in the synthesizer dict:
	"angepaßt": {}, "beschloß": {}, "biß": {}, "entschloß": {}, "ergoß": {}, "faßt": {}, "genoß": {},
	"paßt": {}, "paßte": {}, "preßt": {}, "preßte": {}, "riß": {},
	"schloß": {}, "streßtest": {}, "vergißt": {}, "verlaß": {}, "verläßt": {}, "vermiß": {}, "vermißt": {},
	"wißt": {}, "wußtest": {}, "wüßtest": {},
}

// GermanSynthesizer ports org.languagetool.synthesis.GermanSynthesizer.
type GermanSynthesizer struct {
	*synthesis.BaseSynthesizer
	// StrictCompoundTokenize ports GermanCompoundTokenizer.getStrictInstance().tokenize.
	// Nil → getCompoundForms fail-closed (no invent splits).
	StrictCompoundTokenize func(lemma string) []string
}

func NewGermanSynthesizer(manual *synthesis.ManualSynthesizer) *GermanSynthesizer {
	base := synthesis.NewBaseSynthesizer("de", manual)
	base.ResourceFileName = GermanSynthResource
	base.TagFileName = GermanTagsFile
	return &GermanSynthesizer{BaseSynthesizer: base}
}

// WithStrictCompoundTokenize sets the compound tokenizer (optional).
func (s *GermanSynthesizer) WithStrictCompoundTokenize(fn func(lemma string) []string) *GermanSynthesizer {
	if s != nil {
		s.StrictCompoundTokenize = fn
	}
	return s
}

// Synthesize implements synthesis.Synthesizer (Java: case in lookup, REMOVE, then compound fallback).
func (s *GermanSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.synthesize(token, posTag, false)
}

func (s *GermanSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.synthesize(token, posTag, re)
}

// SynthesizeForPosTags ports BaseSynthesizer.synthesizeForPosTags via German lookup case gate.
// Java SpellingData / LineExpander call this on GermanSynthesizer.INSTANCE (no REMOVE filter).
func (s *GermanSynthesizer) SynthesizeForPosTags(lemma string, acceptTag func(string) bool) []string {
	if s == nil || s.BaseSynthesizer == nil {
		return nil
	}
	forms := s.BaseSynthesizer.SynthesizeForPosTags(lemma, acceptTag)
	// Java: results.addAll(lookup(...)) with German case filter; removeExceptions only.
	return filterCaseMatch(lemma, forms)
}

func (s *GermanSynthesizer) synthesize(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	var forms []string
	var err error
	if posTagRegExp {
		forms, err = s.BaseSynthesizer.SynthesizeRE(token, posTag, true)
	} else {
		forms, err = s.BaseSynthesizer.Synthesize(token, posTag)
	}
	if err != nil {
		return nil, err
	}
	// Java lookup case filter (before empty → compound decision).
	forms = filterCaseMatch(lemmaOf(token), forms)
	if len(forms) == 0 {
		// Java: return getCompoundForms without REMOVE filter on joined forms.
		return s.getCompoundForms(token, posTag, posTagRegExp), nil
	}
	// Java synthesize: filter REMOVE only when super result non-empty.
	return filterRemove(forms), nil
}

// getCompoundForms ports GermanSynthesizer.getCompoundForms.
// Requires StrictCompoundTokenize (Java GermanCompoundTokenizer.getStrictInstance).
func (s *GermanSynthesizer) getCompoundForms(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) []string {
	if s == nil || s.StrictCompoundTokenize == nil || token == nil {
		return nil
	}
	lemma := lemmaOf(token)
	if lemma == "" {
		return nil
	}
	parts := s.StrictCompoundTokenize(lemma)
	if len(parts) == 0 {
		return nil
	}
	maybeHyphen := ""
	if len(parts) == 1 {
		hy := strings.Split(lemma, "-")
		if len(hy) > 1 {
			parts = hy
			maybeHyphen = "-"
		}
	}
	// firstPart = join(parts[0:n-1]); lastPart = uppercaseFirst(parts[n-1])
	firstPart := strings.Join(parts[:len(parts)-1], maybeHyphen)
	lastRaw := parts[len(parts)-1]
	lastPart := tools.UppercaseFirstChar(lastRaw)
	uppercaseLastPart := maybeHyphen != "" && tools.StartsWithUppercase(lastRaw)

	lastLemma := lastPart
	lastTok := languagetool.NewAnalyzedToken(lastPart, nil, &lastLemma)
	var lastForms []string
	var err error
	if posTagRegExp {
		lastForms, err = s.BaseSynthesizer.SynthesizeRE(lastTok, posTag, true)
	} else {
		lastForms, err = s.BaseSynthesizer.Synthesize(lastTok, posTag)
	}
	if err != nil || len(lastForms) == 0 {
		return nil
	}
	// Java super.synthesize applies lookup case filter for last part.
	lastForms = filterCaseMatch(lastPart, lastForms)
	if len(lastForms) == 0 {
		return nil
	}

	var results []string
	seen := map[string]struct{}{}
	for _, part := range lastForms {
		if part == "" {
			continue
		}
		var form string
		if uppercaseLastPart {
			form = firstPart + maybeHyphen + part
		} else {
			form = firstPart + maybeHyphen + tools.LowercaseFirstChar(part)
		}
		if _, ok := seen[form]; ok {
			continue
		}
		seen[form] = struct{}{}
		results = append(results, form)
	}
	return results
}

func lemmaOf(token *languagetool.AnalyzedToken) string {
	if token == nil {
		return ""
	}
	if token.GetLemma() != nil && *token.GetLemma() != "" {
		return *token.GetLemma()
	}
	return token.GetToken()
}

// filterCaseMatch ports GermanSynthesizer.lookup case gate (not REMOVE).
func filterCaseMatch(lemma string, forms []string) []string {
	if len(forms) == 0 {
		return forms
	}
	lcLemma := tools.StartsWithLowercase(lemma)
	var out []string
	seen := map[string]struct{}{}
	for _, form := range forms {
		if form == "" {
			continue
		}
		lcForm := tools.StartsWithLowercase(form)
		// Java: lcLemma == lcLookup || lemma.equals("mein") || (lemma.equals("ich") && !REMOVE)
		// REMOVE is applied later on the outer synthesize path; here only case / mein / ich.
		if lcLemma == lcForm || lemma == "mein" || lemma == "ich" {
			if _, ok := seen[form]; ok {
				continue
			}
			seen[form] = struct{}{}
			out = append(out, form)
		}
	}
	return out
}

// filterRemove ports synthesize-time REMOVE filter.
func filterRemove(forms []string) []string {
	if len(forms) == 0 {
		return forms
	}
	var out []string
	for _, form := range forms {
		if _, bad := germanSynthRemove[form]; bad {
			continue
		}
		out = append(out, form)
	}
	return out
}

// filterGermanForms keeps case+REMOVE for callers/tests that filter a raw form list.
func filterGermanForms(lemma string, forms []string) []string {
	return filterRemove(filterCaseMatch(lemma, forms))
}

var _ synthesis.Synthesizer = (*GermanSynthesizer)(nil)
