package languagetool

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// Constants and enums from org.languagetool.JLanguageTool.

const (
	SentenceStartTagName = "SENT_START"
	SentenceEndTagName   = "SENT_END"
	ParagraphEndTagName  = "PARA_END"

	PatternFile                 = "grammar.xml"
	StyleFile                   = "style.xml"
	CustomPatternFile           = "grammar_custom.xml"
	FalseFriendFile             = "false-friends.xml"
	MessageBundleName           = "org.languagetool.MessagesBundle"
	DictionaryFilenameExtension = ".dict"
)

// Mode ports JLanguageTool.Mode.
type Mode string

const (
	ModeAll             Mode = "ALL"
	ModeTextLevelOnly   Mode = "TEXTLEVEL_ONLY"
	ModeAllButTextLevel Mode = "ALL_BUT_TEXTLEVEL_ONLY"
)

// ParagraphHandling ports JLanguageTool.ParagraphHandling.
type ParagraphHandling string

const (
	ParagraphNormal      ParagraphHandling = "NORMAL"
	ParagraphOnlyPara    ParagraphHandling = "ONLYPARA"
	ParagraphOnlyNonPara ParagraphHandling = "ONLYNONPARA"
)

// CheckCancelledCallback ports JLanguageTool.CheckCancelledCallback.
type CheckCancelledCallback func() bool

// LocalMatch is a cycle-free rule-match surface for JLanguageTool.Check
// (avoids importing rules package into languagetool).
type LocalMatch struct {
	FromPos, ToPos int
	Message        string
	ShortMessage   string
	RuleID         string
	Suggestions    []string
	// Optional rule metadata (from the rule or SoftRuleMeta fallback for known Java families).
	Description  string
	CategoryID   string
	CategoryName string
	IssueType    string
	// URL ports RuleMatch.getUrl / Rule.getUrl (match-level overrides rule-level).
	URL string
	// OriginalErrorStr ports RuleMatch.getOriginalErrorStr (surface under marker).
	// Empty when not set; SwissGerman AI ss→ß drop uses this when present.
	OriginalErrorStr string
	// SentenceText is the analyzed sentence plain text when known (Check fills it).
	// Used by German.filterRuleMatches period-drop (Java getSentence().getText())
	// and CleanOverlappingFilter.isPunctuationOnlyChange sentence fallback.
	// Empty when not set — filters must fail closed without invent.
	SentenceText string
	// FromPosSentence/ToPosSentence port RuleMatch sentencePosition (-1/unset or
	// To<=From means unset). Check fills from pre-remap FromPos/ToPos when unset.
	// CleanOverlapping isPunctuationOnlyChange prefers these over document FromPos.
	FromPosSentence int
	ToPosSentence   int
	// Priority used by CleanOverlappingLocalMatches (higher wins).
	Priority int
	// IsPicky ports Rule tags containing Tag.picky (Java CleanOverlappingFilter demotion
	// and German.filterRuleMatches picky-equality for AI_DE_GGEC merge).
	IsPicky bool
	// IncludedInErrorsCorrectedAllAtOnce ports Rule.isIncludedInErrorsCorrectedAllAtOnce
	// (Java CleanOverlappingFilter punctuation-only preference).
	IncludedInErrorsCorrectedAllAtOnce bool
	// IsPremium ports Premium.isPremiumRule when known (ToLocalMatches / explicit inject).
	// When false, CleanOverlapping still treats RuleID containing "PREMIUM" as premium
	// (Java CleanOverlappingFilter default id heuristic).
	IsPremium bool
	// EnabledRules ports the enabledRules Set passed to Language.filterRuleMatches /
	// adjust*RuleMatch (French APOS_TYP, Catalan EXIGEIX_*/APOSTROF_*). Check stamps
	// JLanguageTool.EnabledRules onto each match before the language filter. Nil/empty
	// → filters fail closed for gated branches (same as empty set).
	EnabledRules map[string]struct{}
	// HasTypographicApostropheInSentence ports anyMatch(token.hasTypographicApostrophe)
	// on the analyzed sentence (Catalan.adjustCatalanMatch). Check stamps from tokens.
	HasTypographicApostropheInSentence bool
}

// SentenceChecker returns matches for one analyzed sentence (offsets relative to sentence text).
type SentenceChecker func(sentence *AnalyzedSentence) []LocalMatch

// TextLevelChecker returns matches across all sentences (document-relative offsets).
type TextLevelChecker func(sentences []*AnalyzedSentence) []LocalMatch

