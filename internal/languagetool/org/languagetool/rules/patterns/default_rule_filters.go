package patterns

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
)

// default_rule_filters registers Java-named RuleFilter classes the grammar loader
// may attach. Unknown filter classes still skip the rule (fail-closed).

func init() {
	registerDefaultRuleFilters(GlobalRuleFilterCreator)
}

func registerDefaultRuleFilters(c *RuleFilterCreator) {
	if c == nil {
		return
	}
	c.Register("org.languagetool.rules.patterns.ApostropheTypeFilter", func() RuleFilter {
		return ApostropheTypeFilter{}
	})
	// RegexAntiPatternFilter is a RegexRuleFilter in Java; also used on pattern rules in EN grammar.
	c.Register("org.languagetool.rules.patterns.RegexAntiPatternFilter", func() RuleFilter {
		return regexAntiPatternAsRuleFilter{}
	})
	c.Register("org.languagetool.rules.UnderlineSpacesFilter", func() RuleFilter {
		return underlineSpacesRuleFilter{}
	})
	// MultitokenSpellerFilter needs Language.getMultitokenSpeller(); without it
	// Java drops the match (empty suggestions). Same fail-closed until wired.
	c.Register("org.languagetool.rules.spelling.multitoken.MultitokenSpellerFilter", func() RuleFilter {
		return multitokenSpellerRuleFilter{}
	})
}

// regexAntiPatternAsRuleFilter adapts RegexAntiPatternFilter to RuleFilter (pattern rules).
type regexAntiPatternAsRuleFilter struct{}

func (regexAntiPatternAsRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	sent := (*languagetool.AnalyzedSentence)(nil)
	if match != nil {
		sent = match.Sentence
	}
	return (RegexAntiPatternFilter{}).AcceptRegexMatch(match, arguments, sent)
}

// underlineSpacesRuleFilter adapts rules.UnderlineSpacesFilter to RuleFilter.
type underlineSpacesRuleFilter struct{}

func (underlineSpacesRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	mode := GetRequired("underlineSpaces", arguments)
	text := ""
	if match.Sentence != nil {
		text = match.Sentence.GetText()
	}
	from, to := rules.NewUnderlineSpacesFilter().Expand(text, match.FromPos, match.ToPos, mode)
	match.SetOffsetPosition(from, to)
	return match
}

var (
	multitokenSpellerMu      sync.RWMutex
	defaultMultitokenSpeller *multitoken.MultitokenSpeller
	defaultIsMisspelled      func(string) bool
)

// SetDefaultMultitokenSpeller wires MultitokenSpellerFilter's dictionary backend
// (Java: Language.getMultitokenSpeller). isMisspelled may be nil.
func SetDefaultMultitokenSpeller(sp *multitoken.MultitokenSpeller, isMisspelled func(string) bool) {
	multitokenSpellerMu.Lock()
	defer multitokenSpellerMu.Unlock()
	defaultMultitokenSpeller = sp
	defaultIsMisspelled = isMisspelled
}

// multitokenSpellerRuleFilter ports MultitokenSpellerFilter onto the optional default speller.
type multitokenSpellerRuleFilter struct{}

func (multitokenSpellerRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	// Java: all pattern tokens ignored by speller → drop.
	if len(patternTokens) > 0 {
		allIgn := true
		for _, t := range patternTokens {
			if t == nil || !t.IsIgnoredBySpeller() {
				allIgn = false
				break
			}
		}
		if allIgn {
			return nil
		}
	}
	multitokenSpellerMu.RLock()
	sp := defaultMultitokenSpeller
	isMiss := defaultIsMisspelled
	multitokenSpellerMu.RUnlock()
	if sp == nil {
		// Empty suggestions → Java returns null (do not invent replacements).
		return nil
	}
	inner := &multitoken.MultitokenSpellerFilter{Speller: sp, IsMisspelled: isMiss}
	// Sentence-start capitalisation uses patternTokenPos vs first content token in Java;
	// leave AtSentenceStart false until full StringTools port is wired (incomplete, not invent).
	_ = patternTokenPos
	original := ""
	if match.Sentence != nil {
		text := match.Sentence.GetText()
		if match.FromPos >= 0 && match.ToPos <= len(text) && match.FromPos < match.ToPos {
			original = text[match.FromPos:match.ToPos]
		}
	}
	return inner.AcceptRuleMatch(match, original)
}
