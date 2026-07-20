package rules

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// CheckPostagsInSuggestionFilter ports org.languagetool.rules.CheckPostagsInSuggestionFilter.
// TagToken returns POS tags for a single token (Java Tagger.tag → readings POS).
// Nil tagger: Java throws IOException; Go panics with the same intent.
type CheckPostagsInSuggestionFilter struct {
	TagToken func(token string) []string
}

func NewCheckPostagsInSuggestionFilter(tag func(string) []string) *CheckPostagsInSuggestionFilter {
	return &CheckPostagsInSuggestionFilter{TagToken: tag}
}

var (
	checkPostagsTagMu      sync.RWMutex
	defaultCheckPostagsTag func(string) []string
	// javaSplitWS ports String.split("\\s+") (limit 0: trailing empties discarded).
	javaSplitWS = regexp.MustCompile(`\s+`)
)

// SetDefaultCheckPostagsTagger wires language tagger for CheckPostagsInSuggestionFilter
// (Java: Language.getTagger()).
func SetDefaultCheckPostagsTagger(tag func(string) []string) {
	checkPostagsTagMu.Lock()
	defer checkPostagsTagMu.Unlock()
	defaultCheckPostagsTag = tag
}

// AcceptRuleMatch ports CheckPostagsInSuggestionFilter.acceptRuleMatch.
func (f *CheckPostagsInSuggestionFilter) AcceptRuleMatch(match *RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if match == nil {
		return nil
	}
	tag := f.TagToken
	if tag == nil {
		checkPostagsTagMu.RLock()
		tag = defaultCheckPostagsTag
		checkPostagsTagMu.RUnlock()
	}
	if tag == nil {
		// Java: throw new IOException("Language tagger not available in rule …")
		panic("Language tagger not available in CheckPostagsInSuggestionFilter")
	}
	postagsListStr, ok := arguments["PostagsList"]
	if !ok {
		panic("Missing key 'PostagsList'")
	}
	filtered := (&CheckPostagsInSuggestionFilter{TagToken: tag}).Filter(match.GetSuggestedReplacements(), postagsListStr)
	if len(filtered) == 0 {
		return nil
	}
	match.SetSuggestedReplacements(filtered)
	return match
}

// Filter keeps suggestions whose tokens match postagsList (comma-separated regexes).
// Java: replacement.split("\\s+"); tagger.tag; matchesPosTagRegex; empty list → null match.
// Token/tag count mismatch throws (do not invent skip).
func (f *CheckPostagsInSuggestionFilter) Filter(replacements []string, postagsListStr string) []string {
	if f.TagToken == nil {
		panic("Language tagger not available in CheckPostagsInSuggestionFilter")
	}
	// Java: postagsListStr.split(",") — no invent TrimSpace on each tag regex.
	postagsList := strings.Split(postagsListStr, ",")
	var out []string
	for _, replacement := range replacements {
		tokens := javaSplitWhitespace(replacement)
		if len(tokens) != len(postagsList) || len(postagsList) == 0 {
			// Java throws IOException — panic for twin (do not invent continue).
			panic(fmt.Sprintf("Mismatch between number of tokens and number of tags: %v vs %v", tokens, postagsList))
		}
		postagsMatch := true
		for i, tok := range tokens {
			if !tokenMatchesPosTagRegex(f.TagToken(tok), postagsList[i]) {
				postagsMatch = false
				break
			}
		}
		if postagsMatch {
			out = append(out, replacement)
		}
	}
	return out
}

// javaSplitWhitespace ports Java String.split("\\s+") with default limit 0
// (trailing empty strings discarded; leading empty kept if string starts with WS).
func javaSplitWhitespace(s string) []string {
	parts := javaSplitWS.Split(s, -1)
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

// tokenMatchesPosTagRegex ports AnalyzedTokenReadings.matchesPosTagRegex:
// Pattern.matcher(posTag).matches() on any reading (full region).
func tokenMatchesPosTagRegex(tags []string, posTagRegex string) bool {
	re, err := regexp.Compile(`\A(?:` + posTagRegex + `)\z`)
	if err != nil {
		// Java Pattern.compile throws; treat as no match for this token
		return false
	}
	for _, tag := range tags {
		if re.MatchString(tag) {
			return true
		}
	}
	return false
}
