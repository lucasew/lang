package patterns

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// MaxSentLength ports RegexPatternRule.MAX_SENT_LENGTH.
const MaxSentLength = 2000

var (
	suggestionPatternRE = regexp.MustCompile(`(?s)<suggestion>(.*?)</suggestion>`)
	matchPatternRE      = regexp.MustCompile(`\\\d+`)
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
// pattern is a Go RE2 expression (anchoring is caller's responsibility; matching uses FindStringSubmatchIndex).
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
	return r
}

func (r *RegexPatternRule) SetRegexFilter(f RegexRuleFilter) { r.RegexFilter = f }

// Match ports RegexPatternRule.match.
func (r *RegexPatternRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if sentence == nil || r.Pattern == nil {
		return nil, nil
	}
	text := sentence.GetText()
	if utf8.RuneCountInString(text) > MaxSentLength || len(text) > MaxSentLength {
		// Java uses text.length() (UTF-16 units); approximate with rune/byte length.
		return nil, nil
	}
	startPos := 0
	if r.RequiredSubstrings != nil {
		startPos = r.RequiredSubstrings.Find(text, r.CaseSensitive)
		if startPos < 0 {
			return nil, nil
		}
		if r.RequiredSubstrings.MustStart {
			// keep startPos
		} else {
			startPos = 0
		}
	}
	return r.doMatch(sentence, text, startPos), nil
}

func (r *RegexPatternRule) doMatch(sentence *languagetool.AnalyzedSentence, text string, startPos int) []*rules.RuleMatch {
	var matches []*rules.RuleMatch
	// scan from startPos
	search := text
	offset := 0
	if startPos > 0 && startPos <= len(text) {
		search = text[startPos:]
		offset = startPos
	}
	all := r.Pattern.FindAllStringSubmatchIndex(search, -1)
	for _, loc := range all {
		if len(loc) < 2 {
			continue
		}
		fullStart, fullEnd := loc[0]+offset, loc[1]+offset
		markStart, markEnd := fullStart, fullEnd
		// group N is at indices 2*N, 2*N+1
		if r.MarkGroup >= 0 && 2*r.MarkGroup+1 < len(loc) && loc[2*r.MarkGroup] >= 0 {
			markStart = loc[2*r.MarkGroup] + offset
			markEnd = loc[2*r.MarkGroup+1] + offset
		}
		// extract groups for message processing
		groups := make([]string, len(loc)/2)
		for g := 0; g < len(groups); g++ {
			if loc[2*g] >= 0 {
				groups[g] = search[loc[2*g]:loc[2*g+1]]
			}
		}
		processedMessage := processRegexMessage(r.Message, groups)
		processedSugOut := processRegexMessage(r.SuggestionsOutMsg, groups)
		startsWithUpperCase := fullStart == 0 && len(text) > 0 && unicode.IsUpper([]rune(text)[0])
		_ = startsWithUpperCase

		rm := rules.NewRuleMatch(r, sentence, markStart, markEnd, processedMessage)
		rm.ShortMessage = r.ShortMessage
		// suggestions from message + suggestionsOutMsg
		var sug []string
		for _, m := range suggestionPatternRE.FindAllStringSubmatch(processedMessage, -1) {
			if len(m) > 1 {
				sug = append(sug, m[1])
			}
		}
		for _, m := range suggestionPatternRE.FindAllStringSubmatch(processedSugOut, -1) {
			if len(m) > 1 {
				sug = append(sug, m[1])
			}
		}
		rm.SetSuggestedReplacements(sug)

		if r.RegexFilter != nil {
			eval := NewRegexRuleFilterEvaluator(r.RegexFilter)
			// groups[0]=full match, groups[1…]=captures (Java Matcher.group).
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

// processRegexMessage replaces \N backrefs (Java-style) with capture groups.
func processRegexMessage(message string, groups []string) string {
	if message == "" {
		return message
	}
	return matchPatternRE.ReplaceAllStringFunc(message, func(ref string) string {
		// ref is like \1
		nStr := strings.TrimPrefix(ref, `\`)
		var n int
		fmt.Sscanf(nStr, "%d", &n)
		if n >= 0 && n < len(groups) {
			return groups[n]
		}
		return ""
	})
}

// String describes the rule pattern.
func (r *RegexPatternRule) String() string {
	if r.Pattern == nil {
		return ""
	}
	return r.Pattern.String()
}

// EstimateContextForSureMatch ports estimateContextForSureMatch.
func (r *RegexPatternRule) EstimateContextForSureMatch() int { return -1 }
