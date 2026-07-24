package patterns

import (
	"strings"
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

func (underlineSpacesRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, pos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	return rules.NewUnderlineSpacesFilter().AcceptRuleMatch(match, arguments, pos, patternTokens, tokenPositions)
}

var (
	multitokenSpellerMu         sync.RWMutex
	defaultMultitokenSpeller    *multitoken.MultitokenSpeller
	defaultIsMisspelled         func(string) bool
	defaultMultitokenShortCode  string // Language.getShortCode() for MultitokenSpellerFilter gate
	defaultMultitokenShortFull  string // Language.getShortCodeWithCountryAndVariant()
)

// SetDefaultMultitokenSpeller wires MultitokenSpellerFilter's dictionary backend
// (Java: Language.getMultitokenSpeller). isMisspelled may be nil.
// Defaults short code to "en" so the Java en/de/pt/nl areTokensAcceptedBySpeller path is on
// (call sites for en/de/pt/nl). Prefer SetDefaultMultitokenSpellerWithOptions with the real short code.
func SetDefaultMultitokenSpeller(sp *multitoken.MultitokenSpeller, isMisspelled func(string) bool) {
	SetDefaultMultitokenSpellerWithOptions(sp, isMisspelled, true)
}

// SetDefaultMultitokenSpellerWithOptions wires getMultitokenSpeller + optional isMisspelled
// and MultitokenSpellerFilter short-code gate.
// checkSpelling true → default ShortCode "en" (en/de/pt/nl gate on; match.Rule language preferred when set).
// checkSpelling false → ShortCode "" (gate off unless match.Rule provides en/de/pt/nl).
// When isMisspelled is non-nil it is also set on sp.IsMisspelledToken for
// MultitokenSpeller.discardRunOnWords (Java SpellingCheckRule on all languages).
func SetDefaultMultitokenSpellerWithOptions(sp *multitoken.MultitokenSpeller, isMisspelled func(string) bool, checkSpelling bool) {
	if sp != nil && isMisspelled != nil {
		// Java MultitokenSpeller(language) always holds getDefaultSpellingRule().
		sp.IsMisspelledToken = isMisspelled
	}
	short, full := "", ""
	if checkSpelling {
		// Wire-time stand-in: any of en/de/pt/nl enables the same Java gate path.
		// Production prefers PatternRule language short code from the match when present.
		short, full = "en", "en"
	}
	multitokenSpellerMu.Lock()
	defer multitokenSpellerMu.Unlock()
	defaultMultitokenSpeller = sp
	defaultIsMisspelled = isMisspelled
	defaultMultitokenShortCode = short
	defaultMultitokenShortFull = full
}

// SetDefaultMultitokenSpellerForLanguage wires getMultitokenSpeller with explicit Language short codes
// (Java PatternRule.getLanguage().getShortCode / getShortCodeWithCountryAndVariant).
func SetDefaultMultitokenSpellerForLanguage(sp *multitoken.MultitokenSpeller, isMisspelled func(string) bool, shortCode, shortCodeWithCountryAndVariant string) {
	if sp != nil && isMisspelled != nil {
		sp.IsMisspelledToken = isMisspelled
	}
	if shortCodeWithCountryAndVariant == "" {
		shortCodeWithCountryAndVariant = shortCode
	}
	multitokenSpellerMu.Lock()
	defer multitokenSpellerMu.Unlock()
	defaultMultitokenSpeller = sp
	defaultIsMisspelled = isMisspelled
	defaultMultitokenShortCode = shortCode
	defaultMultitokenShortFull = shortCodeWithCountryAndVariant
}

// multitokenSpellerRuleFilter ports MultitokenSpellerFilter onto the optional default speller.
type multitokenSpellerRuleFilter struct{}

func (multitokenSpellerRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	multitokenSpellerMu.RLock()
	sp := defaultMultitokenSpeller
	isMiss := defaultIsMisspelled
	short := defaultMultitokenShortCode
	full := defaultMultitokenShortFull
	multitokenSpellerMu.RUnlock()
	if sp == nil {
		// Empty suggestions → Java returns null (do not invent replacements).
		return nil
	}
	// Prefer PatternRule language codes when present (Java getLanguage() on the rule).
	if code := multitokenRuleLanguageCode(match.Rule); code != "" {
		full = code
		short = code
		if i := strings.IndexByte(code, '-'); i >= 0 {
			short = code[:i]
		}
	}
	inner := &multitoken.MultitokenSpellerFilter{
		Speller:                        sp,
		IsMisspelled:                   isMiss,
		ShortCode:                      short,
		ShortCodeWithCountryAndVariant: full,
	}
	return inner.AcceptRuleMatch(match, arguments, patternTokenPos, patternTokens, tokenPositions)
}

// multitokenRuleLanguageCode reads LanguageCode from PatternRule-like rules
// (Java ((PatternRule) match.getRule()).getLanguage()).
func multitokenRuleLanguageCode(rule any) string {
	if rule == nil {
		return ""
	}
	if r, ok := rule.(interface{ GetLanguageCode() string }); ok {
		return r.GetLanguageCode()
	}
	return ""
}
