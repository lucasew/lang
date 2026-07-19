package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	synthca "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ca"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PortarGerundiSuggestionsFilter ports
// org.languagetool.rules.ca.PortarGerundiSuggestionsFilter (1:1 AcceptRuleMatch).
//
// Synthesize ports synthesizer.synthesize(token, postagRE, true).
// Legacy SynthHaverParticiple / SynthFinite remain for unit Suggest helpers.
type PortarGerundiSuggestionsFilter struct {
	// Synthesize ports getSynthesizerFromRuleMatch(...).synthesize(token, postag, true).
	Synthesize func(tok *languagetool.AnalyzedToken, postagRE string) []string
	// SynthHaverParticiple returns "he fet"-style strings (Suggest helper).
	SynthHaverParticiple func(lemma, portarPostagSuffix string) []string
	// SynthFinite returns finite forms of the gerund lemma (Suggest helper).
	SynthFinite func(lemma, portarPostagSuffix string) []string
}

func NewPortarGerundiSuggestionsFilter() *PortarGerundiSuggestionsFilter {
	return &PortarGerundiSuggestionsFilter{}
}

// Suggest builds replacements for "porto fent-ho" style matches (unit helper).
func (f *PortarGerundiSuggestionsFilter) Suggest(portarPostag, lemma, pronounsAfter, casingModel string) []string {
	if len(portarPostag) < 8 {
		return nil
	}
	suffix := portarPostag[2:]
	var raw []string
	if f.SynthHaverParticiple != nil {
		raw = append(raw, f.SynthHaverParticiple(lemma, suffix)...)
	}
	if f.SynthFinite != nil {
		if forms := f.SynthFinite(lemma, suffix); len(forms) > 0 {
			raw = append(raw, forms[0])
		}
	}
	// Prefer full Synthesize when wired
	if len(raw) == 0 && f.Synthesize != nil {
		raw = f.synthAll(lemma, portarPostag)
	}
	if len(raw) == 0 {
		return nil
	}
	var out []string
	for _, r := range raw {
		s := r
		if pronounsAfter != "" {
			s = TransformDavant(pronounsAfter, r) + r
		}
		if casingModel != "" {
			s = tools.PreserveCase(s, casingModel)
		}
		out = append(out, s)
	}
	return out
}

func (f *PortarGerundiSuggestionsFilter) synthAll(lemma, portarPostag string) []string {
	if f.Synthesize == nil || len(portarPostag) < 3 {
		return nil
	}
	suffix := portarPostag[2:]
	haverLemma := "haver"
	var replacements []string
	// he fet
	synthForms1 := f.Synthesize(languagetool.NewAnalyzedToken("", nil, &haverLemma), "VA"+suffix)
	synthForms2 := f.Synthesize(languagetool.NewAnalyzedToken("", nil, &lemma), "V.P..SM.")
	for _, h := range synthForms1 {
		for _, p := range synthForms2 {
			replacements = append(replacements, h+" "+p)
		}
	}
	// faig
	synthForms3 := f.Synthesize(languagetool.NewAnalyzedToken("", nil, &lemma), "V."+suffix)
	if len(synthForms3) > 0 {
		replacements = append(replacements, synthForms3[0])
	}
	return replacements
}

// JoinHaverParticiple is a helper for tests building "he fet" pairs.
func JoinHaverParticiple(haverForms, partForms []string) []string {
	var out []string
	for _, h := range haverForms {
		for _, p := range partForms {
			out = append(out, h+" "+p)
		}
	}
	return out
}

// AcceptRuleMatch ports PortarGerundiSuggestionsFilter.acceptRuleMatch.
func (f *PortarGerundiSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = patternTokenPos
	_ = patternTokens
	_ = tokenPositions
	if f == nil || match == nil || match.Sentence == nil {
		return nil
	}
	// Need Synthesize or legacy hooks
	if f.Synthesize == nil && f.SynthHaverParticiple == nil && f.SynthFinite == nil {
		return nil
	}

	tokens := match.Sentence.GetTokensWithoutWhitespace()
	posWord := 0
	for posWord < len(tokens) &&
		(tokens[posWord].GetStartPos() < match.GetFromPos() || tokens[posWord].IsSentenceStart()) {
		posWord++
	}
	if posWord+1 >= len(tokens) || tokens[posWord] == nil || tokens[posWord+1] == nil {
		return nil
	}

	// Java: readingWithTagRegex("V.[IS].*") — full match, no ^ in Java pattern
	atr1 := readingWithTagRegex(tokens[posWord], `V.[IS].*`)
	atr2 := readingWithTagRegex(tokens[posWord+1], `V.G.*`)
	if atr1 == nil || atr2 == nil || atr1.GetPOSTag() == nil {
		return nil
	}
	portag := *atr1.GetPOSTag()
	if len(portag) < 3 {
		return nil
	}

	newLemma := patterns.GetOptionalDefault("newLemma", arguments, "")
	lemma := newLemma
	if lemma == "" {
		if atr2.GetLemma() != nil {
			lemma = *atr2.GetLemma()
		}
	}
	if lemma == "" {
		return nil
	}

	var replacements []string
	if f.Synthesize != nil {
		replacements = f.synthAll(lemma, portag)
	} else {
		// Legacy hooks (unit tests without full Synthesize)
		suffix := portag[2:]
		if f.SynthHaverParticiple != nil {
			replacements = append(replacements, f.SynthHaverParticiple(lemma, suffix)...)
		}
		if f.SynthFinite != nil {
			if forms := f.SynthFinite(lemma, suffix); len(forms) > 0 {
				replacements = append(replacements, forms[0])
			}
		}
	}
	if len(replacements) == 0 {
		return nil
	}

	verbSynthesizer := synthca.NewVerbSynthesizerAt(tokens, posWord, false)
	if f.Synthesize != nil {
		verbSynthesizer.Synthesize = f.Synthesize
	}
	correctStartIndex := 0
	correctEndIndex := 0
	for i := range replacements {
		pronounsSuggestion := ""
		if verbSynthesizer.GetNumPronounsAfter() > 0 {
			pronounsSuggestion = TransformDavant(verbSynthesizer.GetPronounsStrAfter(), replacements[i])
			correctEndIndex = verbSynthesizer.GetNumPronounsAfter()
		} else if verbSynthesizer.GetNumPronounsBefore() > 0 {
			pronounsSuggestion = TransformDavant(verbSynthesizer.GetPronounsStrBefore(), replacements[i])
			correctStartIndex = -verbSynthesizer.GetNumPronounsBefore()
		}
		casingTok := tokens[posWord+correctStartIndex].GetToken()
		replacements[i] = tools.PreserveCase(pronounsSuggestion+replacements[i], casingTok)
	}

	startIdx := posWord + correctStartIndex
	endIdx := posWord + 1 + correctEndIndex
	if startIdx < 0 || endIdx >= len(tokens) || startIdx > endIdx {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence,
		tokens[startIdx].GetStartPos(), tokens[endIdx].GetEndPos(),
		match.GetMessage())
	out.ShortMessage = match.GetShortMessage()
	out.SetSuggestedReplacements(replacements)
	return out
}
