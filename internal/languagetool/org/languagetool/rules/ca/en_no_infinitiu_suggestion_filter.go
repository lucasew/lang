package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	synthca "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ca"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// EnNoInfinitiuSuggestionFilter ports
// org.languagetool.rules.ca.EnNoInfinitiuSuggestionFilter (1:1 AcceptRuleMatch).
//
// Synthesize ports language.getSynthesizer().synthesize used by VerbSynthesizer.
// Synth is a legacy unit-test hook for Suggest (lemma, postag) → form.
type EnNoInfinitiuSuggestionFilter struct {
	// Synthesize ports getSynthesizerFromRuleMatch / language synthesizer.
	Synthesize func(tok *languagetool.AnalyzedToken, postag string) []string
	// Synth synthesizes the infinitive verb with a full postag (Suggest helper).
	Synth func(lemma, postag string) string
	// AdaptSuggestion ports language.adaptSuggestion for VerbSynthesizer; nil → identity.
	AdaptSuggestion func(s, originalErrorStr string) string
}

func NewEnNoInfinitiuSuggestionFilter() *EnNoInfinitiuSuggestionFilter {
	return &EnNoInfinitiuSuggestionFilter{}
}

// tempsVerbalsPresent ports Java tempsVerbalsPresent.
var tempsVerbalsPresent = map[string]struct{}{
	"IP": {},
	"IF": {},
}

// EnNoInfinitiuInput describes tense/person context around "en no + infinitive".
type EnNoInfinitiuInput struct {
	TempsVerbal       string
	PassatPerifrastic bool
	VerbBefore        bool
	Lemma             string
	PronounsAfter     string
	CasingModel       string
}