// JLanguageTool is a minimal façade for pure-Go check orchestration (growing).
// Full Java parity is not attempted here.
type JLanguageTool struct {
	LanguageCode string
	Mode         Mode
	Level        Level
	// sentenceMatchers reserved for MultiThreaded error surface.
	sentenceMatchers []func(sentence *AnalyzedSentence) error
	// checkers are pluggable sentence rules for Check.
	checkers []SentenceChecker
	// textLevelCheckers are multi-sentence rules (e.g. word-repeat-beginning).
	textLevelCheckers []struct {
		id string
		fn TextLevelChecker
	}
	// activeRuleIDs tracks rule IDs registered via AddRuleChecker (order preserved).
	activeRuleIDs []string
	// DisabledRuleIDs soft-disable matches / registration filtering.
	DisabledRuleIDs map[string]struct{}
	// EnabledRules tracks rule IDs known to be enabled for language filter hooks
	// (Java Set<String> enabledRules in filterRuleMatches). Populated by EnableRule
	// and variant default enabled lists. Incomplete vs Java full active-rule set —
	// only explicitly tracked IDs (not invent full registry scan).
	EnabledRules map[string]struct{}
	// DefaultOffRuleIDs are rules that registered with XML default="off" (optional packs).
	// SOFT_OPTIONAL re-enables these in addition to SOFT_OPT_* inventeds.
	DefaultOffRuleIDs map[string]struct{}
	// Cancelled optional early exit for Check.
	Cancelled CheckCancelledCallback
	// ListUnknownWords enables GetUnknownWords population during Check/AnalyzeUnknown.
	ListUnknownWords bool
	// IsKnownWord optional dictionary probe for unknown-word listing.
	IsKnownWord func(token string) bool
	// TagWord optional POS/lemma inject used by Analyze (MapWordTagger-friendly).
	TagWord func(token string) []TokenTag
	// Disambiguator optional post-tag sentence filter (multiword chunker / XML rules).
	Disambiguator SentenceDisambiguator
	// Chunker ports Java Language.getChunker(); runs on tagged tokens before
	// disambiguation (JLanguageTool.getRawAnalyzedSentence).
	// Interface lives here to avoid import cycles with package chunking.
	Chunker SentenceChunker
	// PostDisambiguationChunker ports Language.getPostDisambiguationChunker();
	// runs after disambiguation when set.
	PostDisambiguationChunker SentenceChunker
	// IgnoreWords soft user-dictionary / spell-ignore set (surface forms).
	IgnoreWords map[string]struct{}
	// UserConfig optional user preferences (accepted phrases, speller words).
	UserConfig *UserConfig
	// PriorityForId ports Language.getPriorityForId when set (e.g. GermanPriorityForId).
	// Applied in Check via applyRulePriorities (Java getRulePriority: rule id then category id).
	PriorityForId func(id string) int
	// DefaultRulePriorityForStyle ports Language.getDefaultRulePriorityForStyle
	// (English/Catalan return -50; base Language returns 0).
	// Applied when rule/category priority is 0 and IssueType is style.
	DefaultRulePriorityForStyle int
	// FilterRuleMatches ports Language.filterRuleMatches when set (e.g. German AI_DE_GGEC merge).
	// Runs before CleanOverlappingLocalMatches (Java LanguageDependentRuleMatchFilter order).
	FilterRuleMatches func(matches []LocalMatch) []LocalMatch
	// FilterRuleMatchesAfterOverlapping ports Language.filterRuleMatchesAfterOverlapping.
	// Runs after CleanOverlappingLocalMatches. Nil = identity.
	FilterRuleMatchesAfterOverlapping func(matches []LocalMatch) []LocalMatch
	// CleanOverlapping when false skips CleanOverlappingLocalMatches (Java setCleanOverlappingMatches).
	// Zero value false means enabled by default via cleanOverlappingEnabled().
	// Use DisableCleanOverlapping to turn off.
	disableCleanOverlapping bool
	// IgnoredCharacters ports Language.getIgnoredCharactersRegex (applied per word token
	// after tokenize, Java replaceSoftHyphens). Nil = no strip. German uses [\u00AD].
	IgnoredCharacters *regexp.Regexp
	unknown           map[string]struct{}
}

// DisableCleanOverlapping ports JLanguageTool.setCleanOverlappingMatches(false).
func (lt *JLanguageTool) DisableCleanOverlapping() {
	if lt != nil {
		lt.disableCleanOverlapping = true
	}
}

// EnableCleanOverlapping ports JLanguageTool.setCleanOverlappingMatches(true).
func (lt *JLanguageTool) EnableCleanOverlapping() {
	if lt != nil {
		lt.disableCleanOverlapping = false
	}
}

// SentenceDisambiguator filters/augments POS on an analyzed sentence (soft LT disambiguator hook).
type SentenceDisambiguator interface {
	Disambiguate(input *AnalyzedSentence) *AnalyzedSentence
}

// SentenceChunker ports org.languagetool.chunking.Chunker for Analyze wiring.
type SentenceChunker interface {
	AddChunkTags(tokens []*AnalyzedTokenReadings)
}

// FilterFrenchRuleMatchesHook ports French.filterRuleMatches for Check wiring.
// Set by package language init (avoids import cycle: rules/fr ↔ language tests).
var FilterFrenchRuleMatchesHook func(matches []LocalMatch) []LocalMatch

// FilterSpanishRuleMatchesHook ports Spanish.filterRuleMatches for Check wiring.
// Set by package language init (same cycle-avoidance pattern as French).
var FilterSpanishRuleMatchesHook func(matches []LocalMatch) []LocalMatch

// FilterEnglishRuleMatchesHook ports English.filterRuleMatches for Check wiring.
var FilterEnglishRuleMatchesHook func(matches []LocalMatch) []LocalMatch

// FrenchPriorityForIdHook ports French.getPriorityForId for Check wiring
// (set by package language init; avoids rules/fr ↔ language test import cycle).
var FrenchPriorityForIdHook func(id string) int

// EnglishPriorityForIdHook ports English.getPriorityForId for Check wiring
// (set by package language init).
var EnglishPriorityForIdHook func(id string) int

// EnglishPriorityForIdForCodeHook selects BritishEnglish vs English by lang code
// (set by package language init).
var EnglishPriorityForIdForCodeHook func(langCode string) func(id string) int

// FilterCatalanRuleMatchesHook ports Catalan.filterRuleMatches for Check wiring.
var FilterCatalanRuleMatchesHook func(matches []LocalMatch) []LocalMatch

// FilterCatalanRuleMatchesAfterOverlappingHook ports Catalan.filterRuleMatchesAfterOverlapping.
var FilterCatalanRuleMatchesAfterOverlappingHook func(matches []LocalMatch) []LocalMatch

// VariantDefaultRulesHook ports Language.getDefaultEnabled/DisabledRulesForVariant
// lookup by short code. Set by package language init (avoids import cycle).
var VariantDefaultRulesHook func(langCode string) (enabled, disabled []string)

// LanguageAdaptSuggestionByCode maps language short codes (e.g. "ca", "es") to
// Language.adaptSuggestion. Set by package language / RegisterCore* (Java AdaptSuggestionsFilter).
// Lookups use the primary subtag (ca-ES → ca). Nil entry → identity.
var LanguageAdaptSuggestionByCode = map[string]func(replacement, originalError string) string{}

