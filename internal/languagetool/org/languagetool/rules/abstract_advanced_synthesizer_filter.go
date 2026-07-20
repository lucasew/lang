package rules

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AbstractAdvancedSynthesizerFilter ports
// org.languagetool.rules.AbstractAdvancedSynthesizerFilter with a pluggable synthesizer.
// Combines lemma from one pattern token with POS from another.
type AbstractAdvancedSynthesizerFilter struct {
	// Synthesize(lemma, postag) → surface forms (Java synth.synthesize with regex postag).
	// Nil → fail-closed (return nil; do not invent forms).
	Synthesize func(lemma, postag string) []string
	// GetNewLemma ports protected getNewLemma(word, newLemma) when newLemma starts with "_".
	// Java base returns null (drop match). Language subclasses (e.g. CA) override.
	// Empty string result means null — Accept returns nil.
	GetNewLemma func(word, newLemma string) string
	// IsSuggestionException optional; Java language overrides.
	IsSuggestionException func(token, desiredPostag string) bool
	// GetCompositePostag optional override of getCompositePostag (e.g. PT keepPronoun).
	// When nil, package GetCompositePostag is used.
	GetCompositePostag func(lemmaSelect, postagSelect, originalPostag, desiredPostag, postagReplace string) string
	// AdaptSuggestion optional language.adaptSuggestion.
	AdaptSuggestion func(replacement, original string) string
}

// AcceptRuleMatch requires lemmaFrom, postagFrom, lemmaSelect, postagSelect.
// Optional: newLemma, postagReplace.
func (f *AbstractAdvancedSynthesizerFilter) AcceptRuleMatch(match *RuleMatch, args map[string]string,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	_ = tokenPositions
	if f.Synthesize == nil {
		// Without synthesizer cannot produce forms (fail-closed).
		return nil
	}
	postagSelect := requireArg(args, "postagSelect")
	lemmaSelect := requireArg(args, "lemmaSelect")
	postagFromStr := requireArg(args, "postagFrom")
	lemmaFromStr := requireArg(args, "lemmaFrom")
	newLemma := args["newLemma"]
	postagReplace := args["postagReplace"]

	postagFrom := resolveIndex(postagFromStr, match, patternTokens)
	lemmaFrom := resolveIndex(lemmaFromStr, match, patternTokens)
	if postagFrom < 0 || postagFrom >= len(patternTokens) || lemmaFrom < 0 || lemmaFrom >= len(patternTokens) {
		panic(fmt.Sprintf("AdvancedSynthesizerFilter: Index out of bounds, postagFrom=%s lemmaFrom=%s", postagFromStr, lemmaFromStr))
	}

	lemmaTok := getAnalyzedToken(patternTokens[lemmaFrom], lemmaSelect)
	posTok := getAnalyzedToken(patternTokens[postagFrom], postagSelect)
	if lemmaTok == nil {
		return nil
	}
	desiredLemma := ""
	if lemmaTok.GetLemma() != nil {
		desiredLemma = *lemmaTok.GetLemma()
	}
	originalPostag := ""
	if lemmaTok.GetPOSTag() != nil {
		originalPostag = *lemmaTok.GetPOSTag()
	}
	if desiredLemma == "" {
		return nil
	}
	if posTok == nil || posTok.GetPOSTag() == nil {
		panic(fmt.Sprintf("AdvancedSynthesizerFilter: undefined POS tag with POS regex '%s'", postagSelect))
	}
	desiredPostag := *posTok.GetPOSTag()
	if newLemma != "" {
		if strings.HasPrefix(newLemma, "_") {
			// Java: desiredLemma = getNewLemma(...); if null return null
			if f.GetNewLemma != nil {
				desiredLemma = f.GetNewLemma(desiredLemma, newLemma)
			} else {
				desiredLemma = "" // Java base getNewLemma returns null
			}
		} else {
			desiredLemma = newLemma
		}
	}
	if desiredLemma == "" {
		return nil
	}
	if postagReplace != "" {
		if f.GetCompositePostag != nil {
			desiredPostag = f.GetCompositePostag(lemmaSelect, postagSelect, originalPostag, desiredPostag, postagReplace)
		} else {
			desiredPostag = GetCompositePostag(lemmaSelect, postagSelect, originalPostag, desiredPostag, postagReplace)
		}
	}

	isCap := tools.IsCapitalizedWord(patternTokens[lemmaFrom].GetToken())
	isAllUpper := tools.IsAllUppercase(patternTokens[lemmaFrom].GetToken())
	replacements := f.Synthesize(desiredLemma, desiredPostag)
	if len(replacements) == 0 {
		// Java returns original match when synth yields nothing.
		return match
	}

	out := NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	var list []string
	suggestionUsed := false
	existing := match.GetSuggestedReplacements()
	if len(existing) == 0 {
		existing = []string{""}
	}
	for _, r := range existing {
		for _, nr := range replacements {
			if f.IsSuggestionException != nil && f.IsSuggestionException(nr, desiredPostag) {
				continue
			}
			if strings.Contains(r, "{suggestion}") || strings.Contains(r, "{Suggestion}") || strings.Contains(r, "{SUGGESTION}") {
				suggestionUsed = true
			}
			form := nr
			if isCap {
				form = tools.UppercaseFirstChar(form)
			}
			if isAllUpper {
				form = strings.ToUpper(form)
			}
			complete := r
			if complete == "" {
				complete = form
			} else {
				complete = strings.ReplaceAll(complete, "{suggestion}", form)
				complete = strings.ReplaceAll(complete, "{Suggestion}", tools.UppercaseFirstChar(form))
				complete = strings.ReplaceAll(complete, "{SUGGESTION}", strings.ToUpper(form))
			}
			if !sliceHasString(list, complete) {
				list = append(list, complete)
			}
		}
	}
	if !suggestionUsed {
		for _, nr := range replacements {
			if f.IsSuggestionException != nil && f.IsSuggestionException(nr, desiredPostag) {
				continue
			}
			form := nr
			if isCap {
				form = tools.UppercaseFirstChar(form)
			}
			if isAllUpper {
				form = strings.ToUpper(form)
			}
			if !sliceHasString(list, form) {
				list = append(list, form)
			}
		}
	}
	if f.AdaptSuggestion != nil {
		adj := make([]string, 0, len(list))
		for _, r := range list {
			adj = append(adj, f.AdaptSuggestion(r, ""))
		}
		list = adj
	}
	out.SetSuggestedReplacements(list)
	return out
}

