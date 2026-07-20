package rules

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/symspell/implementation"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const findSuggestionsMax = 10

// AbstractFindSuggestionsFilter ports org.languagetool.rules.AbstractFindSuggestionsFilter.
//
// SpellingSuggestions / Tag are required for full behavior (Java abstract methods).
// Optional Synthesize enables the synthesizer fallback path (replacements2).
type AbstractFindSuggestionsFilter struct {
	// SpellingSuggestions returns candidates for the focused token (Java getSpellingSuggestions).
	SpellingSuggestions func(atr *languagetool.AnalyzedTokenReadings) []string
	// Tag tags a single word form (Java getTagger().tag).
	Tag func(word string) *languagetool.AnalyzedTokenReadings
	// Synthesize ports getSynthesizer().synthesize(token, desiredPostag, true); nil skips path.
	Synthesize func(tok *languagetool.AnalyzedToken, postagRE string) []string
	// IsSuggestionException ports isSuggestionException (default false).
	IsSuggestionException func(atr *languagetool.AnalyzedTokenReadings) bool
	// PreProcessWrongWord ports preProcessWrongWord (default: strip spaces).
	PreProcessWrongWord func(word string) string
	// CleanSuggestion ports cleanSuggestion (default: identity).
	CleanSuggestion func(s string) string
	// MatchesDesiredPostag optional override; when nil, Tag + MatchesPosTagRegex is used.
	MatchesDesiredPostag func(suggestion string, desiredPostag string) bool
}

