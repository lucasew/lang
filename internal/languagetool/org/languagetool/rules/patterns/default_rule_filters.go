package patterns

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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
	// DateRangeChecker is language-agnostic (x/y integer args).
	c.Register("org.languagetool.rules.DateRangeChecker", func() RuleFilter {
		return dateRangeCheckerRuleFilter{}
	})
	// ShortenedYearRangeChecker: two-digit end year under x's century (PL/PT grammars).
	c.Register("org.languagetool.rules.ShortenedYearRangeChecker", func() RuleFilter {
		return shortenedYearRangeCheckerRuleFilter{}
	})
	// WhitespaceCheckFilter: exact preceding whitespace char (DE/CA grammars).
	c.Register("org.languagetool.rules.WhitespaceCheckFilter", func() RuleFilter {
		return whitespaceCheckRuleFilter{}
	})
	c.Register("org.languagetool.rules.ConvertToSentenceCaseFilter", func() RuleFilter {
		return convertToSentenceCaseRuleFilter{inner: rules.NewConvertToSentenceCaseFilter()}
	})
	// Core filters used across CA/ES/PT/… grammars (Java languagetool-core).
	c.Register("org.languagetool.rules.AdaptSuggestionsFilter", func() RuleFilter {
		return adaptSuggestionsRuleFilter{inner: rules.NewAdaptSuggestionsFilter(nil)}
	})
	c.Register("org.languagetool.rules.AddCommasFilter", func() RuleFilter {
		return addCommasRuleFilter{}
	})
	c.Register("org.languagetool.rules.IsEnglishWordFilter", func() RuleFilter {
		return isEnglishWordRuleFilter{}
	})
	c.Register("org.languagetool.rules.CheckPostagsInSuggestionFilter", func() RuleFilter {
		return checkPostagsRuleFilter{}
	})
	c.Register("org.languagetool.rules.SuppressIfAnyRuleMatchesFilter", func() RuleFilter {
		return suppressIfAnyRuleFilter{}
	})
	// Demo partial POS filter (xx grammar); tags only "accurate" as JJ (Java twin).
	c.Register("org.languagetool.rules.DemoPartialPosTagFilter", func() RuleFilter {
		return demoPartialPosTagRuleFilter{}
	})
}

// demoPartialPosTagRuleFilter ports DemoPartialPosTagFilter.
type demoPartialPosTagRuleFilter struct{}

func (demoPartialPosTagRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, pos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	f := rules.NewPartialPosTagFilter(func(partial string) []string {
		// Java: if ("accurate".equals(token)) → JJ; else null.
		if partial == "accurate" {
			return []string{"JJ"}
		}
		return nil
	})
	return f.AcceptRuleMatch(match, arguments, pos, patternTokens, tokenPositions)
}

// shortenedYearRangeCheckerRuleFilter ports ShortenedYearRangeChecker.
type shortenedYearRangeCheckerRuleFilter struct{}

func (shortenedYearRangeCheckerRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	return rules.NewShortenedYearRangeChecker().AcceptRuleMatch(match, arguments, 0, nil, nil)
}

// whitespaceCheckRuleFilter ports WhitespaceCheckFilter.
type whitespaceCheckRuleFilter struct{}

func (whitespaceCheckRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, pos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	return rules.NewWhitespaceCheckFilter().AcceptRuleMatch(match, arguments, pos, patternTokens, tokenPositions)
}

// adaptSuggestionsRuleFilter ports AdaptSuggestionsFilter.
type adaptSuggestionsRuleFilter struct {
	inner *rules.AdaptSuggestionsFilter
}

func (f adaptSuggestionsRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, pos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f.inner == nil {
		return match
	}
	return f.inner.AcceptRuleMatch(match, arguments, pos, patternTokens, tokenPositions)
}

// addCommasRuleFilter ports AddCommasFilter.
type addCommasRuleFilter struct{}

func (addCommasRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, pos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	return rules.NewAddCommasFilter().AcceptRuleMatch(match, arguments, pos, patternTokens, tokenPositions)
}

// isEnglishWordRuleFilter ports IsEnglishWordFilter (uses default EN tagger hook).
type isEnglishWordRuleFilter struct{}

func (isEnglishWordRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, pos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	return rules.NewIsEnglishWordFilter(nil).AcceptRuleMatch(match, arguments, pos, patternTokens, tokenPositions)
}

// checkPostagsRuleFilter ports CheckPostagsInSuggestionFilter.
type checkPostagsRuleFilter struct{}

func (checkPostagsRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, pos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	return rules.NewCheckPostagsInSuggestionFilter(nil).AcceptRuleMatch(match, arguments, pos, patternTokens, tokenPositions)
}

// suppressIfAnyRuleFilter ports SuppressIfAnyRuleMatchesFilter.
type suppressIfAnyRuleFilter struct{}

func (suppressIfAnyRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, pos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	return rules.NewSuppressIfAnyRuleMatchesFilter(nil).AcceptRuleMatch(match, arguments, pos, patternTokens, tokenPositions)
}

// convertToSentenceCaseRuleFilter ports ConvertToSentenceCaseFilter as RuleFilter.
type convertToSentenceCaseRuleFilter struct {
	inner *rules.ConvertToSentenceCaseFilter
}

func (f convertToSentenceCaseRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f.inner == nil {
		return nil
	}
	return f.inner.AcceptRuleMatch(match, arguments, patternTokenPos, patternTokens, tokenPositions)
}

// dateRangeCheckerRuleFilter ports org.languagetool.rules.DateRangeChecker.
type dateRangeCheckerRuleFilter struct{}

func (dateRangeCheckerRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	if rules.NewDateRangeChecker().Accept(arguments["x"], arguments["y"]) {
		return match
	}
	return nil
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
	// Java MultitokenSpellerFilter: capitalize when patternTokenPos is first content token
	// (skip SENT_START + leading punctuation / non-word).
	inner := &multitoken.MultitokenSpellerFilter{
		Speller:         sp,
		IsMisspelled:    isMiss,
		AtSentenceStart: multitokenAtSentenceStart(match, patternTokenPos),
	}
	original := ""
	if match.Sentence != nil {
		text := match.Sentence.GetText()
		if match.FromPos >= 0 && match.ToPos <= len(text) && match.FromPos < match.ToPos {
			original = text[match.FromPos:match.ToPos]
		}
	}
	return inner.AcceptRuleMatch(match, original)
}

// multitokenAtSentenceStart ports MultitokenSpellerFilter sentence-start detection
// (StringTools.isPunctuationMark / isNotWordString skip).
func multitokenAtSentenceStart(match *rules.RuleMatch, patternTokenPos int) bool {
	if match == nil || match.Sentence == nil {
		return false
	}
	tokens := match.Sentence.GetTokensWithoutWhitespace()
	// Java: wordsStartPos = 1 (index 0 is SENT_START).
	wordsStartPos := 1
	for wordsStartPos < len(tokens) {
		t := tokens[wordsStartPos]
		if t == nil {
			wordsStartPos++
			continue
		}
		s := t.GetToken()
		if tools.IsPunctuationMark(s) || tools.IsNotWordString(s) {
			wordsStartPos++
			continue
		}
		break
	}
	return patternTokenPos == wordsStartPos
}
