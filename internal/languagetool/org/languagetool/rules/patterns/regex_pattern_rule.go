package patterns

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// MaxSentLength ports RegexPatternRule.MAX_SENT_LENGTH (Java String.length = UTF-16 units).
const MaxSentLength = 2000

// matchesInSuggestionsNumberedFrom ports MATCHES_IN_SUGGESTIONS_NUMBERED_FROM (0).
const matchesInSuggestionsNumberedFrom = 0

var (
	suggestionPatternRE = regexp.MustCompile(`(?s)<suggestion>(.*?)</suggestion>`)
	// Java matchPattern = "\\\\\\d" — single digit backref only.
	matchPatternRE = regexp.MustCompile(`\\\d`)
)

// RegexPatternRule ports org.languagetool.rules.patterns.RegexPatternRule.
type RegexPatternRule struct {
	*AbstractPatternRule
	Pattern            *regexp.Regexp
	MarkGroup          int
	RegexFilter        RegexRuleFilter
	RequiredSubstrings *Substrings
	CaseSensitive      bool
}

// NewRegexPatternRule constructs a regexp-based pattern rule.
// pattern is a Go RE2 expression (matching uses FindStringSubmatchIndex).
// Ports RegexPatternRule ctor: requiredSubstrings from pattern, caseSensitive default true.
func NewRegexPatternRule(id, description, message, shortMessage, suggestionsOutMsg, languageCode string, pattern *regexp.Regexp, regexpMark int) *RegexPatternRule {
	base := NewAbstractPatternRule(id, description, languageCode, nil, false)
	base.Message = message
	base.ShortMessage = shortMessage
	if shortMessage == "" {
		base.ShortMessage = ""
	}
	base.SuggestionsOutMsg = suggestionsOutMsg
	r := &RegexPatternRule{
		AbstractPatternRule: base,
		Pattern:             pattern,
		MarkGroup:           regexpMark,
		CaseSensitive:       true,
	}
	if pattern != nil {
		r.RequiredSubstrings = GetRequiredSubstrings(pattern.String())
	}
	return r
}

func (r *RegexPatternRule) SetRegexFilter(f RegexRuleFilter) { r.RegexFilter = f }

// SetCaseSensitive ports caseSensitive when the Java Pattern had CASE_INSENSITIVE.
func (r *RegexPatternRule) SetCaseSensitive(v bool) {
	if r != nil {
		r.CaseSensitive = v
	}
}

// Match ports RegexPatternRule.match.
func (r *RegexPatternRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if sentence == nil || r.Pattern == nil {
		return nil, nil
	}
	text := sentence.GetText()
	startPos := 0
	if r.RequiredSubstrings != nil {
		startPos = r.RequiredSubstrings.Find(text, r.CaseSensitive)
		if startPos < 0 {
			return nil, nil
		}
		// Java: requiredSubstrings != null && !mustStart ? 0 : startPos
		if !r.RequiredSubstrings.MustStart {
			startPos = 0
		}
	}
	return r.doMatch(sentence, text, startPos), nil
}

func (r *RegexPatternRule) doMatch(sentence *languagetool.AnalyzedSentence, text string, startPos int) []*rules.RuleMatch {
	var matches []*rules.RuleMatch
	// Java: if (text.length() > MAX_SENT_LENGTH) — UTF-16 code units.
	if tokenizers.UTF16Len(text) > MaxSentLength {
		return matches
	}
	// Precompute backref / suggestion spans (Java getClausePositionsInMessage).
	sugInMsg := clausePositions(suggestionPatternRE, r.Message)
	brefInMsg := clausePositions(matchPatternRE, r.Message)
	sugInOut := clausePositions(suggestionPatternRE, r.SuggestionsOutMsg)
	brefInOut := clausePositions(matchPatternRE, r.SuggestionsOutMsg)

	// scan from startPos (byte index into Go string; Java find uses UTF-16 — ASCII-safe for LT rules)
	search := text
	offset := 0
	if startPos > 0 && startPos <= len(text) {
		search = text[startPos:]
		offset = startPos
	}
	// Java while (find(startPos)) { … startPos = end(); } — non-overlapping left-to-right
	all := r.Pattern.FindAllStringSubmatchIndex(search, -1)
	for _, loc := range all {
		if len(loc) < 2 {
			continue
		}
		fullStartB, fullEndB := loc[0]+offset, loc[1]+offset
		markStartB, markEndB := fullStartB, fullEndB
		if r.MarkGroup >= 0 && 2*r.MarkGroup+1 < len(loc) && loc[2*r.MarkGroup] >= 0 {
			markStartB = loc[2*r.MarkGroup] + offset
			markEndB = loc[2*r.MarkGroup+1] + offset
		}
		groups := make([]string, len(loc)/2)
		for g := 0; g < len(groups); g++ {
			if loc[2*g] >= 0 {
				groups[g] = search[loc[2*g]:loc[2*g+1]]
			}
		}
		// Java processMessage with Match list for regexReplace + case conversion.
		var sugMatches, sugMatchesOut []*Match
		if r.AbstractPatternRule != nil {
			sugMatches = r.SuggestionMatches
			sugMatchesOut = r.SuggestionMatchesOutMsg
		}
		processedMessage := processRegexMessage(r.Message, groups, brefInMsg, sugInMsg, sugMatches, r.LanguageCode)
		processedSugOut := processRegexMessage(r.SuggestionsOutMsg, groups, brefInOut, sugInOut, sugMatchesOut, r.LanguageCode)

		// Positions are UTF-16 (Java String indices / RuleMatch).
		markStart := byteIndexToUTF16(text, markStartB)
		markEnd := byteIndexToUTF16(text, markEndB)
		patStart := byteIndexToUTF16(text, fullStartB)
		patEnd := byteIndexToUTF16(text, fullEndB)

		// startsWithUpperCase fed into Java RuleMatch for suggestion casing.
		startsWithUpperCase := fullStartB == 0 && len(text) > 0 && firstRuneIsUpper(text)
		_ = startsWithUpperCase // used when RuleMatch gains AdjustSuggestionsCase; keep Java check

		rm := rules.NewRuleMatch(r, sentence, markStart, markEnd, processedMessage)
		rm.ShortMessage = r.ShortMessage
		rm.SetPatternPosition(patStart, patEnd)
		// suggestions from message + suggestionsOutMsg bodies
		var sug []string
		for _, m := range suggestionPatternRE.FindAllStringSubmatch(processedMessage, -1) {
			if len(m) > 1 {
				s := m[1]
				if startsWithUpperCase {
					s = ConvertCase(CaseStartUpper, s, s)
				}
				sug = append(sug, s)
			}
		}
		for _, m := range suggestionPatternRE.FindAllStringSubmatch(processedSugOut, -1) {
			if len(m) > 1 {
				s := m[1]
				if startsWithUpperCase {
					s = ConvertCase(CaseStartUpper, s, s)
				}
				sug = append(sug, s)
			}
		}
		rm.SetSuggestedReplacements(sug)

		if r.RegexFilter != nil {
			eval := NewRegexRuleFilterEvaluator(r.RegexFilter)
			filtered := eval.RunFilter(r.FilterArgs, rm, sentence, groups)
			if filtered != nil {
				matches = append(matches, rm)
			}
		} else {
			matches = append(matches, rm)
		}
	}
	return matches
}

