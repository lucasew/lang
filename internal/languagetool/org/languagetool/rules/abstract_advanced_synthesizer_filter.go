package rules

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractAdvancedSynthesizerFilter ports
// org.languagetool.rules.AbstractAdvancedSynthesizerFilter with a pluggable synthesizer.
// Combines lemma from one pattern token with POS from another.
type AbstractAdvancedSynthesizerFilter struct {
	// Synthesize(lemma, postag) → surface forms.
	Synthesize func(lemma, postag string) []string
}

// AcceptRuleMatch requires lemmaFrom, postagFrom, lemmaSelect, postagSelect.
// Optional: newLemma.
func (f *AbstractAdvancedSynthesizerFilter) AcceptRuleMatch(match *RuleMatch, args map[string]string,
	patternTokens []*languagetool.AnalyzedTokenReadings) *RuleMatch {
	if f == nil || f.Synthesize == nil || match == nil {
		return nil
	}
	postagSelect := requireArg(args, "postagSelect")
	lemmaSelect := requireArg(args, "lemmaSelect")
	postagFromStr := requireArg(args, "postagFrom")
	lemmaFromStr := requireArg(args, "lemmaFrom")
	newLemma := args["newLemma"]

	postagFrom := resolveIndex(postagFromStr, match, patternTokens)
	lemmaFrom := resolveIndex(lemmaFromStr, match, patternTokens)
	if postagFrom < 0 || postagFrom >= len(patternTokens) || lemmaFrom < 0 || lemmaFrom >= len(patternTokens) {
		panic(fmt.Sprintf("AdvancedSynthesizerFilter: Index out of bounds, postagFrom=%s lemmaFrom=%s", postagFromStr, lemmaFromStr))
	}

	lemma := selectLemma(patternTokens[lemmaFrom], lemmaSelect)
	if newLemma != "" {
		lemma = newLemma
	}
	postag := selectPostag(patternTokens[postagFrom], postagSelect)
	if lemma == "" || postag == "" {
		return nil
	}
	forms := f.Synthesize(lemma, postag)
	if len(forms) == 0 {
		return nil
	}
	match.SetSuggestedReplacements(forms)
	return match
}

func requireArg(args map[string]string, key string) string {
	v, ok := args[key]
	if !ok {
		panic("Missing key '" + key + "'")
	}
	return v
}

func resolveIndex(spec string, match *RuleMatch, patternTokens []*languagetool.AnalyzedTokenReadings) int {
	if strings.HasPrefix(spec, "marker") {
		i := 0
		for i < len(patternTokens) && patternTokens[i].GetStartPos() < match.GetFromPos() {
			i++
		}
		i++ // 1-based like Java after marker walk
		if len(spec) > 6 {
			off, _ := strconv.Atoi(strings.TrimPrefix(spec, "marker"))
			i += off
		}
		return i - 1 // convert to 0-based for Go slice
	}
	n, err := strconv.Atoi(spec)
	if err != nil {
		panic(err)
	}
	return n - 1 // Java is 1-based
}

func selectLemma(atr *languagetool.AnalyzedTokenReadings, lemmaSelect string) string {
	re, err := regexp.Compile(lemmaSelect)
	if err != nil {
		re = regexp.MustCompile("^(?:" + regexp.QuoteMeta(lemmaSelect) + ")$")
	}
	for _, r := range atr.GetReadings() {
		if lem := r.GetLemma(); lem != nil && re.MatchString(*lem) {
			return *lem
		}
	}
	// fallback first lemma
	for _, r := range atr.GetReadings() {
		if lem := r.GetLemma(); lem != nil {
			return *lem
		}
	}
	return ""
}

func selectPostag(atr *languagetool.AnalyzedTokenReadings, postagSelect string) string {
	re, err := regexp.Compile(postagSelect)
	if err != nil {
		re = regexp.MustCompile(regexp.QuoteMeta(postagSelect))
	}
	for _, r := range atr.GetReadings() {
		if pt := r.GetPOSTag(); pt != nil && re.MatchString(*pt) {
			return *pt
		}
	}
	for _, r := range atr.GetReadings() {
		if pt := r.GetPOSTag(); pt != nil {
			return *pt
		}
	}
	return ""
}