// AdaptSuggestionForLanguage returns Language.adaptSuggestion for langCode, or nil (identity).
func AdaptSuggestionForLanguage(langCode string) func(replacement, originalError string) string {
	if langCode == "" {
		return nil
	}
	// exact then primary subtag
	if f, ok := LanguageAdaptSuggestionByCode[langCode]; ok {
		return f
	}
	base := langCode
	for i := 0; i < len(langCode); i++ {
		if langCode[i] == '-' || langCode[i] == '_' {
			base = langCode[:i]
			break
		}
	}
	if f, ok := LanguageAdaptSuggestionByCode[base]; ok {
		return f
	}
	// case-insensitive primary
	lower := ""
	for i := 0; i < len(base); i++ {
		c := base[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		lower += string(c)
	}
	return LanguageAdaptSuggestionByCode[lower]
}

func NewJLanguageTool(languageCode string) *JLanguageTool {
	lt := &JLanguageTool{
		LanguageCode:    languageCode,
		Mode:            ModeAll,
		Level:           LevelDefault,
		DisabledRuleIDs: map[string]struct{}{},
		EnabledRules:    map[string]struct{}{},
	}
	// Java activateDefaultPatternRules: apply variant default on/off lists.
	lt.applyVariantDefaultRules()
	return lt
}

// applyVariantDefaultRules ports JLanguageTool.activateDefaultPatternRules enabled/disabled
// lists from Language.getDefaultEnabledRulesForVariant / getDefaultDisabledRulesForVariant.
// Maps Java setDefaultOn → EnableRule (clears Disabled + DefaultOff); setDefaultOff → MarkDefaultOff.
// Call again after pattern rules load (RegisterGrammar*) so default-off XML rules can be turned on.
func (lt *JLanguageTool) applyVariantDefaultRules() {
	if lt == nil || VariantDefaultRulesHook == nil {
		return
	}
	enabled, disabled := VariantDefaultRulesHook(lt.LanguageCode)
	for _, id := range disabled {
		// Java patternRule.setDefaultOff()
		lt.MarkDefaultOff(id)
	}
	for _, id := range enabled {
		// Java patternRule.setDefaultOn() — re-enable even if XML default="off"
		lt.EnableRule(id)
	}
}

// ApplyVariantDefaultRules is the public re-entry after grammar registration
// (Java activates defaults after getPatternRules()).
func (lt *JLanguageTool) ApplyVariantDefaultRules() {
	lt.applyVariantDefaultRules()
}

// enabledRulesForFilters builds the Set passed to Language.filterRuleMatches.
// Java: currently enabled rule IDs. Go: explicit EnabledRules plus registered
// active (not disabled) rule IDs — incomplete vs every pattern rule instance
// default-on state beyond DisableRule tracking (no invent of unregistered IDs).
func (lt *JLanguageTool) enabledRulesForFilters() map[string]struct{} {
	if lt == nil {
		return nil
	}
	out := make(map[string]struct{})
	for id := range lt.EnabledRules {
		if id != "" && !lt.isRuleDisabled(id) {
			out[id] = struct{}{}
		}
	}
	for _, id := range lt.activeRuleIDs {
		if id != "" && !lt.isRuleDisabled(id) {
			out[id] = struct{}{}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (lt *JLanguageTool) GetLanguageCode() string { return lt.LanguageCode }
func (lt *JLanguageTool) GetMode() Mode           { return lt.Mode }
func (lt *JLanguageTool) SetMode(m Mode)          { lt.Mode = m }
func (lt *JLanguageTool) GetLevel() Level         { return lt.Level }
func (lt *JLanguageTool) SetLevel(l Level)        { lt.Level = l }

// AddChecker registers a sentence-level rule for Check.
func (lt *JLanguageTool) AddChecker(c SentenceChecker) {
	if lt == nil || c == nil {
		return
	}
	lt.checkers = append(lt.checkers, c)
}

// AddRuleChecker registers a checker and records its rule ID for enable/disable.
func (lt *JLanguageTool) AddRuleChecker(ruleID string, c SentenceChecker) {
	if lt == nil || c == nil {
		return
	}
	if ruleID != "" {
		lt.activeRuleIDs = append(lt.activeRuleIDs, ruleID)
	}
	id := ruleID
	lt.checkers = append(lt.checkers, func(s *AnalyzedSentence) []LocalMatch {
		if id != "" && lt.isRuleDisabled(id) {
			return nil
		}
		return c(s)
	})
}

// AddTextLevelRuleChecker registers a multi-sentence rule (document-relative offsets).
func (lt *JLanguageTool) AddTextLevelRuleChecker(ruleID string, c TextLevelChecker) {
	if lt == nil || c == nil {
		return
	}
	if ruleID != "" {
		lt.activeRuleIDs = append(lt.activeRuleIDs, ruleID)
	}
	lt.textLevelCheckers = append(lt.textLevelCheckers, struct {
		id string
		fn TextLevelChecker
	}{id: ruleID, fn: c})
}

// DisableRule ports disableRule.
func (lt *JLanguageTool) DisableRule(ruleID string) {
	if lt == nil || ruleID == "" {
		return
	}
	if lt.DisabledRuleIDs == nil {
		lt.DisabledRuleIDs = map[string]struct{}{}
	}
	lt.DisabledRuleIDs[ruleID] = struct{}{}
}

// MarkDefaultOff records that ruleID was registered with XML/Java default="off"
// and disables it until EnableRule (Java Rule.setDefaultOff + startup state).
func (lt *JLanguageTool) MarkDefaultOff(ruleID string) {
	if lt == nil || ruleID == "" {
		return
	}
	if lt.DefaultOffRuleIDs == nil {
		lt.DefaultOffRuleIDs = map[string]struct{}{}
	}
	lt.DefaultOffRuleIDs[ruleID] = struct{}{}
	lt.DisableRule(ruleID)
}

// GetDefaultOffRuleIDs returns rule IDs registered with default="off".
func (lt *JLanguageTool) GetDefaultOffRuleIDs() []string {
	if lt == nil || len(lt.DefaultOffRuleIDs) == 0 {
		return nil
	}
	out := make([]string, 0, len(lt.DefaultOffRuleIDs))
	for id := range lt.DefaultOffRuleIDs {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// EnableRule ports enableRule / AbstractPatternRule.setDefaultOn:
// re-enables a disabled rule, clears DefaultOff tracking, and tracks the ID in
// EnabledRules for language filter hooks (APOS_TYP, EXIGEIX_*, …).
func (lt *JLanguageTool) EnableRule(ruleID string) {
	if lt == nil || ruleID == "" {
		return
	}
	if lt.DisabledRuleIDs != nil {
		delete(lt.DisabledRuleIDs, ruleID)
	}
	if lt.DefaultOffRuleIDs != nil {
		delete(lt.DefaultOffRuleIDs, ruleID)
	}
	if lt.EnabledRules == nil {
		lt.EnabledRules = map[string]struct{}{}
	}
	lt.EnabledRules[ruleID] = struct{}{}
}

// IsRuleEnabled reports whether ruleID is tracked as enabled (EnabledRules) and
// not in DisabledRuleIDs. Incomplete vs Java full active set — only tracked IDs.
func (lt *JLanguageTool) IsRuleEnabled(ruleID string) bool {
	if lt == nil || ruleID == "" {
		return false
	}
	if lt.isRuleDisabled(ruleID) {
		return false
	}
	if lt.EnabledRules == nil {
		return false
	}
	_, ok := lt.EnabledRules[ruleID]
	return ok
}

// GetAllRegisteredRuleIDs returns every rule ID registered via AddRuleChecker / AddTextLevelRuleChecker.
func (lt *JLanguageTool) GetAllRegisteredRuleIDs() []string {
	if lt == nil {
		return nil
	}
	return append([]string(nil), lt.activeRuleIDs...)
}

// GetAllActiveRuleIDs returns registered rule IDs that are not disabled.
func (lt *JLanguageTool) GetAllActiveRuleIDs() []string {
	if lt == nil {
		return nil
	}
	var out []string
	for _, id := range lt.activeRuleIDs {
		if !lt.isRuleDisabled(id) {
			out = append(out, id)
		}
	}
	return out
}

func (lt *JLanguageTool) isRuleDisabled(id string) bool {
	if lt == nil || lt.DisabledRuleIDs == nil {
		return false
	}
	_, ok := lt.DisabledRuleIDs[id]
	return ok
}

// IsRuleDisabled reports whether ruleID is in DisabledRuleIDs.
func (lt *JLanguageTool) IsRuleDisabled(ruleID string) bool {
	return lt.isRuleDisabled(ruleID)
}

// SetListUnknownWords ports setListUnknownWords.
func (lt *JLanguageTool) SetListUnknownWords(v bool) {
	if lt != nil {
		lt.ListUnknownWords = v
	}
}

// GetUnknownWords ports getUnknownWords (sorted unique).
func (lt *JLanguageTool) GetUnknownWords() []string {
	if lt == nil || len(lt.unknown) == 0 {
		return nil
	}
	out := make([]string, 0, len(lt.unknown))
	for w := range lt.unknown {
		out = append(out, w)
	}
	sort.Strings(out)
	return out
}

// Analyze splits text into sentences and builds plain analyzed sentences.
func (lt *JLanguageTool) Analyze(text string) []*AnalyzedSentence {
	st := tokenizers.NewSRXSentenceTokenizer(lt.LanguageCode)
	parts := st.Tokenize(text)
	if len(parts) == 0 {
		if text == "" {
			return nil
		}
		parts = []string{text}
	}
	out := make([]*AnalyzedSentence, 0, len(parts))
	wt := WordTokenizerForLanguage(lt.LanguageCode)
	// Java PolishWordTokenizer.setTagger: hyphen compounds split using POS (adja+adj, …).
	attachPolishHyphenTagger(wt, lt.TagWord)
	var ignore *regexp.Regexp
	if lt != nil {
		ignore = lt.IgnoredCharacters
	}
	for _, p := range parts {
		var s *AnalyzedSentence
		if lt.TagWord != nil {
			s = AnalyzeWithTaggerTokenizerAndIgnore(p, lt.TagWord, wt, ignore)
		} else {
			s = AnalyzeWithTokenizerAndIgnore(p, wt, ignore)
		}
		// Java getRawAnalyzedSentence: tagger then language.getChunker().
		if s != nil && lt.Chunker != nil {
			lt.Chunker.AddChunkTags(s.GetTokens())
		}
		// Preserve pre-disambiguation tokens for pattern raw_pos="yes"
		// (Java AnalyzedSentence keeps both token arrays).
		var preDisambig []*AnalyzedTokenReadings
		if s != nil {
			preDisambig = cloneAnalyzedTokenSlice(s.GetTokens())
		}
		if lt.Disambiguator != nil && s != nil {
			if d := lt.Disambiguator.Disambiguate(s); d != nil {
				s = d
			}
		}
		if s != nil && preDisambig != nil {
			s = NewAnalyzedSentenceFull(s.GetTokens(), preDisambig)
		}
		// Java getAnalyzedSentence: optional post-disambiguation chunker.
		if s != nil && lt.PostDisambiguationChunker != nil {
			lt.PostDisambiguationChunker.AddChunkTags(s.GetTokens())
		}
		out = append(out, s)
	}
	return out
}

// sentenceHasTypographicApostrophe ports Java stream anyMatch(hasTypographicApostrophe).
func sentenceHasTypographicApostrophe(s *AnalyzedSentence) bool {
	if s == nil {
		return false
	}
	for _, tok := range s.GetTokensWithoutWhitespace() {
		if tok != nil && tok.HasTypographicApostrophe() {
			return true
		}
	}
	return false
}

// Check runs registered checkers over analyzed sentences and returns document-offset matches.
func (lt *JLanguageTool) Check(text string) []LocalMatch {
	if lt == nil {
		return nil
	}
	if lt.Cancelled != nil && lt.Cancelled() {
		return nil
	}
	lt.unknown = map[string]struct{}{}
	sents := lt.Analyze(text)
	var out []LocalMatch
	runSentence := lt.Mode != ModeTextLevelOnly
	runTextLevel := lt.Mode != ModeAllButTextLevel

	// Map sentence-local offsets to document by searching each sentence text in remaining source.
	// AnalyzePlain token positions are relative to the sentence string.
	if runSentence {
		srcRunes := []rune(text)
		searchFrom := 0
		for _, s := range sents {
			if lt.Cancelled != nil && lt.Cancelled() {
				break
			}
			if s == nil {
				continue
			}
			stext := s.GetText()
			// find sentence start in document
			docBase := indexRunesFrom(srcRunes, []rune(stext), searchFrom)
			if docBase < 0 {
				docBase = searchFrom
			}
			if lt.ListUnknownWords {
				lt.collectUnknown(s)
			}
			// Java Catalan.adjustCatalanMatch: any token hasTypographicApostrophe().
			sentTypoApos := sentenceHasTypographicApostrophe(s)
			for _, c := range lt.checkers {
				for _, m := range c(s) {
					// Java RuleMatch keeps sentencePosition + AnalyzedSentence.
					// Capture sentence-local span before document remap for
					// CleanOverlappingFilter.isPunctuationOnlyChange fallback.
					if m.FromPosSentence < 0 || m.ToPosSentence <= m.FromPosSentence {
						m.FromPosSentence = m.FromPos
						m.ToPosSentence = m.ToPos
					}
					m.FromPos += docBase
					m.ToPos += docBase
					if m.SentenceText == "" {
						m.SentenceText = stext
					}
					m.HasTypographicApostropheInSentence = sentTypoApos
					out = append(out, m)
				}
			}
			searchFrom = docBase + len([]rune(stext))
		}
	} else if lt.ListUnknownWords {
		for _, s := range sents {
			lt.collectUnknown(s)
		}
	}

	if runTextLevel && len(lt.textLevelCheckers) > 0 {
		if lt.Cancelled == nil || !lt.Cancelled() {
			for _, tc := range lt.textLevelCheckers {
				if tc.id != "" && lt.isRuleDisabled(tc.id) {
					continue
				}
				out = append(out, tc.fn(sents)...)
			}
		}
	}
	// Java filterMatches order:
	// SameRuleGroupFilter → LanguageDependentRuleMatchFilter (filterRuleMatches)
	// → CleanOverlappingFilter → filterRuleMatchesAfterOverlapping
	out = lt.applyRulePriorities(out)
	out = CleanSameRuleGroupLocalMatches(out)
	// Stamp enabled rule set for language filters (Java filterRuleMatches(..., enabledRules)).
	if en := lt.enabledRulesForFilters(); len(en) > 0 {
		for i := range out {
			out[i].EnabledRules = en
		}
	}
	if lt.FilterRuleMatches != nil {
		out = lt.FilterRuleMatches(out)
	}
	// Default cleanOverlappingMatches is true in Java JLanguageTool.
	if lt == nil || !lt.disableCleanOverlapping {
		hidePremium := lt != nil && lt.UserConfig != nil && lt.UserConfig.HidePremiumMatches
		out = CleanOverlappingLocalMatchesOpts(out, CleanOverlapOpts{HidePremiumMatches: hidePremium})
	}
	if lt != nil && lt.FilterRuleMatchesAfterOverlapping != nil {
		out = lt.FilterRuleMatchesAfterOverlapping(out)
	}
	return lt.filterMatchesByIgnore(text, out)
}

// applyRulePriorities ports Language.getRulePriority for each match when PriorityForId is set
// (or DefaultRulePriorityForStyle is non-zero).
// Java order: rule id priority → rule.getPriority (LocalMatch.Priority inject) →
// category id priority → getDefaultRulePriorityForStyle if ITS Style → 0.
// Leaves Priority unchanged when already non-zero (explicit inject / rule priority).
func (lt *JLanguageTool) applyRulePriorities(ms []LocalMatch) []LocalMatch {
	if lt == nil || len(ms) == 0 {
		return ms
	}
	if lt.PriorityForId == nil && lt.DefaultRulePriorityForStyle == 0 {
		return ms
	}
	for i := range ms {
		if ms[i].Priority != 0 {
			continue
		}
		if lt.PriorityForId != nil {
			if id := ms[i].RuleID; id != "" {
				if p := lt.PriorityForId(id); p != 0 {
					ms[i].Priority = p
					continue
				}
			}
			if cat := ms[i].CategoryID; cat != "" {
				if p := lt.PriorityForId(cat); p != 0 {
					ms[i].Priority = p
					continue
				}
			}
		}
		// Java: getDefaultRulePriorityForStyle when ITSIssueType.Style
		if lt.DefaultRulePriorityForStyle != 0 && isStyleIssueType(ms[i].IssueType) {
			ms[i].Priority = lt.DefaultRulePriorityForStyle
		}
	}
	return ms
}

func isStyleIssueType(it string) bool {
	return strings.EqualFold(strings.TrimSpace(it), "style")
}

// collectUnknown ports JLanguageTool.rememberUnknownWords:
// if (!reading.isTagged()) unknownWords.add(reading.getToken()).
// Soft incomplete vs full Java DE tagger: when tokens stay untagged (no TagWord),
// IsKnownWord acts as a soft “known lexicon” so soft tests without Morphy still work.
// Do not skip IsSentenceEnd — Java attaches SENT_END on the last content token
// (hasNoTag), and still lists that surface when untagged (e.g. "description").
func (lt *JLanguageTool) collectUnknown(s *AnalyzedSentence) {
	for _, tok := range s.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		w := tok.GetToken()
		if w == "" || !hasLetterLocal(w) {
			continue
		}
		// Java: only untagged tokens are unknown.
		if tok.IsTagged() {
			continue
		}
		// Soft dict: treat as known when no real POS (fail-closed without invent tags).
		if lt.IsKnownWord != nil && lt.IsKnownWord(w) {
			continue
		}
		// Without TagWord and without IsKnownWord, listing everything untagged would
		// spam every token — require a soft dict or real tags (Java always has tagger).
		if lt.IsKnownWord == nil && lt.TagWord == nil {
			continue
		}
		lt.unknown[w] = struct{}{}
	}
}

func hasLetterLocal(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func indexRunesFrom(haystack, needle []rune, from int) int {
	if len(needle) == 0 {
		return from
	}
	if from < 0 {
		from = 0
	}
	for i := from; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// cleanOverlappingPickyDemotion ports Java CleanOverlappingFilter.negativeConstant
// (Integer.MIN_VALUE + 10000) applied when the rule has Tag.picky.
const cleanOverlappingPickyDemotion = math.MinInt32 + 10000

// CleanOverlapOpts ports CleanOverlappingFilter constructor flags used by Check.
type CleanOverlapOpts struct {
	// HidePremiumMatches demotes premium matches to MinInt32 (Java hidePremiumMatches).
	HidePremiumMatches bool
}

// CleanOverlappingLocalMatches ports CleanOverlappingFilter for LocalMatch with default opts.
func CleanOverlappingLocalMatches(matches []LocalMatch) []LocalMatch {
	return CleanOverlappingLocalMatchesOpts(matches, CleanOverlapOpts{})
}

// CleanOverlappingLocalMatchesOpts ports CleanOverlappingFilter for LocalMatch:
// sort by FromPos, walk sequentially; on overlap keep higher effective priority;
// on equal priority keep the longer span; on still equal keep the later match.
// Juxtaposed (non-overlapping) matches are both kept.
// Picky matches are demoted by cleanOverlappingPickyDemotion (Java).
// When both matches are punctuation-only changes (letters/digits unchanged), prefer
// IncludedInErrorsCorrectedAllAtOnce like Java.
// When HidePremiumMatches, premium rules get priority MinInt32 (Java).
func CleanOverlappingLocalMatchesOpts(matches []LocalMatch, opts CleanOverlapOpts) []LocalMatch {
	if len(matches) <= 1 {
		return matches
	}
	sorted := append([]LocalMatch(nil), matches...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].FromPos < sorted[j].FromPos
	})
	var cleanList []LocalMatch
	var prev LocalMatch
	havePrev := false
	for _, cur := range sorted {
		if !havePrev {
			prev = cur
			havePrev = true
			continue
		}
		// Java requires non-decreasing FromPos (guaranteed by sort).
		isDupSug := cleanOverlapDuplicateSuggestion(prev, cur)
		// no overlapping (juxtaposed errors are not removed)
		if cur.FromPos >= prev.ToPos && !isDupSug {
			cleanList = append(cleanList, prev)
			prev = cur
			continue
		}
		// overlapping — compare effective priorities
		curP := cleanOverlapEffectivePriority(cur, opts.HidePremiumMatches)
		prevP := cleanOverlapEffectivePriority(prev, opts.HidePremiumMatches)
		// Java: both punctuation-only → prefer isIncludedInErrorsCorrectedAllAtOnce.
		if cleanOverlapIsPunctuationOnly(cur) && cleanOverlapIsPunctuationOnly(prev) {
			curAll := cur.IncludedInErrorsCorrectedAllAtOnce
			prevAll := prev.IncludedInErrorsCorrectedAllAtOnce
			if curAll != prevAll {
				if curAll {
					if curP < prevP {
						curP = prevP + 1
					}
				} else if prevP < curP {
					prevP = curP + 1
				}
			}
		}
		if curP == prevP {
			// take the longest error
			curP = cur.ToPos - cur.FromPos
			prevP = prev.ToPos - prev.FromPos
		}
		if curP == prevP {
			curP++ // take the last one
		}
		if curP > prevP {
			prev = cur
		}
	}
	if havePrev {
		cleanList = append(cleanList, prev)
	}
	return cleanList
}

func cleanOverlapEffectivePriority(m LocalMatch, hidePremium bool) int {
	p := m.Priority
	if hidePremium && cleanOverlapIsPremium(m) {
		p = math.MinInt32
	}
	if m.IsPicky && p != math.MinInt32 {
		p += cleanOverlappingPickyDemotion
	}
	return p
}

// cleanOverlapIsPremium ports CleanOverlappingFilter.isPremiumRule:
//   1. explicit LocalMatch.IsPremium (ToLocalMatches / inject)
//   2. DefaultPremium.IsPremiumRule(ruleID) — Java Premium.get().isPremiumRule(rule)
//   3. RuleID contains "PREMIUM" — LocalMatch fallback when no Premium registry
//      and the rule id itself encodes premium (open-source PremiumOff is always false)
func cleanOverlapIsPremium(m LocalMatch) bool {
	if m.IsPremium {
		return true
	}
	if DefaultPremium != nil && DefaultPremium.IsPremiumRule(m.RuleID) {
		return true
	}
	return strings.Contains(m.RuleID, "PREMIUM")
}

// cleanOverlapIsPunctuationOnly ports CleanOverlappingFilter.isPunctuationOnlyChange:
// first suggestion vs original surface; letters+digits equal ⇒ punctuation-only change.
// Original surface via LocalMatch.OriginalSurface (Java getOriginalErrorStr / sentence).
// Without surface or suggestions → false (fail-closed).
func cleanOverlapIsPunctuationOnly(m LocalMatch) bool {
	if len(m.Suggestions) == 0 {
		return false
	}
	replacement := m.Suggestions[0]
	original := m.OriginalSurface()
	if original == "" {
		return false
	}
	if replacement == original {
		return false
	}
	return cleanOverlapKeepLettersAndDigits(original) == cleanOverlapKeepLettersAndDigits(replacement)
}

// OriginalSurface ports marked-error text used by CleanOverlappingFilter and
// SwissGerman.filterRuleMatches: OriginalErrorStr when set, else
// SentenceText[from:to] with sentence positions first then document positions
// (Java getOriginalErrorStr / sentence.substring(from,to) — UTF-16 indices).
// Empty when unknown — filters must fail closed without invent.
func (m LocalMatch) OriginalSurface() string {
	if m.OriginalErrorStr != "" {
		return m.OriginalErrorStr
	}
	st := m.SentenceText
	if st == "" {
		return ""
	}
	// Prefer sentence-relative span (Java fromPosSentence / Check pre-remap).
	if s := localMatchUTF16Span(st, m.FromPosSentence, m.ToPosSentence); s != "" {
		return s
	}
	// Java SwissGerman / CleanOverlapping second branch: fromPos/toPos on sentence text.
	return localMatchUTF16Span(st, m.FromPos, m.ToPos)
}

// TrimMatchEnds ports RuleMatch.trimMatchEnds: while all suggestions share a
// leading/trailing space-separated token with OriginalSurface, strip that token
// from suggestions and shrink FromPos/ToPos (and sentence spans) accordingly.
// Position deltas use Java String.length (UTF-16 code units); string slices use
// Go UTF-8 byte indices (faithful to Java substring + length).
// Unchanged when no suggestions or no surface (fail-closed).
func (m LocalMatch) TrimMatchEnds() LocalMatch {
	if len(m.Suggestions) == 0 {
		return m
	}
	errorStr := m.OriginalSurface()
	if errorStr == "" {
		return m
	}
	fromPos, toPos := m.FromPos, m.ToPos
	fromSent, toSent := m.FromPosSentence, m.ToPosSentence
	repls := append([]string(nil), m.Suggestions...)
	origFrom, origTo := fromPos, toPos
	changed := true
	for changed {
		changed = false
		// Try trimming from the end
		lastSpaceIdx := strings.LastIndex(errorStr, " ")
		if lastSpaceIdx > 0 {
			lastToken := errorStr[lastSpaceIdx+1:]
			endSuffix := " " + lastToken
			allEnd := true
			for _, r := range repls {
				if !strings.HasSuffix(r, endSuffix) {
					allEnd = false
					break
				}
			}
			if allEnd {
				// Java: errorStr.length() - lastSpaceIdx (UTF-16 units of " "+lastToken)
				errorTrimLen := localMatchUTF16Len(errorStr[lastSpaceIdx:])
				for i := range repls {
					repls[i] = repls[i][:len(repls[i])-len(endSuffix)]
				}
				toPos -= errorTrimLen
				if toSent >= 0 {
					toSent -= errorTrimLen
				}
				errorStr = errorStr[:lastSpaceIdx]
				changed = true
			}
		}
		// Try trimming from the beginning
		firstSpaceIdx := strings.Index(errorStr, " ")
		if firstSpaceIdx > 0 {
			firstToken := errorStr[:firstSpaceIdx]
			startPrefix := firstToken + " "
			allStart := true
			for _, r := range repls {
				if !strings.HasPrefix(r, startPrefix) {
					allStart = false
					break
				}
			}
			if allStart {
				// Java: firstSpaceIdx + 1 as UTF-16 length of firstToken+" "
				errorTrimLen := localMatchUTF16Len(startPrefix)
				for i := range repls {
					repls[i] = repls[i][len(startPrefix):]
				}
				fromPos += errorTrimLen
				if fromSent >= 0 {
					fromSent += errorTrimLen
				}
				// Advance errorStr by Go byte length of the same prefix.
				errorStr = errorStr[len(startPrefix):]
				changed = true
			}
		}
	}
	if fromPos == origFrom && toPos == origTo {
		return m
	}
	out := m
	out.FromPos = fromPos
	out.ToPos = toPos
	out.FromPosSentence = fromSent
	out.ToPosSentence = toSent
	out.Suggestions = repls
	out.OriginalErrorStr = errorStr
	return out
}

// localMatchUTF16Span slices sentence text by Java UTF-16 code-unit indices.
func localMatchUTF16Span(st string, from, to int) string {
	if from < 0 || to < 0 || from >= to {
		return ""
	}
	if to > localMatchUTF16Len(st) {
		return ""
	}
	return localMatchUTF16Substring(st, from, to)
}

func localMatchUTF16Len(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}

func localMatchUTF16Substring(s string, from, to int) string {
	u := make([]uint16, 0, len(s))
	for _, r := range s {
		if r >= 0x10000 {
			r -= 0x10000
			u = append(u, uint16(0xD800+(r>>10)), uint16(0xDC00+(r&0x3FF)))
		} else {
			u = append(u, uint16(r))
		}
	}
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	var runes []rune
	for i := from; i < to; {
		r := rune(u[i])
		if r >= 0xD800 && r <= 0xDBFF && i+1 < to {
			r2 := rune(u[i+1])
			if r2 >= 0xDC00 && r2 <= 0xDFFF {
				runes = append(runes, 0x10000+((r-0xD800)<<10)|(r2-0xDC00))
				i += 2
				continue
			}
		}
		runes = append(runes, r)
		i++
	}
	return string(runes)
}

func cleanOverlapKeepLettersAndDigits(s string) string {
	var b strings.Builder
	for _, ch := range s {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

// cleanOverlapDuplicateSuggestion ports Java CleanOverlappingFilter juxtaposed
// duplicate-comma / split-suggestion detection (treated as overlap).
func cleanOverlapDuplicateSuggestion(prev, cur LocalMatch) bool {
	if len(prev.Suggestions) == 0 || len(cur.Suggestions) == 0 {
		return false
	}
	sug, prevSug := cur.Suggestions[0], prev.Suggestions[0]
	if cur.FromPos == prev.ToPos {
		if strings.HasSuffix(prevSug, ",") && strings.HasPrefix(sug, ", ") {
			return true
		}
	}
	// Java: indexOf(" ") > 0 — space must not be at index 0
	if strings.Index(sug, " ") > 0 && strings.Index(prevSug, " ") > 0 &&
		cur.FromPos == prev.ToPos+1 {
		parts := strings.Split(sug, " ")
		partsPrev := strings.Split(prevSug, " ")
		if len(partsPrev) > 1 && len(parts) > 1 && partsPrev[1] == parts[0] {
			return true
		}
	}
	return false
}

func spansOverlap(a0, a1, b0, b1 int) bool {
	return a0 < b1 && b0 < a1
}

// CleanSameRuleGroupLocalMatches ports SameRuleGroupFilter for LocalMatch:
// sort by FromPos; keep first of overlapping matches with the same RuleID.
func CleanSameRuleGroupLocalMatches(matches []LocalMatch) []LocalMatch {
	if len(matches) <= 1 {
		return matches
	}
	sorted := append([]LocalMatch(nil), matches...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].FromPos < sorted[j].FromPos
	})
	var filtered []LocalMatch
	for i := 0; i < len(sorted); i++ {
		match := sorted[i]
		for i < len(sorted)-1 &&
			sameRuleGroupOverlapLocal(match, sorted[i+1]) {
			i++ // skip next match (Java SameRuleGroupFilter)
		}
		filtered = append(filtered, match)
	}
	return filtered
}

func sameRuleGroupOverlapLocal(match, next LocalMatch) bool {
	if match.RuleID == "" || match.RuleID != next.RuleID {
		return false
	}
	// Java RuleMatch overlaps: fromPos <= next.toPos && toPos >= next.fromPos
	return match.FromPos <= next.ToPos && match.ToPos >= next.FromPos
}

// PreferredWordRepeatFactory is set by package rules to the faithful WordRepeatRule
// checker (Java equalsIgnoreCase + ignore list). When nil, SimpleWordRepeatChecker
// fails closed (no soft case-sensitive invent).
var PreferredWordRepeatFactory func(ruleID string) SentenceChecker

// SimpleWordRepeatChecker returns the faithful WordRepeatRule path when wired.
// Soft case-sensitive surface invent was removed.
func SimpleWordRepeatChecker(ruleID string) SentenceChecker {
	if PreferredWordRepeatFactory != nil {
		return PreferredWordRepeatFactory(ruleID)
	}
	return func(*AnalyzedSentence) []LocalMatch { return nil }
}

// KnownWordSet builds an IsKnownWord from a set of dictionary forms (case-sensitive).
func KnownWordSet(words ...string) func(string) bool {
	m := map[string]struct{}{}
	for _, w := range words {
		m[w] = struct{}{}
	}
	return func(tok string) bool {
		if _, ok := m[tok]; ok {
			return true
		}
		// soft: lowercase probe
		_, ok := m[strings.ToLower(tok)]
		return ok
	}
}

// SimpleMapSpellerChecker flags letter tokens not in known; optional suggestion map.
// When no map entry exists, soft edit-distance suggestions are taken from known
// (capped dictionary size so demo packs stay cheap).
func SimpleMapSpellerChecker(ruleID string, known map[string]struct{}, suggestions map[string][]string) SentenceChecker {
	isKnown := func(w string) bool {
		if _, ok := known[w]; ok {
			return true
		}
		_, ok := known[strings.ToLower(w)]
		return ok
	}
	return SimplePredicateSpellerChecker(ruleID, isKnown, suggestions, known, nil)
}

// SimplePredicateSpellerChecker flags letter tokens rejected by isKnown.
// nearestKnown is optional (edit-distance peers when non-nil and small).
// suggestFn is optional (e.g. CFSA2 edit-candidate Contains); tried after the map.
func SimplePredicateSpellerChecker(ruleID string, isKnown func(string) bool, suggestions map[string][]string, nearestKnown map[string]struct{}, suggestFn func(string) []string) SentenceChecker {
	if ruleID == "" {
		ruleID = "MORFOLOGIK_RULE"
	}
	if isKnown == nil {
		isKnown = func(string) bool { return true }
	}
	return func(sentence *AnalyzedSentence) []LocalMatch {
		if sentence == nil {
			return nil
		}
		var out []LocalMatch
		for _, tok := range sentence.GetTokensWithoutWhitespace() {
			// Skip pure SENT_START marker (empty token). Do not skip content that
			// also carries SENT_END — Java spell-checks the last word of the sentence.
			if tok == nil || tok.IsSentenceStart() {
				continue
			}
			// multiword chunker / disambiguator IGNORE_SPELLING / IMMUNIZE
			if tok.IsIgnoredBySpeller() || tok.IsImmunized() {
				continue
			}
			w := tok.GetToken()
			// Java SpellingCheckRule.isUrl / isEMail (WordTokenizer)
			if tokenizers.IsURL(w) || tokenizers.IsEMail(w) {
				continue
			}
			if w == "" || !hasLetterLocal(w) {
				continue
			}
			if isKnown(w) {
				continue
			}
			m := LocalMatch{
				FromPos:      tok.GetStartPos(),
				ToPos:        tok.GetEndPos(),
				Message:      "Possible spelling mistake",
				ShortMessage: "Spelling mistake",
				RuleID:       ruleID,
			}
			if suggestions != nil {
				if s, ok := suggestions[w]; ok {
					m.Suggestions = append([]string(nil), s...)
				} else if s, ok := suggestions[strings.ToLower(w)]; ok {
					m.Suggestions = append([]string(nil), s...)
				}
			}
			if len(m.Suggestions) == 0 && suggestFn != nil {
				m.Suggestions = suggestFn(w)
			}
			if len(m.Suggestions) == 0 && nearestKnown != nil {
				m.Suggestions = nearestKnownWords(w, nearestKnown, 2, 5)
			}
			out = append(out, m)
		}
		return out
	}
}

// nearestKnownWords returns up to maxN dictionary words within maxDist edit distance.
func nearestKnownWords(word string, known map[string]struct{}, maxDist, maxN int) []string {
	if word == "" || known == nil || len(known) == 0 || len(known) > 10000 || maxN <= 0 {
		return nil
	}
	type cand struct {
		w    string
		dist int
	}
	var cands []cand
	low := strings.ToLower(word)
	seen := map[string]struct{}{}
	for k := range known {
		kl := strings.ToLower(k)
		if _, ok := seen[kl]; ok {
			continue
		}
		d := runeLevenshtein(low, kl)
		if d > 0 && d <= maxDist {
			seen[kl] = struct{}{}
			cands = append(cands, cand{w: kl, dist: d})
		}
	}
	// sort by distance then alphabetically (stable, no import sort for tiny N)
	for i := 0; i < len(cands); i++ {
		for j := i + 1; j < len(cands); j++ {
			if cands[j].dist < cands[i].dist || (cands[j].dist == cands[i].dist && cands[j].w < cands[i].w) {
				cands[i], cands[j] = cands[j], cands[i]
			}
		}
	}
	if len(cands) > maxN {
		cands = cands[:maxN]
	}
	out := make([]string, 0, len(cands))
	for _, c := range cands {
		out = append(out, c.w)
	}
	return out
}

func runeLevenshtein(a, b string) int {
	ar, br := []rune(a), []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}
	// band optimization for short words
	if absInt(len(ar)-len(br)) > 4 {
		return absInt(len(ar) - len(br))
	}
	prev := make([]int, len(br)+1)
	cur := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		cur[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			cur[j] = m
		}
		prev, cur = cur, prev
	}
	return prev[len(br)]
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// PreferredAvsAnChecker is set by package en (init/register) to the faithful AvsAnRule
// path with DT inject. When nil, SimpleAvsAnChecker fails closed (no soft invent lexicon).
var PreferredAvsAnChecker SentenceChecker

// SimpleAvsAnChecker returns the faithful EN_A_VS_AN checker when en has registered it.
// Soft phonetic invent maps were removed — do not invent a/an exceptions here.
func SimpleAvsAnChecker() SentenceChecker {
	if PreferredAvsAnChecker != nil {
		return PreferredAvsAnChecker
	}
	// Fail closed without wired AvsAnRule (package en sets PreferredAvsAnChecker).
	return func(*AnalyzedSentence) []LocalMatch { return nil }
}

// CorrectTextFromLocalMatches applies first suggestion of each match (byte offsets; ASCII-safe).
// Ports Tools.correctTextFromMatches without importing tools package.
func CorrectTextFromLocalMatches(contents string, matches []LocalMatch) string {
	if len(matches) == 0 {
		return contents
	}
	sb := []byte(contents)
	// sort by FromPos ascending so offset adjustments work left-to-right
	sorted := append([]LocalMatch(nil), matches...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].FromPos < sorted[j].FromPos })
	offset := 0
	for _, rm := range sorted {
		if len(rm.Suggestions) == 0 {
			continue
		}
		repl := rm.Suggestions[0]
		from := rm.FromPos - offset
		to := rm.ToPos - offset
		if from < 0 || to < from || to > len(sb) {
			continue
		}
		sb = append(sb[:from], append([]byte(repl), sb[to:]...)...)
		offset += (to - from) - len(repl)
	}
	return string(sb)
}