type intPair struct{ left, right int }

func clausePositions(re *regexp.Regexp, message string) []intPair {
	if message == "" || re == nil {
		return nil
	}
	var out []intPair
	for _, loc := range re.FindAllStringIndex(message, -1) {
		if len(loc) >= 2 {
			out = append(out, intPair{loc[0], loc[1]})
		}
	}
	return out
}

// processRegexMessage ports RegexPatternRule.processMessage —
// backrefs with optional Match.regexReplace + case conversion.
// When matches is shorter than backrefs (common when no <match> tags), use bare groups.
func processRegexMessage(message string, groups []string, backRefs, suggestions []intPair, matches []*Match, langCode string) string {
	if message == "" {
		return message
	}
	if len(backRefs) == 0 {
		return message
	}
	closestSug := 0
	allSugPassed := len(suggestions) == 0
	var b strings.Builder
	startOfProcessing := 0
	for i, reference := range backRefs {
		for !allSugPassed && reference.left > suggestions[closestSug].right {
			closestSug++
			if closestSug == len(suggestions) {
				allSugPassed = true
			}
		}
		insideSuggestion := !allSugPassed && reference.left >= suggestions[closestSug].left

		refStr := message[reference.left:reference.right]
		nStr := strings.TrimPrefix(refStr, `\`)
		var inXML int
		fmt.Sscanf(nStr, "%d", &inXML)
		actual := inXML
		if insideSuggestion {
			actual = inXML - matchesInSuggestionsNumberedFrom
		}
		val := ""
		if actual >= 0 && actual < len(groups) {
			val = groups[actual]
		}

		suggestion := val
		if i < len(matches) && matches[i] != nil {
			m := matches[i]
			if m.RegexReplace != "" {
				if re := m.GetRegexMatch(); re != nil {
					// Java Matcher.replaceFirst
					suggestion = re.ReplaceAllString(val, m.RegexReplace)
				}
				suggestion = ConvertCaseLang(m.CaseConversionType, suggestion, val, langCode)
			}
		}
		b.WriteString(message[startOfProcessing:reference.left])
		b.WriteString(suggestion)
		startOfProcessing = reference.right
	}
	b.WriteString(message[startOfProcessing:])
	return b.String()
}

func firstRuneIsUpper(s string) bool {
	for _, r := range s {
		return unicode.IsUpper(r)
	}
	return false
}

// byteIndexToUTF16 converts a Go byte offset into UTF-16 code unit offset (Java String index).
func byteIndexToUTF16(s string, byteIdx int) int {
	if byteIdx <= 0 {
		return 0
	}
	if byteIdx >= len(s) {
		return tokenizers.UTF16Len(s)
	}
	return len(utf16.Encode([]rune(s[:byteIdx])))
}

// String describes the rule pattern.
func (r *RegexPatternRule) String() string {
	if r.Pattern == nil {
		return ""
	}
	return r.Pattern.String() + "/flags:0"
}

// EstimateContextForSureMatch ports estimateContextForSureMatch.
func (r *RegexPatternRule) EstimateContextForSureMatch() int { return -1 }
