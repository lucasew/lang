package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	synthca "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ca"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AdjustPronounsFilter ports
// org.languagetool.rules.ca.AdjustPronounsFilter (1:1 AcceptRuleMatch).
//
// Synthesize is used only when arguments newLemma / newOnlyLemma are set
// (via VerbSynthesizer.synthesize). Action rewrites do not require it.
type AdjustPronounsFilter struct {
	// Synthesize ports language.getSynthesizer().synthesize(token, postag).
	Synthesize func(tok *languagetool.AnalyzedToken, postag string) []string
	// AdaptSuggestion ports language.adaptSuggestion for VerbSynthesizer; nil → identity.
	AdaptSuggestion func(s, originalErrorStr string) string
}

func NewAdjustPronounsFilter() *AdjustPronounsFilter {
	return &AdjustPronounsFilter{}
}

// PronounVerbContext is the surface input for Suggest (unit helper).
type PronounVerbContext struct {
	PronounsStr            string
	VerbStr                string
	VerbStr2               string
	FirstVerbPersonaNumber string
	PronounsAfter          bool
	WholeOriginal          string
	CasingModel            string
}

// Suggest applies comma-separated actions and returns unique replacements (unit helper).
func (f *AdjustPronounsFilter) Suggest(ctx PronounVerbContext, actionsCSV string) []string {
	var out []string
	seen := map[string]struct{}{}
	actions := strings.Split(actionsCSV, ",")
	for _, action := range actions {
		action = tools.JavaStringTrim(action)
		var replacement string
		switch action {
		case "removePronounEn":
			pr := tools.JavaStringTrim(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(ctx.PronounsStr, "en", ""), "n'", ""), "'n", ""))
			replacement = TransformDavant(pr, ctx.VerbStr) + ctx.VerbStr
		case "addPronounEn":
			newPronoun := DoAddPronounEn(ctx.PronounsStr, ctx.VerbStr, ctx.PronounsAfter)
			if newPronoun != "" {
				if !ctx.PronounsAfter {
					replacement = newPronoun + ctx.VerbStr
				} else {
					replacement = ctx.VerbStr + newPronoun
				}
			}
		case "removePronounReflexive":
			replacement = DoRemovePronounReflexive(ctx.PronounsStr, ctx.VerbStr, false)
		case "replaceEmEn":
			replacement = DoReplaceEmEn(ctx.PronounsStr, ctx.VerbStr, false)
		case "replaceHiEn":
			norm := Transform(tools.JavaStringTrim(strings.ReplaceAll(ctx.PronounsStr, "hi", "")), PronounNormalized)
			replacement = DoAddPronounEn(norm, ctx.VerbStr, false) + ctx.VerbStr
		case "addPronounReflexive":
			replacement = DoAddPronounReflexive(ctx.PronounsStr, ctx.VerbStr, ctx.FirstVerbPersonaNumber, ctx.PronounsAfter)
		case "addPronounReflexiveHi":
			replacement = DoAddPronounReflexive(ctx.PronounsStr, "hi "+ctx.VerbStr, ctx.FirstVerbPersonaNumber, false)
		case "addPronounReflexiveImperative":
			replacement = DoAddPronounReflexiveImperative(ctx.PronounsStr, ctx.VerbStr, ctx.FirstVerbPersonaNumber)
		case "changeOnlyLemma":
			pronounsStr := ctx.PronounsStr
			if containsAction(actions, "replaceHiEn") {
				pronounsStr = tools.JavaStringTrim(strings.ReplaceAll(pronounsStr, "hi", ""))
			}
			verb2 := ctx.VerbStr2
			if verb2 == "" {
				verb2 = ctx.VerbStr
			}
			pronounsStr = TransformDavant(pronounsStr, verb2)
			replacement = pronounsStr + verb2
		}
		if replacement == "" {
			continue
		}
		if strings.EqualFold(replacement, ctx.WholeOriginal) {
			continue
		}
		if ctx.CasingModel != "" {
			replacement = tools.PreserveCase(replacement, ctx.CasingModel)
		}
		if _, ok := seen[replacement]; ok {
			continue
		}
		seen[replacement] = struct{}{}
		out = append(out, replacement)
	}
	return out
}

func containsAction(actions []string, want string) bool {
	for _, a := range actions {
		if tools.JavaStringTrim(a) == want {
			return true
		}
	}
	return false
}