// AcceptRuleMatch ports AbstractFindSuggestionsFilter.acceptRuleMatch.
// Required args: wordFrom, desiredPostag.
// Optional: priorityPostag, removeSuggestionsRegexp, suppressMatch, Mode=diacritics.
func (f *AbstractFindSuggestionsFilter) AcceptRuleMatch(match *RuleMatch, args map[string]string,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	if f.SpellingSuggestions == nil {
		return nil
	}

	wordFrom, ok := args["wordFrom"]
	if !ok {
		panic("Missing key 'wordFrom'")
	}
	desiredPostag, ok := args["desiredPostag"]
	if !ok {
		panic("Missing key 'desiredPostag'")
	}
	priorityPostag := args["priorityPostag"]
	removeSuggestionsRegexp := args["removeSuggestionsRegexp"]
	bSuppressMatch := strings.EqualFold(args["suppressMatch"], "true")
	diacriticsMode := args["Mode"] == "diacritics"
	generateSuggestions := true

	var replacements, replacements2 []string
	var regexpPattern *regexp.Regexp
	usedLemmas := map[string]struct{}{}
	stringComparatorWord := ""

	if wordFrom != "" && desiredPostag != "" {
		var atrWord *languagetool.AnalyzedTokenReadings
		if wordFrom == "inmarker" {
			match.SetOriginalErrorStr()
			pre := match.GetOriginalErrorStr()
			if f.PreProcessWrongWord != nil {
				pre = f.PreProcessWrongWord(pre)
			} else {
				pre = strings.ReplaceAll(pre, " ", "")
			}
			atrWord = languagetool.NewAnalyzedTokenReadings(
				languagetool.NewAnalyzedToken(pre, nil, nil))
		} else {
			atrWord = resolveWordFromFS(wordFrom, match, patternTokens, tokenPositions)
		}
		if atrWord == nil {
			return nil
		}
		stringComparatorWord = atrWord.GetToken()
		isWordCapitalized := tools.IsCapitalizedWord(atrWord.GetToken())
		isWordAllupper := tools.IsAllUppercase(atrWord.GetToken())

		// Check if the original token meets desiredPostag → diacritics mode drops match
		if f.Tag != nil {
			aOriginal := f.Tag(atrWord.GetToken())
			if aOriginal != nil && aOriginal.MatchesPosTagRegex(desiredPostag) {
				if diacriticsMode {
					return nil
				}
			}
		}

		if generateSuggestions {
			if removeSuggestionsRegexp != "" {
				// Java: Pattern.compile(removeSuggestionsRegexp, Pattern.UNICODE_CASE)
				// UNICODE_CASE alone is still case-sensitive (no CASE_INSENSITIVE) — do not invent (?i).
				// Matcher.matches() → full region via \A(?:…)\z.
				regexpPattern, _ = regexp.Compile(`\A(?:` + removeSuggestionsRegexp + `)\z`)
			}
			suggestions := f.SpellingSuggestions(atrWord)
			usedPriorityPostagPos := 0
			for _, suggestion := range suggestions {
				if len(replacements) >= 2*findSuggestionsMax {
					break
				}
				clean := suggestion
				if f.CleanSuggestion != nil {
					clean = f.CleanSuggestion(suggestion)
				}
				var analyzedList []*languagetool.AnalyzedTokenReadings
				if f.Tag != nil {
					if atr := f.Tag(clean); atr != nil {
						analyzedList = []*languagetool.AnalyzedTokenReadings{atr}
					}
				} else if f.MatchesDesiredPostag != nil {
					// Optional inject when Tag unset (tests); Java always has getTagger().
					if f.MatchesDesiredPostag(suggestion, desiredPostag) && suggestion != atrWord.GetToken() {
						if !containsStrFS(replacements, suggestion) && !containsStrFS(replacements, strings.ToLower(suggestion)) {
							if !diacriticsMode || equalWithoutDiacritics(suggestion, atrWord.GetToken()) {
								if regexpPattern == nil || !regexpPattern.MatchString(suggestion) {
									repl := suggestion
									if isWordAllupper {
										repl = strings.ToUpper(repl)
									}
									if isWordCapitalized {
										repl = tools.UppercaseFirstChar(repl)
									}
									replacements = append(replacements, repl)
								}
							}
						}
					}
					continue
				}
				for _, analyzedSuggestion := range analyzedList {
					if f.IsSuggestionException != nil && f.IsSuggestionException(analyzedSuggestion) {
						continue
					}
					if len(replacements) >= 2*findSuggestionsMax {
						break
					}
					used := false
					if suggestion != atrWord.GetToken() && analyzedSuggestion.MatchesPosTagRegex(desiredPostag) {
						if !containsStrFS(replacements, suggestion) &&
							!containsStrFS(replacements, strings.ToLower(suggestion)) &&
							(!diacriticsMode || equalWithoutDiacritics(suggestion, atrWord.GetToken())) {
							if regexpPattern == nil || !regexpPattern.MatchString(suggestion) {
								replacement := suggestion
								if isWordAllupper {
									replacement = strings.ToUpper(replacement)
								}
								if isWordCapitalized {
									replacement = tools.UppercaseFirstChar(replacement)
								}
								if priorityPostag != "" && analyzedSuggestion.MatchesPosTagRegex(priorityPostag) {
									// insert at usedPriorityPostagPos
									replacements = append(replacements[:usedPriorityPostagPos],
										append([]string{replacement}, replacements[usedPriorityPostagPos:]...)...)
									usedPriorityPostagPos++
									used = true
								} else {
									replacements = append(replacements, replacement)
									used = true
								}
							}
						}
					}
					// synthesizer path — Java accumulates synthesizedSuggestions across readings
					// and re-adds the whole list to replacements2 each reading (bug-for-bug).
					if !used && f.Synthesize != nil {
						var synthesizedSuggestions []string
						for _, at := range analyzedSuggestion.GetReadings() {
							if at == nil {
								continue
							}
							lemma := ""
							if at.GetLemma() != nil {
								lemma = *at.GetLemma()
							}
							if _, ok := usedLemmas[lemma]; ok {
								continue
							}
							usedLemmas[lemma] = struct{}{}
							for _, syn := range f.Synthesize(at, desiredPostag) {
								if !containsStrFS(synthesizedSuggestions, syn) {
									synthesizedSuggestions = append(synthesizedSuggestions, syn)
								}
							}
							for _, replacement := range synthesizedSuggestions {
								if isWordAllupper {
									replacement = strings.ToUpper(replacement)
								}
								if isWordCapitalized {
									replacement = tools.UppercaseFirstChar(replacement)
								}
								replacements2 = append(replacements2, replacement)
							}
						}
					}
				}
			}
		}
	}

	matchContainsSomeFinishedSuggestion := false
	for _, k := range match.GetSuggestedReplacements() {
		if !strings.Contains(strings.ToLower(k), "{suggestion}") {
			matchContainsSomeFinishedSuggestion = true
			break
		}
	}
	if diacriticsMode && len(replacements) == 0 && !matchContainsSomeFinishedSuggestion {
		return nil
	}
	if len(replacements)+len(replacements2) == 0 && bSuppressMatch && !matchContainsSomeFinishedSuggestion {
		return nil
	}

	out := NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.GetShortMessage()
	// preserve type-like fields
	out.IssueType = match.IssueType
	out.CategoryID = match.CategoryID
	out.CategoryName = match.CategoryName

	var definitiveReplacements []string
	replacementsUsed := false
	if generateSuggestions {
		for _, s := range match.GetSuggestedReplacements() {
			if strings.Contains(s, "{suggestion}") || strings.Contains(s, "{Suggestion}") || strings.Contains(s, "{SUGGESTION}") {
				replacementsUsed = true
				for _, s2 := range replacements {
					if len(definitiveReplacements) >= findSuggestionsMax {
						break
					}
					switch {
					case strings.Contains(s, "{suggestion}"):
						if !containsStrFS(definitiveReplacements, s2) {
							definitiveReplacements = append(definitiveReplacements, strings.ReplaceAll(s, "{suggestion}", s2))
						}
					case strings.Contains(s, "{Suggestion}"):
						u := tools.UppercaseFirstChar(s2)
						if !containsStrFS(definitiveReplacements, u) {
							definitiveReplacements = append(definitiveReplacements, strings.ReplaceAll(s, "{Suggestion}", u))
						}
					default: // {SUGGESTION}
						u := strings.ToUpper(s2)
						if !containsStrFS(definitiveReplacements, u) {
							definitiveReplacements = append(definitiveReplacements, strings.ReplaceAll(s, "{SUGGESTION}", u))
						}
					}
				}
			} else {
				if !containsStrFS(definitiveReplacements, s) {
					definitiveReplacements = append(definitiveReplacements, s)
				}
			}
		}
		if !replacementsUsed {
			if len(replacements) == 0 {
				sort.SliceStable(replacements2, func(i, j int) bool {
					return stringComparatorLess(stringComparatorWord, replacements2[i], replacements2[j])
				})
				for _, replacement := range replacements2 {
					if !containsStrFS(replacements, replacement) && !containsStrFS(definitiveReplacements, replacement) {
						replacements = append(replacements, replacement)
					}
				}
			}
			for _, replacement := range replacements {
				if len(definitiveReplacements) >= findSuggestionsMax {
					break
				}
				if !containsStrFS(definitiveReplacements, replacement) {
					definitiveReplacements = append(definitiveReplacements, replacement)
				}
			}
		}
	}
	// remove original error string
	orig := match.GetOriginalErrorStr()
	if orig == "" && match.Sentence != nil {
		text := match.Sentence.GetText()
		if match.FromPos >= 0 && match.ToPos <= len(text) {
			orig = text[match.FromPos:match.ToPos]
		}
	}
	definitiveReplacements = removeStringFS(definitiveReplacements, orig)
	// distinct
	definitiveReplacements = distinctFS(definitiveReplacements)
	if len(definitiveReplacements) > 0 {
		out.SetSuggestedReplacements(definitiveReplacements)
	}
	return out
}

