package filters

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ArabicNumberPhraseFilter ports org.languagetool.rules.ar.filters.ArabicNumberPhraseFilter.
// Full unit inflection via ArabicSynthesizer is incomplete; unit forms use tools helpers.
type ArabicNumberPhraseFilter struct{}

func NewArabicNumberPhraseFilter() *ArabicNumberPhraseFilter {
	return &ArabicNumberPhraseFilter{}
}

// AcceptRuleMatch ports ArabicNumberPhraseFilter.acceptRuleMatch.
func (f *ArabicNumberPhraseFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	previousWord := arguments["previous"]
	inflectArg := arguments["inflect"]
	nextWordArg := arguments["next"]
	previousWordPos := previousPosFromArgs(arguments)
	nextWordPos := nextPosFromArgs(arguments, len(patternTokens))

	// Java: startPos = (previousWordPos > 0) ? previousWordPos + 1 : 0
	startPos := 0
	if previousWordPos > 0 {
		startPos = previousWordPos + 1
	}
	// Java: endPos = (nextWordPos > 0) ? min(nextWordPos, length) : length + nextWordPos
	// Note: getNextPos already converts negative to size+n, so nextWordPos is absolute.
	endPos := len(patternTokens)
	if nextWordPos > 0 {
		if nextWordPos < endPos {
			endPos = nextWordPos
		}
	}

	var numWordTokens []string
	for i := startPos; i < endPos && i < len(patternTokens); i++ {
		if patternTokens[i] == nil {
			continue
		}
		numWordTokens = append(numWordTokens, strings.TrimSpace(patternTokens[i].GetToken()))
	}
	numPhrase := strings.Join(numWordTokens, " ")
	feminine := false
	inflection := inflectedCase(patternTokens, previousWordPos, inflectArg)

	// Java: if nextWord.isEmpty() use prepareSuggestion; else prepareSuggestionWithUnits.
	// nextWord arg may be empty while nextPos still points at a unit token (grammar: nextPos:-1).
	var suggestionList []string
	useUnits := nextWordArg != ""
	unitSurface := nextWordArg
	if !useUnits && nextWordPos > 0 && nextWordPos < len(patternTokens) && patternTokens[nextWordPos] != nil {
		// Token after numeric span is the unit (Java nextWordToken = patternTokens[endPos]).
		useUnits = true
		unitSurface = patternTokens[nextWordPos].GetToken()
	}
	if useUnits {
		suggestionList = PrepareSuggestionWithUnit(numPhrase, previousWord, unitSurface, inflection, feminine)
	} else {
		suggestionList = PrepareSuggestion(numPhrase, previousWord, feminine)
	}

	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	if len(suggestionList) > 0 {
		out.SetSuggestedReplacements(suggestionList)
	}
	return out
}

// PrepareSuggestion builds suggestions for a numeric phrase (digit or known).
// previousWord is optionally prefixed.
func PrepareSuggestion(numPhrase, previousWord string, feminine bool) []string {
	sug := SuggestionsForNumericPhrase(numPhrase, feminine)
	if len(sug) == 0 {
		return nil
	}
	out := make([]string, 0, len(sug))
	for _, s := range sug {
		if previousWord != "" {
			out = append(out, previousWord+" "+s)
		} else {
			out = append(out, s)
		}
	}
	return out
}

// PrepareSuggestionWithUnit appends a unit form after the numeric phrase.
func PrepareSuggestionWithUnit(numPhrase, previousWord, unit, inflection string, feminine bool) []string {
	base := PrepareSuggestion(numPhrase, previousWord, feminine)
	if unit == "" {
		return base
	}
	unitForm := tools.GetArabicUnitOneForm(unit, orDefault(inflection, "raf3"))
	if n, err := parseLeadingInt(numPhrase); err == nil {
		if n == 2 {
			unitForm = tools.GetArabicUnitTwoForm(unit, orDefault(inflection, "raf3"))
		} else if n >= 3 && n <= 10 {
			unitForm = tools.GetArabicUnitPluralForm(unit, orDefault(inflection, "raf3"))
		}
	}
	out := make([]string, 0, len(base))
	for _, s := range base {
		out = append(out, s+" "+unitForm)
	}
	return out
}

// SuggestionsForNumericPhrase converts a phrase of digits to Arabic words.
func SuggestionsForNumericPhrase(numPhrase string, feminine bool) []string {
	numPhrase = strings.TrimSpace(numPhrase)
	if numPhrase == "" {
		return nil
	}
	if isAllDigits(numPhrase) {
		w := tools.NumberToArabicWordsGender(numPhrase, feminine)
		if w == "" {
			return nil
		}
		return []string{w}
	}
	for _, tok := range strings.Fields(numPhrase) {
		if isAllDigits(tok) {
			w := tools.NumberToArabicWordsGender(tok, feminine)
			if w != "" {
				return []string{w}
			}
		}
	}
	return nil
}

// InflectionFromPrevious returns "jar" when previous token starts with ب/ل/ك.
func InflectionFromPrevious(previousWord string) string {
	if previousWord == "" {
		return ""
	}
	r := []rune(previousWord)
	if len(r) == 0 {
		return ""
	}
	switch r[0] {
	case 'ب', 'ل', 'ك':
		return "jar"
	}
	return ""
}

func previousPosFromArgs(args map[string]string) int {
	s := args["previousPos"]
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	return n - 1
}

func nextPosFromArgs(args map[string]string, size int) int {
	s := args["nextPos"]
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	if n < 0 {
		return size + n
	}
	return n
}

func inflectedCase(patternTokens []*languagetool.AnalyzedTokenReadings, previousPos int, inflect string) string {
	if inflect != "" {
		return inflect
	}
	if previousPos >= 0 && previousPos < len(patternTokens) && patternTokens[previousPos] != nil {
		for _, tk := range patternTokens[previousPos].GetReadings() {
			if tk == nil || tk.GetPOSTag() == nil {
				continue
			}
			if strings.HasPrefix(*tk.GetPOSTag(), "PR") {
				return "jar"
			}
		}
	}
	if previousPos+1 >= 0 && previousPos+1 < len(patternTokens) && patternTokens[previousPos+1] != nil {
		first := patternTokens[previousPos+1].GetToken()
		if strings.HasPrefix(first, "ب") || strings.HasPrefix(first, "ل") || strings.HasPrefix(first, "ك") {
			return "jar"
		}
	}
	return ""
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func parseLeadingInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if b.Len() > 0 {
			break
		}
	}
	return strconv.Atoi(b.String())
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}
