package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	synthca "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ca"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AnarASuggestionsFilter ports
// org.languagetool.rules.ca.AnarASuggestionsFilter (1:1 AcceptRuleMatch).
//
// Synthesize ports synthesizer.synthesize(token, postagRE, true).
// SynthFuturePresent is a legacy unit-test hook for Suggest.
type AnarASuggestionsFilter struct {
	// Synthesize ports getSynthesizerFromRuleMatch(...).synthesize(token, postag, true).
	Synthesize func(tok *languagetool.AnalyzedToken, postagRE string) []string
	// SynthFuturePresent returns future then present forms (Suggest helper).
	// personNumberSuffix is anarPostag[4:8] (e.g. "1S00" or "1S0.").
	SynthFuturePresent func(lemma, personNumberSuffix string) []string
}

func NewAnarASuggestionsFilter() *AnarASuggestionsFilter {
	return &AnarASuggestionsFilter{}
}

// Suggest builds "li ho farem / li ho fem" style replacements (unit helper).
func (f *AnarASuggestionsFilter) Suggest(lemma, personNumberSuffix, pronouns, casingModel string) []string {
	var forms []string
	if f.SynthFuturePresent != nil {
		forms = f.SynthFuturePresent(lemma, personNumberSuffix)
	} else if f.Synthesize != nil {
		forms = f.synthFuturePresent(lemma, personNumberSuffix)
	}
	if len(forms) == 0 {
		return nil
	}
	var out []string
	for _, verb := range forms {
		s := ""
		if pronouns != "" {
			s = TransformDavant(pronouns, verb)
		}
		s += verb
		if casingModel != "" {
			s = tools.PreserveCase(s, casingModel)
		}
		s = AdaptSuggestion(s, "")
		out = append(out, s)
	}
	return out
}

// synthFuturePresent ports Java future + present synthesize with V[MS]IF/IP + person suffix.
func (f *AnarASuggestionsFilter) synthFuturePresent(lemma, personNumberSuffix string) []string {
	if f.Synthesize == nil || lemma == "" {
		return nil
	}
	at := languagetool.NewAnalyzedToken("", nil, &lemma)
	var out []string
	// Java: "V[MS]IF" + verbPostag.substring(4, 8)
	for _, form := range f.Synthesize(at, "V[MS]IF"+personNumberSuffix) {
		out = append(out, form)
	}
	for _, form := range f.Synthesize(at, "V[MS]IP"+personNumberSuffix) {
		out = append(out, form)
	}
	return out
}

// AcceptRuleMatch ports AnarASuggestionsFilter.acceptRuleMatch.
func (f *AnarASuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = arguments
	_ = patternTokenPos
	_ = patternTokens
	_ = tokenPositions
	if f == nil || match == nil || match.Sentence == nil {
		return nil
	}
	if f.Synthesize == nil && f.SynthFuturePresent == nil {
		return nil
	}

	tokens := match.Sentence.GetTokensWithoutWhitespace()
	initPos := 0
	for initPos < len(tokens) &&
		(tokens[initPos].GetStartPos() < match.GetFromPos() || tokens[initPos].IsSentenceStart()) {
		initPos++
	}
	if initPos >= len(tokens) {
		return nil
	}

	verbSynthesizer := synthca.NewVerbSynthesizerAt(tokens, initPos, false)
	if f.Synthesize != nil {
		verbSynthesizer.Synthesize = f.Synthesize
	}
	if verbSynthesizer.IsUndefined() {
		return nil
	}
	if tokens[verbSynthesizer.GetLastVerbIndex()].GetEndPos() > match.GetToPos() {
		return nil
	}
	initPos = verbSynthesizer.GetFirstVerbIndex()
	// need anar + a + infinitive at initPos, initPos+2
	if initPos+2 >= len(tokens) {
		return nil
	}

	atrAnar := readingWithTagRegex(tokens[initPos], `V.IP.*`)
	atrInf := readingWithTagRegex(tokens[initPos+2], `V.N.*`)
	if atrAnar == nil || atrInf == nil || atrAnar.GetPOSTag() == nil {
		return nil
	}
	verbPostag := *atrAnar.GetPOSTag()
	if len(verbPostag) < 8 {
		return nil
	}
	lemma := ""
	if atrInf.GetLemma() != nil {
		lemma = *atrInf.GetLemma()
	}
	if lemma == "" {
		return nil
	}

	personSuffix := verbPostag[4:8]
	var synthFormsList []string
	if f.Synthesize != nil {
		synthFormsList = f.synthFuturePresent(lemma, personSuffix)
	} else if f.SynthFuturePresent != nil {
		synthFormsList = f.SynthFuturePresent(lemma, personSuffix)
	}
	if len(synthFormsList) == 0 {
		return nil
	}

	adjustEndPos := 0
	pronomsDarrere := verbSynthesizer.GetPronounsStrAfter()
	adjustEndPos += verbSynthesizer.GetNumPronounsAfter()

	adjustStartPos := 0
	pronomsDavant := verbSynthesizer.GetPronounsStrBefore()
	adjustStartPos += verbSynthesizer.GetNumPronounsBefore()

	var replacements []string
	// Java starts with existing match suggestions
	replacements = append(replacements, match.GetSuggestedReplacements()...)
	for _, verb := range synthFormsList {
		suggestion := ""
		if pronomsDarrere != "" {
			suggestion = TransformDavant(pronomsDarrere, verb)
		} else if pronomsDavant != "" {
			suggestion = TransformDavant(pronomsDavant, verb)
		}
		suggestion += verb
		casingIdx := initPos - adjustStartPos
		if casingIdx < 0 || casingIdx >= len(tokens) {
			continue
		}
		suggestion = tools.PreserveCase(suggestion, tokens[casingIdx].GetToken())
		replacements = append(replacements, suggestion)
	}
	if len(replacements) == 0 {
		return nil
	}

	startIdx := initPos - adjustStartPos
	endIdx := initPos + 2 + adjustEndPos
	if startIdx < 0 || endIdx >= len(tokens) {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence,
		tokens[startIdx].GetStartPos(), tokens[endIdx].GetEndPos(),
		match.GetMessage())
	out.ShortMessage = match.GetShortMessage()
	// Java: adaptSuggestionsList(replacements, "")
	adapted := make([]string, 0, len(replacements))
	for _, r := range replacements {
		adapted = append(adapted, AdaptSuggestion(r, ""))
	}
	out.SetSuggestedReplacements(adapted)
	return out
}