// Suggest builds "com que no …" / "perquè no …" finite rewrites (unit helper).
func (f *EnNoInfinitiuSuggestionFilter) Suggest(in EnNoInfinitiuInput) []string {
	if f == nil || f.Synth == nil || len(in.TempsVerbal) < 6 {
		return nil
	}
	prefix := "VMII"
	moodTense := in.TempsVerbal[2:4]
	if _, ok := tempsVerbalsPresent[moodTense]; ok && !in.PassatPerifrastic {
		prefix = "VMIP"
	}
	var synthVerbs []string
	personNumber := in.TempsVerbal[4:6]
	if personNumber != "3S" {
		if s := f.Synth(in.Lemma, prefix+"3S"+in.TempsVerbal[6:]); s != "" {
			synthVerbs = append(synthVerbs, s)
		}
	}
	if s := f.Synth(in.Lemma, prefix+in.TempsVerbal[4:]); s != "" {
		synthVerbs = append(synthVerbs, s)
	}
	intro := "com que no "
	if in.VerbBefore {
		intro = "perquè no "
	}
	var out []string
	seen := map[string]struct{}{}
	for _, v := range synthVerbs {
		var b strings.Builder
		b.WriteString(intro)
		if in.PronounsAfter != "" {
			b.WriteString(TransformDavant(in.PronounsAfter, v))
		}
		b.WriteString(v)
		s := b.String()
		if in.CasingModel != "" {
			s = tools.PreserveCase(s, in.CasingModel)
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func (f *EnNoInfinitiuSuggestionFilter) synthesizeHook() func(tok *languagetool.AnalyzedToken, postag string) []string {
	if f.Synthesize != nil {
		return f.Synthesize
	}
	if f.Synth != nil {
		return func(tok *languagetool.AnalyzedToken, postag string) []string {
			lemma := ""
			if tok != nil && tok.GetLemma() != nil {
				lemma = *tok.GetLemma()
			}
			s := f.Synth(lemma, postag)
			if s == "" {
				// Java still adds empty string from synthesize(); return [""] to match append.
				return []string{""}
			}
			return []string{s}
		}
	}
	return nil
}

// AcceptRuleMatch ports EnNoInfinitiuSuggestionFilter.acceptRuleMatch.
func (f *EnNoInfinitiuSuggestionFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = arguments
	_ = patternTokenPos
	_ = patternTokens
	_ = tokenPositions
	if f == nil || match == nil || match.Sentence == nil {
		return nil
	}
	synthHook := f.synthesizeHook()
	if synthHook == nil {
		return nil
	}

	tokens := match.Sentence.GetTokensWithoutWhitespace()
	posWord := 0
	for posWord < len(tokens) &&
		(tokens[posWord].GetStartPos() < match.GetFromPos() || tokens[posWord].IsSentenceStart()) {
		posWord++
	}
	if posWord >= len(tokens) {
		return nil
	}

	verbSynthInfinitiu := synthca.NewVerbSynthesizerAt(tokens, posWord, false)
	verbSynthInfinitiu.Synthesize = synthHook
	if f.AdaptSuggestion != nil {
		verbSynthInfinitiu.AdaptSuggestion = f.AdaptSuggestion
	}

	if verbSynthInfinitiu.IsUndefined() {
		return nil
	}
	lastVerbIdx := verbSynthInfinitiu.GetLastVerbIndex()
	if lastVerbIdx < 0 || lastVerbIdx >= len(tokens) {
		return nil
	}
	if tokens[lastVerbIdx].GetEndPos() > match.GetToPos() {
		return nil
	}

	posAfter := verbSynthInfinitiu.GetLastIndex() + 1
	verbAfter := synthca.NewVerbSynthesizerAt(tokens, posAfter, false)
	verbAfter.Synthesize = synthHook

	beforeStart := verbSynthInfinitiu.GetFirstVerbIndex() - 1
	verbBefore := synthca.NewVerbSynthesizerAt(tokens, beforeStart, true)
	verbBefore.Synthesize = synthHook

	postagTempsVerbal := ""
	isPassatPerifrastic := false
	verbAfterPostag := verbAfter.GetFirstVerbISPostag()
	verbBeforePostag := verbBefore.GetFirstVerbISPostag()
	if !verbAfter.IsUndefined() && verbAfterPostag != "" {
		postagTempsVerbal = verbAfterPostag
		isPassatPerifrastic = verbAfter.IsPassatPerifrastic()
	} else if !verbBefore.IsUndefined() && verbBeforePostag != "" {
		postagTempsVerbal = verbBeforePostag
		isPassatPerifrastic = verbBefore.IsPassatPerifrastic()
	} else {
		return nil
	}
	if len(postagTempsVerbal) < 6 {
		return nil
	}

	postagPrefix := "VMII"
	moodTense := postagTempsVerbal[2:4]
	if _, ok := tempsVerbalsPresent[moodTense]; ok && !isPassatPerifrastic {
		postagPrefix = "VMIP"
	}

	var synthVerbs []string
	if postagTempsVerbal[4:6] != "3S" {
		verbSynthInfinitiu.SetPostag(postagPrefix + "3S" + postagTempsVerbal[6:])
		synthVerbs = append(synthVerbs, verbSynthInfinitiu.SynthesizeForm())
	}
	verbSynthInfinitiu.SetPostag(postagPrefix + postagTempsVerbal[4:])
	synthVerbs = append(synthVerbs, verbSynthInfinitiu.SynthesizeForm())

	startPos := verbSynthInfinitiu.GetFirstVerbIndex() - 2
	if startPos < 0 || startPos >= len(tokens) {
		return nil
	}
	if strings.EqualFold(tokens[startPos].GetToken(), "l") {
		startPos--
		if startPos < 0 {
			return nil
		}
	}
	endPos := verbSynthInfinitiu.GetLastIndex()
	if endPos < 0 || endPos >= len(tokens) {
		return nil
	}

	text := match.Sentence.GetText()
	from := tokens[startPos].GetStartPos()
	to := tokens[endPos].GetEndPos()
	originalStr := ""
	if from >= 0 && to <= len(text) && from <= to {
		originalStr = text[from:to]
	} else {
		originalStr = tokens[startPos].GetToken()
	}

	pronounsAfter := verbSynthInfinitiu.GetPronounsStrAfter()
	usePerque := !verbBefore.IsUndefined() && verbBeforePostag != ""

	var suggestions []string
	for _, synthVerb := range synthVerbs {
		var suggestion strings.Builder
		if usePerque {
			suggestion.WriteString("perquè no ")
		} else {
			suggestion.WriteString("com que no ")
		}
		if pronounsAfter != "" {
			suggestion.WriteString(TransformDavant(pronounsAfter, synthVerb))
		}
		suggestion.WriteString(synthVerb)
		suggestionStr := tools.PreserveCase(suggestion.String(), originalStr)
		// Java List.contains
		dup := false
		for _, s := range suggestions {
			if s == suggestionStr {
				dup = true
				break
			}
		}
		if !dup {
			suggestions = append(suggestions, suggestionStr)
		}
	}
	if len(suggestions) == 0 {
		return nil
	}

	out := rules.NewRuleMatch(match.GetRule(), match.Sentence,
		tokens[startPos].GetStartPos(), tokens[endPos].GetEndPos(),
		match.GetMessage())
	out.ShortMessage = match.GetShortMessage()
	out.SetSuggestedReplacements(suggestions)
	return out
}