func sliceHasString(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// GetCompositePostag ports AbstractAdvancedSynthesizerFilter.getCompositePostag (\aN / \bN).
func GetCompositePostag(lemmaSelect, postagSelect, originalPostag, desiredPostag, postagReplace string) string {
	aRE, err1 := regexp.Compile("(?i)" + lemmaSelect)
	bRE, err2 := regexp.Compile("(?i)" + postagSelect)
	if err1 != nil || err2 != nil {
		return postagReplace
	}
	aM := aRE.FindStringSubmatch(originalPostag)
	bM := bRE.FindStringSubmatch(desiredPostag)
	// Java Matcher.matches requires full string
	if aM == nil || aM[0] != originalPostag || bM == nil || bM[0] != desiredPostag {
		return postagReplace
	}
	result := postagReplace
	for i := 1; i < len(aM); i++ {
		if aM[i] != "" {
			result = strings.ReplaceAll(result, `\a`+strconv.Itoa(i), aM[i])
		}
	}
	for i := 1; i < len(bM); i++ {
		if bM[i] != "" {
			result = strings.ReplaceAll(result, `\b`+strconv.Itoa(i), bM[i])
		}
	}
	return result
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

func getAnalyzedToken(atr *languagetool.AnalyzedTokenReadings, regexpStr string) *languagetool.AnalyzedToken {
	if atr == nil {
		return nil
	}
	re, err := regexp.Compile(regexpStr)
	if err != nil {
		re = regexp.MustCompile("^(?:" + regexp.QuoteMeta(regexpStr) + ")$")
	}
	// Java getAnalyzedToken: prefer lemma match then POS
	for _, r := range atr.GetReadings() {
		if r == nil {
			continue
		}
		if lem := r.GetLemma(); lem != nil && re.MatchString(*lem) {
			return r
		}
	}
	for _, r := range atr.GetReadings() {
		if r == nil {
			continue
		}
		if pt := r.GetPOSTag(); pt != nil && re.MatchString(*pt) {
			return r
		}
	}
	// fallback first reading
	rs := atr.GetReadings()
	if len(rs) > 0 {
		return rs[0]
	}
	return nil
}

func selectLemma(atr *languagetool.AnalyzedTokenReadings, lemmaSelect string) string {
	tok := getAnalyzedToken(atr, lemmaSelect)
	if tok == nil || tok.GetLemma() == nil {
		return ""
	}
	return *tok.GetLemma()
}

func selectPostag(atr *languagetool.AnalyzedTokenReadings, postagSelect string) string {
	tok := getAnalyzedToken(atr, postagSelect)
	if tok == nil || tok.GetPOSTag() == nil {
		return ""
	}
	return *tok.GetPOSTag()
}