// AcceptRuleMatch ports AdjustPronounsFilter.acceptRuleMatch.
func (f *AdjustPronounsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	_ = patternTokenPos
	_ = patternTokens
	_ = tokenPositions
	if f == nil || match == nil || match.Sentence == nil {
		return nil
	}
	if arguments == nil {
		return nil
	}
	// Java getRequired("actions") panics if missing — same.
	actionsCSV := patterns.GetRequired("actions", arguments)
	actions := strings.Split(actionsCSV, ",")
	newLemma := patterns.GetOptional("newLemma", arguments)
	newOnlyLemma := patterns.GetOptional("newOnlyLemma", arguments)

	tokens := match.Sentence.GetTokensWithoutWhitespace()
	posWord := 0
	for posWord < len(tokens) &&
		(tokens[posWord].GetStartPos() < match.GetFromPos() || tokens[posWord].IsSentenceStart()) {
		posWord++
	}
	if posWord >= len(tokens) {
		return nil
	}

	verbSynth := synthca.NewVerbSynthesizerAt(tokens, posWord, false)
	if f.Synthesize != nil {
		verbSynth.Synthesize = f.Synthesize
	}
	if f.AdaptSuggestion != nil {
		verbSynth.AdaptSuggestion = f.AdaptSuggestion
	}
	if verbSynth.IsUndefined() {
		return nil
	}

	verbStr := verbSynth.GetVerbStr()
	if newLemma != "" {
		// Java only works when synthesizer is present; empty synth → empty verbStr
		if f.Synthesize == nil {
			return nil
		}
		verbSynth.SetLemma(newLemma)
		verbStr = verbSynth.SynthesizeForm()
	}
	verbStr2 := verbStr
	if newOnlyLemma != "" {
		if f.Synthesize == nil {
			return nil
		}
		verbSynth.SetLemma(newOnlyLemma)
		verbStr2 = verbSynth.SynthesizeForm()
	}

	firstVerbPersonaNumber := verbSynth.GetFirstVerbPersonaNumber()
	pronounsStr := ""
	if verbSynth.GetNumPronounsBefore() > 0 {
		pronounsStr = verbSynth.GetPronounsStrBefore()
	} else if verbSynth.GetNumPronounsAfter() > 0 {
		pronounsStr = verbSynth.GetPronounsStrAfter()
	}
	startUnderlineIndex := verbSynth.GetFirstVerbIndex() - verbSynth.GetNumPronounsBefore()
	endUnderlineIndex := verbSynth.GetLastVerbIndex() + verbSynth.GetNumPronounsAfter()
	if startUnderlineIndex < 0 || endUnderlineIndex >= len(tokens) || startUnderlineIndex > endUnderlineIndex {
		return nil
	}

	// isFirstVerbIS → pronouns after flag is !isFirstVerbIS for addPronounEn / reflexive
	pronounsAfter := !verbSynth.IsFirstVerbIS()
	wholeOriginal := verbSynth.GetWholeOriginalStr()
	casingModel := verbSynth.GetCasingModel()

	var replacements []string
	for _, action := range actions {
		action = tools.JavaStringTrim(action)
		var replacement string
		switch action {
		case "removePronounEn":
			pr := tools.JavaStringTrim(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(pronounsStr, "en", ""), "n'", ""), "'n", ""))
			replacement = TransformDavant(pr, verbStr) + verbStr
		case "addPronounEn":
			newPronoun := DoAddPronounEn(pronounsStr, verbStr, pronounsAfter)
			if newPronoun != "" {
				if verbSynth.IsFirstVerbIS() {
					replacement = newPronoun + verbStr
				} else {
					replacement = verbStr + newPronoun
				}
			}
		case "removePronounReflexive":
			replacement = DoRemovePronounReflexive(pronounsStr, verbStr, false)
		case "replaceEmEn":
			replacement = DoReplaceEmEn(pronounsStr, verbStr, false)
		case "replaceHiEn":
			norm := Transform(tools.JavaStringTrim(strings.ReplaceAll(pronounsStr, "hi", "")), PronounNormalized)
			replacement = DoAddPronounEn(norm, verbStr, false) + verbStr
		case "addPronounReflexive":
			replacement = DoAddPronounReflexive(pronounsStr, verbStr, firstVerbPersonaNumber, pronounsAfter)
		case "addPronounReflexiveHi":
			replacement = DoAddPronounReflexive(pronounsStr, "hi "+verbStr, firstVerbPersonaNumber, false)
		case "addPronounReflexiveImperative":
			replacement = DoAddPronounReflexiveImperative(pronounsStr, verbStr, verbSynth.GetFirstVerbPersonaNumberImperative())
		case "changeOnlyLemma":
			ps := pronounsStr
			if containsAction(actions, "replaceHiEn") {
				ps = tools.JavaStringTrim(strings.ReplaceAll(ps, "hi", ""))
			}
			ps = TransformDavant(ps, verbStr2)
			replacement = ps + verbStr2
		}
		if replacement != "" && !strings.EqualFold(replacement, wholeOriginal) {
			replacements = append(replacements, tools.PreserveCase(replacement, casingModel))
		}
	}
	if len(replacements) == 0 {
		return nil
	}

	out := rules.NewRuleMatch(match.GetRule(), match.Sentence,
		tokens[startUnderlineIndex].GetStartPos(), tokens[endUnderlineIndex].GetEndPos(),
		match.GetMessage())
	out.ShortMessage = match.GetShortMessage()
	out.SetSuggestedReplacements(replacements)
	return out
}
