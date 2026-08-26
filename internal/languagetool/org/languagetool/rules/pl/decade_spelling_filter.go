package pl

import (
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// DecadeSpellingFilter ports org.languagetool.rules.pl.DecadeSpellingFilter.
// Used by DATA_DEKADY grammar rules (args lata:…).
type DecadeSpellingFilter struct{}

func NewDecadeSpellingFilter() *DecadeSpellingFilter {
	return &DecadeSpellingFilter{}
}

// Roman number tables — same order/values as Java DecadeSpellingFilter.numbers/letters.
var (
	decadeSpellingNumbers = []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	decadeSpellingLetters = []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}
)

// AcceptRuleMatch ports DecadeSpellingFilter.acceptRuleMatch.
//
// Java control flow (bug-for-bug):
//   - arguments.get("lata") null → NPE (propagates; not IllegalArgumentException)
//   - substring on short lata → StringIndexOutOfBoundsException (propagates; not IAE)
//   - Integer.parseInt fail → NumberFormatException extends IllegalArgumentException → null
//   - RuleMatch ctor toPos<=fromPos → IllegalArgumentException → null
//   - RuleMatch ctor null rule/message → NPE (propagates)
//   - new RuleMatch(..., patternFrom/To, message, shortMessage, fromPos==0, null); setType
func (f *DecadeSpellingFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	// Java: arguments.get("lata") — null map or missing key → NPE on .substring
	if arguments == nil {
		panic("java.lang.NullPointerException: arguments")
	}
	lata, ok := arguments["lata"]
	if !ok {
		// Java Map.get missing → null; null.substring → NPE
		panic("java.lang.NullPointerException: lata")
	}
	// Java String.substring is UTF-16 code units; DATA_DEKADY years are ASCII digits.
	// length < 2: substring(2) throws StringIndexOutOfBoundsException (NOT IAE → not swallowed).
	if len(lata) < 2 {
		panic("java.lang.StringIndexOutOfBoundsException: lata length < 2")
	}
	decade := lata[2:]
	century := lata[:2]
	cent, err := strconv.Atoi(century)
	if err != nil {
		// Java Integer.parseInt → NumberFormatException extends IllegalArgumentException → return null
		return nil
	}
	// Java: match.getMessage() — null match → NPE
	if match == nil {
		panic("java.lang.NullPointerException: match")
	}
	// Java String.replace(CharSequence, CharSequence) replaces all non-overlapping occurrences.
	msg := strings.ReplaceAll(match.GetMessage(), "{dekada}", decade)
	msg = strings.ReplaceAll(msg, "{wiek}", getRomanNumber(cent+1))

	fromPos := match.GetFromPos()
	toPos := match.GetToPos()
	// Java RuleMatch ctor: toPos <= fromPos → IllegalArgumentException → swallowed → null
	if toPos <= fromPos {
		return nil
	}
	// Java Objects.requireNonNull(rule) → NPE if null (not swallowed)
	if match.GetRule() == nil {
		panic("java.lang.NullPointerException: rule")
	}
	// Java Objects.requireNonNull(message) — msg is never null after ReplaceAll on non-null GetMessage

	// Java:
	// new RuleMatch(rule, sentence, from, to, patternFrom, patternTo, message, shortMessage,
	//               match.getFromPos() == 0, null)
	//   → isAllUppercase=false, suggestionsOutMsg=null, setOriginalErrorStr=false
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, fromPos, toPos, msg)
	out.SetPatternPosition(match.GetPatternFromPos(), match.GetPatternToPos())
	out.ShortMessage = match.GetShortMessage()
	// Extract <suggestion> bodies like Java RuleMatch ctor (startWithUppercase = fromPos==0).
	decadeExtractSuggestions(out, msg, fromPos == 0)
	// Java: ruleMatch.setType(match.getType())
	out.SetType(match.GetType())
	return out
}

// getRomanNumber ports DecadeSpellingFilter.getRomanNumber.
func getRomanNumber(num int) string {
	var roman strings.Builder
	n := num
	for i := 0; i < len(decadeSpellingNumbers); i++ {
		for n >= decadeSpellingNumbers[i] {
			roman.WriteString(decadeSpellingLetters[i])
			n -= decadeSpellingNumbers[i]
		}
	}
	return roman.String()
}

// decadeExtractSuggestions ports the suggestion extraction inside Java RuleMatch's
// 12-arg constructor used by DecadeSpellingFilter (isAllUppercase=false, suggestionsOutMsg=null).
func decadeExtractSuggestions(rm *rules.RuleMatch, message string, startWithUppercase bool) {
	const (
		sugStart = "<suggestion>"
		sugEnd   = "</suggestion>"
		mistake  = "<mistake/>" // PatternRuleMatcher.MISTAKE
	)
	// Java LinkedHashSet order + uniqueness by replacement string.
	seen := map[string]struct{}{}
	var reps []string
	s := message // suggestionsOutMsg is null in the DecadeSpellingFilter call
	for {
		pos := strings.Index(s, sugStart)
		if pos < 0 {
			break
		}
		rest := s[pos+len(sugStart):]
		end := strings.Index(rest, sugEnd)
		if end < 0 {
			break
		}
		replacement := rest[:end]
		s = rest[end+len(sugEnd):]
		if strings.Contains(replacement, mistake) {
			continue
		}
		if startWithUppercase {
			replacement = tools.UppercaseFirstChar(replacement)
		}
		if _, ok := seen[replacement]; ok {
			continue
		}
		seen[replacement] = struct{}{}
		reps = append(reps, replacement)
	}
	if len(reps) > 0 {
		rm.SetSuggestedReplacements(reps)
	}
}