func resolveWordFromFS(wordFrom string, match *RuleMatch, patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *languagetool.AnalyzedTokenReadings {
	if wordFrom == "marker" {
		// getPosition("marker") style
		i := 0
		for i < len(patternTokens) &&
			(patternTokens[i].GetStartPos() < match.GetFromPos() || patternTokens[i].IsSentenceStart()) {
			i++
		}
		i++ // Java getPosition then returns i-1 after increment — wait:
		// getPosition: while...; i++; return i-1 → first token at/after match
		// Actually after while i points to first token with start>=from and not sent start skip,
		// then i++, then return i-1 which is the token after the while stop.
		if i-1 >= 0 && i-1 < len(patternTokens) {
			return patternTokens[i-1]
		}
		return nil
	}
	if wordFrom == "inmarker" {
		return nil // handled by caller
	}
	n, err := strconv.Atoi(wordFrom)
	if err != nil {
		// try getPosition numeric path without skip: Java Integer.parseInt
		return nil
	}
	// Java getPosition for numeric: i = parseInt; return i-1 (1-based)
	// Also getSkipCorrectedReference when used — Abstract uses getPosition only.
	idx := n - 1
	if idx < 0 || idx >= len(patternTokens) {
		// also try skipCorrectedRef for tokenPositions compatibility
		idx = skipCorrectedRef(tokenPositions, n)
	}
	if idx < 0 || idx >= len(patternTokens) {
		return nil
	}
	return patternTokens[idx]
}

// equalWithoutDiacritics ports AbstractFindSuggestionsFilter.equalWithoutDiacritics:
// StringTools.removeDiacritics(s).equalsIgnoreCase(StringTools.removeDiacritics(t)).
// Do not invent a hand-written accent map.
func equalWithoutDiacritics(s, t string) bool {
	return tools.EqualsIgnoreCaseAndDiacritics(s, t)
}

func containsStrFS(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func removeStringFS(list []string, s string) []string {
	if s == "" {
		return list
	}
	var out []string
	for _, x := range list {
		if x != s {
			out = append(out, x)
		}
	}
	return out
}

func distinctFS(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

// stringComparatorLess ports AbstractFindSuggestionsFilter.StringComparator
// (EditDistance Damerau, maxDistance 4; over-max → 2*maxDistance).
// Uses the Java twin EditDistance — not invent plain Levenshtein.
func stringComparatorLess(word, o1, o2 string) bool {
	ed := implementation.NewEditDistance(word, implementation.Damerau)
	const maxDistance = 4
	d1 := ed.Compare(o1, maxDistance)
	d2 := ed.Compare(o2, maxDistance)
	if d1 < 0 {
		d1 = 2 * maxDistance
	}
	if d2 < 0 {
		d2 = 2 * maxDistance
	}
	return d1 < d2
}