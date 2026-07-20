package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// ToLocalMatches converts RuleMatch list to cycle-free LocalMatch for JLanguageTool.Check.
// Copies rule ID / category / ITS / description when the rule implements the Java getters.
func ToLocalMatches(ms []*RuleMatch) []languagetool.LocalMatch {
	if len(ms) == 0 {
		return nil
	}
	out := make([]languagetool.LocalMatch, 0, len(ms))
	for _, m := range ms {
		if m == nil {
			continue
		}
		// Java SwissGerman.filterRuleMatches uses sentence.substring(from,to);
		// setOriginalErrorStr fills the same surface when not already set.
		if m.GetOriginalErrorStr() == "" {
			m.SetOriginalErrorStr()
		}
		lm := languagetool.LocalMatch{
			FromPos:      m.GetFromPos(),
			ToPos:        m.GetToPos(),
			Message:      m.GetMessage(),
			ShortMessage: m.GetShortMessage(),
			Suggestions:  append([]string(nil), m.GetSuggestedReplacements()...),
			IssueType:    m.IssueType,
			CategoryID:   m.CategoryID,
			CategoryName: m.CategoryName,
			// Java RuleMatch.getUrl: match URL first, else rule URL.
			URL: m.GetURL(),
			// Surface under marker (Swiss AI ss→ß skip, AdaptSuggestionsFilter, …).
			OriginalErrorStr: m.GetOriginalErrorStr(),
			// Sentence-relative span for CleanOverlapping isPunctuationOnlyChange.
			FromPosSentence: m.GetFromPosSentence(),
			ToPosSentence:   m.GetToPosSentence(),
		}
		if m.Sentence != nil {
			lm.SentenceText = m.Sentence.GetText()
		}
		// Java JSON/API use RuleMatch.getSpecificRuleId() as the public rule id
		// (specificRuleId when set, else rule.getId()) — e.g. DE_REPEATEDWORDS_AUSSERDEM.
		lm.RuleID = m.GetSpecificRuleId()
		if g, ok := m.Rule.(interface{ GetDescription() string }); ok {
			lm.Description = g.GetDescription()
		}
		if g, ok := m.Rule.(interface{ GetCategory() *Category }); ok {
			if cat := g.GetCategory(); cat != nil {
				if lm.CategoryID == "" {
					lm.CategoryID = cat.GetID().String()
				}
				if lm.CategoryName == "" {
					lm.CategoryName = cat.GetName()
				}
			}
		}
		if g, ok := m.Rule.(interface{ GetLocQualityIssueType() ITSIssueType }); ok {
			if it := g.GetLocQualityIssueType(); it != "" && lm.IssueType == "" {
				lm.IssueType = string(it)
			}
		}
		// Java RuleMatch.getUrl falls back to rule.getUrl() when match URL empty.
		if lm.URL == "" {
			if g, ok := m.Rule.(interface{ GetURL() string }); ok {
				lm.URL = g.GetURL()
			}
		}
		// Rule.getTags → LocalMatch.Tags (JSON rule.tags) + IsPicky for demotion/merge.
		if g, ok := m.Rule.(interface{ GetTags() []Tag }); ok {
			if tags := g.GetTags(); len(tags) > 0 {
				lm.Tags = make([]string, len(tags))
				for i, t := range tags {
					lm.Tags[i] = string(t)
					if t == TagPicky {
						lm.IsPicky = true
					}
				}
			}
		} else if g, ok := m.Rule.(interface{ HasTag(Tag) bool }); ok {
			lm.IsPicky = g.HasTag(TagPicky)
			if lm.IsPicky {
				lm.Tags = []string{string(TagPicky)}
			}
		}
		// Java Rule.isIncludedInErrorsCorrectedAllAtOnce for punctuation-only overlap preference.
		if g, ok := m.Rule.(interface{ IsIncludedInErrorsCorrectedAllAtOnce() bool }); ok {
			lm.IncludedInErrorsCorrectedAllAtOnce = g.IsIncludedInErrorsCorrectedAllAtOnce()
		}
		// Premium flag: rule method, else DefaultPremium (Java Premium.get().isPremiumRule),
		// else RuleID contains "PREMIUM" for LocalMatch-only paths without a Premium registry.
		if g, ok := m.Rule.(interface{ IsPremium() bool }); ok {
			lm.IsPremium = g.IsPremium()
		} else if g, ok := m.Rule.(interface{ GetPremium() bool }); ok {
			lm.IsPremium = g.GetPremium()
		}
		if !lm.IsPremium && lm.RuleID != "" {
			if languagetool.DefaultPremium != nil && languagetool.DefaultPremium.IsPremiumRule(lm.RuleID) {
				lm.IsPremium = true
			} else if strings.Contains(lm.RuleID, "PREMIUM") {
				lm.IsPremium = true
			}
		}
		// Java Rule.estimateContextForSureMatch (default 0; TextLevelRuleBase → -1).
		if g, ok := m.Rule.(interface{ EstimateContextForSureMatch() int }); ok {
			lm.EstimateContextForSureMatch = g.EstimateContextForSureMatch()
		}
		// Tone tags + goalSpecific for isRuleActiveForLevelAndToneTags.
		if g, ok := m.Rule.(interface {
			GetToneTags() []languagetool.ToneTag
		}); ok {
			if tags := g.GetToneTags(); len(tags) > 0 {
				lm.ToneTags = append([]languagetool.ToneTag(nil), tags...)
			}
		}
		if g, ok := m.Rule.(interface{ IsGoalSpecific() bool }); ok {
			lm.GoalSpecific = g.IsGoalSpecific()
		}
		// RuleMeta fallback when rule getters left category/ITS empty (CLI/API parity).
		// Only apply known Java families — skip uncategorized invent for unknown IDs.
		if lm.RuleID != "" && (lm.CategoryID == "" || lm.IssueType == "" || lm.Description == "" || lm.ShortMessage == "") {
			catID, catName, issue, short := languagetool.RuleMeta(lm.RuleID)
			if issue != "" && issue != "uncategorized" {
				if lm.CategoryID == "" {
					lm.CategoryID = catID
				}
				if lm.CategoryName == "" {
					lm.CategoryName = catName
				}
				if lm.IssueType == "" {
					lm.IssueType = issue
				}
				if lm.ShortMessage == "" && short != "" {
					lm.ShortMessage = short
				}
				if lm.Description == "" {
					if d := languagetool.RuleDescription(lm.RuleID); d != "" && d != lm.RuleID {
						lm.Description = d
					}
				}
			}
		}
		out = append(out, lm)
	}
	return out
}

// AsSentenceChecker wraps a Match(sentence)([]*RuleMatch, error) as SentenceChecker.
func AsSentenceChecker(match func(*languagetool.AnalyzedSentence) ([]*RuleMatch, error)) languagetool.SentenceChecker {
	return func(s *languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		if match == nil {
			return nil
		}
		ms, err := match(s)
		if err != nil {
			return nil
		}
		return ToLocalMatches(ms)
	}
}

// AsSentenceCheckerSimple wraps Match(sentence) []*RuleMatch (no error).
func AsSentenceCheckerSimple(match func(*languagetool.AnalyzedSentence) []*RuleMatch) languagetool.SentenceChecker {
	return func(s *languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		if match == nil {
			return nil
		}
		return ToLocalMatches(match(s))
	}
}

// AsTextLevelChecker wraps MatchList(sentences) []*RuleMatch as TextLevelChecker.
// Offsets are already document-relative (rules accumulate GetCorrectedTextLength).
// Java TextLevelRule.estimateContextForSureMatch is always -1 — apply when the rule
// object did not already set a value via EstimateContextForSureMatch().
func AsTextLevelChecker(matchList func([]*languagetool.AnalyzedSentence) []*RuleMatch) languagetool.TextLevelChecker {
	return func(sents []*languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		if matchList == nil {
			return nil
		}
		ms := ToLocalMatches(matchList(sents))
		for i := range ms {
			// Default Java TextLevelRule → -1 when rule left zero (sentence-rule default).
			// Rules that implement EstimateContextForSureMatch already filled the field.
			if ms[i].EstimateContextForSureMatch == 0 {
				// Distinguish unset 0 from explicit 0: text-level Java always returns -1.
				ms[i].EstimateContextForSureMatch = -1
			}
		}
		return ms
	}
}

// SharedLayoutOptions customizes RegisterSharedLayoutRules for language modules
// that replace some core layout rules (e.g. German DE_DOUBLE_PUNCTUATION).
// Zero value = full shared set (backward compatible).
type SharedLayoutOptions struct {
	// CommaException is CommaWhitespaceRule.IsException (e.g. German ". de-Domain").
	CommaException func(tokens []*languagetool.AnalyzedTokenReadings, tokenIdx int) bool
	// SkipDoublePunctuation skips core DOUBLE_PUNCTUATION (language registers DE_*).
	SkipDoublePunctuation bool
	// SkipSentenceWhitespace skips core SENTENCE_WHITESPACE (language registers DE_*).
	SkipSentenceWhitespace bool
	// SkipWhitespaceBeforePunct skips WHITESPACE_PUNCTUATION when Java language
	// getRelevantRules does not include WhitespaceBeforePunctuationRule (e.g. German).
	SkipWhitespaceBeforePunct bool
	// SkipUnpairedBrackets skips core UNPAIRED_BRACKETS (language registers its own).
	SkipUnpairedBrackets bool
	// SkipParagraphRepeatBeginning skips core PARAGRAPH_REPEAT_BEGINNING when the
	// language registers a language-specific rule (e.g. GermanParagraphRepeatBeginningRule).
	SkipParagraphRepeatBeginning bool
	// UppercaseMatchList, when non-nil, replaces core UppercaseSentenceStartRule.MatchList
	// (same rule ID UPPERCASE_SENTENCE_START; e.g. DE URL via de.UppercaseSentenceStartRule).
	UppercaseMatchList func(sentences []*languagetool.AnalyzedSentence) []*RuleMatch
}

// RegisterSharedLayoutRules installs cross-language layout/punctuation rules.
func RegisterSharedLayoutRules(lt *languagetool.JLanguageTool, uppercaseLang string) {
	RegisterSharedLayoutRulesWithOptions(lt, uppercaseLang, SharedLayoutOptions{})
}

// RegisterSharedLayoutRulesWithCommaException is RegisterSharedLayoutRules with an optional
// CommaWhitespaceRule.IsException hook (e.g. German ". de-Domain" exemption).
func RegisterSharedLayoutRulesWithCommaException(lt *languagetool.JLanguageTool, uppercaseLang string, commaException func(tokens []*languagetool.AnalyzedTokenReadings, tokenIdx int) bool) {
	RegisterSharedLayoutRulesWithOptions(lt, uppercaseLang, SharedLayoutOptions{CommaException: commaException})
}

// RegisterSharedLayoutRulesWithOptions installs shared layout rules with language-specific skips/hooks.
func RegisterSharedLayoutRulesWithOptions(lt *languagetool.JLanguageTool, uppercaseLang string, opt SharedLayoutOptions) {
	if lt == nil {
		return
	}
	if uppercaseLang == "" {
		uppercaseLang = "en"
	}
	ws := NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	cw := NewCommaWhitespaceRule(nil)
	if opt.CommaException != nil {
		cw.IsException = opt.CommaException
	}
	lt.AddRuleChecker(cw.GetID(), AsSentenceCheckerSimple(cw.Match))

	if !opt.SkipDoublePunctuation {
		dp := NewDoublePunctuationRule(nil)
		lt.AddRuleChecker(dp.GetID(), AsSentenceCheckerSimple(dp.Match))
	}

	if !opt.SkipWhitespaceBeforePunct {
		wbp := NewWhitespaceBeforePunctuationRule(map[string]string{
			"no_space_before_colon":     "Don't put a space before the colon",
			"no_space_before_semicolon": "Don't put a space before the semicolon",
		})
		lt.AddRuleChecker(wbp.GetID(), AsSentenceCheckerSimple(wbp.Match))
	}

	wpb := NewWhiteSpaceAtBeginOfParagraph(map[string]string{
		"whitespace_at_begin_parapgraph_msg": "Don't start a paragraph with whitespace",
	})
	lt.AddRuleChecker(wpb.GetID(), AsSentenceCheckerSimple(wpb.Match))
	// Java WhiteSpaceAtBeginOfParagraph default ctor: setDefaultOff().
	if wpb.IsDefaultOff() {
		lt.MarkDefaultOff(wpb.GetID())
	}

	if opt.UppercaseMatchList != nil {
		upID := "UPPERCASE_SENTENCE_START"
		lt.AddRuleChecker(upID, AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*RuleMatch {
			return opt.UppercaseMatchList([]*languagetool.AnalyzedSentence{s})
		}))
	} else {
		up := NewUppercaseSentenceStartRule(nil, uppercaseLang)
		lt.AddRuleChecker(up.GetID(), AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*RuleMatch {
			return up.MatchList([]*languagetool.AnalyzedSentence{s})
		}))
	}

	if !opt.SkipUnpairedBrackets {
		// inject unpaired brackets (GenericUnpairedBracketsRule is multi-sentence)
		lt.AddRuleChecker("UNPAIRED_BRACKETS", languagetool.SimpleUnpairedBracketsChecker())
	}

	if !opt.SkipSentenceWhitespace {
		// text-level: missing space between sentences ("end.Start")
		sw := NewSentenceWhitespaceRule(map[string]string{
			"addSpaceBetweenSentences": "Add a space between sentences",
		})
		lt.AddTextLevelRuleChecker(sw.GetID(), AsTextLevelChecker(sw.MatchList))
	}

	// text-level: long paragraph (default 150 words, Java-ish)
	// Java LongParagraphRule: setDefaultOff() + Tag.picky.
	lp := NewLongParagraphRule(map[string]string{
		"long_paragraph_rule_msg": "This paragraph is too long (%d words)",
	}, 150)
	lt.AddTextLevelRuleChecker(lp.GetID(), AsTextLevelChecker(lp.MatchList))
	if lp.IsDefaultOff() {
		lt.MarkDefaultOff(lp.GetID())
	}

	if !opt.SkipParagraphRepeatBeginning {
		// text-level: successive paragraphs starting with the same word
		prb := NewParagraphRepeatBeginningRule(map[string]string{
			"repetition_paragraph_beginning_last_msg": "Paragraphs should not begin with the same words",
		})
		lt.AddTextLevelRuleChecker(prb.GetID(), AsTextLevelChecker(prb.MatchList))
		if prb.IsDefaultOff() {
			lt.MarkDefaultOff(prb.GetID())
		}
	}

	// text-level: trailing whitespace before paragraph end
	// Java WhiteSpaceBeforeParagraphEnd default ctor: setDefaultOff().
	wpe := NewWhiteSpaceBeforeParagraphEnd(map[string]string{
		"whitespace_before_parapgraph_end_msg": "Don't end a paragraph with whitespace",
	})
	lt.AddTextLevelRuleChecker(wpe.GetID(), AsTextLevelChecker(wpe.MatchList))
	if wpe.IsDefaultOff() {
		lt.MarkDefaultOff(wpe.GetID())
	}

	// text-level: empty line (extra blank line between paragraphs)
	// Java EmptyLineRule default ctor: setDefaultOff().
	el := NewEmptyLineRule(map[string]string{
		"empty_line_rule_msg": "Empty line",
	})
	lt.AddTextLevelRuleChecker(el.GetID(), AsTextLevelChecker(el.MatchList))
	if el.IsDefaultOff() {
		lt.MarkDefaultOff(el.GetID())
	}

	// text-level: missing punctuation at paragraph end
	// Java PunctuationMarkAtParagraphEnd default ctor: setDefaultOff() + Tag.picky.
	ppe := NewPunctuationMarkAtParagraphEnd(map[string]string{
		"punctuation_mark_paragraph_end_msg": "Add a punctuation mark at paragraph end",
	})
	lt.AddTextLevelRuleChecker(ppe.GetID(), AsTextLevelChecker(ppe.MatchList))
	if ppe.IsDefaultOff() {
		lt.MarkDefaultOff(ppe.GetID())
	}
}

// RegisterCoreEnglishRules installs shared layout + EN a/an + word-repeat (real rule).
func RegisterCoreEnglishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	RegisterSharedLayoutRules(lt, "en")

	wr := NewWordRepeatRule(map[string]string{"repetition": "Word repetition"})
	lt.AddRuleChecker(wr.GetID(), AsSentenceCheckerSimple(wr.Match))

	// Full EN AvsAn uses package en (PreferredAvsAnChecker). When en is not imported,
	// skip rather than soft invent (RegisterCoreEnglishLanguageRules wires it).
	if languagetool.PreferredAvsAnChecker != nil {
		lt.AddRuleChecker("EN_A_VS_AN", languagetool.PreferredAvsAnChecker)
	}
	// Soft invent PHRASE_REPLACE packs removed (faithful-port: use grammar.xml when loaded).
}

// RegisterCoreRules picks a language-appropriate core pack (shared layout + base word-repeat).
// EN also gets a/an and phrase injects.
func RegisterCoreRules(lt *languagetool.JLanguageTool, langCode string) {
	if lt == nil {
		return
	}
	base := langCode
	if i := indexByteLocal(langCode, '-'); i > 0 {
		base = langCode[:i]
	}
	switch base {
	case "en":
		RegisterCoreEnglishRules(lt)
	default:
		RegisterSharedLayoutRules(lt, base)
		wr := NewWordRepeatRule(map[string]string{"repetition": "Word repetition"})
		lt.AddRuleChecker(wr.GetID(), AsSentenceCheckerSimple(wr.Match))
	}
}

func indexByteLocal(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// RegisterPatternRule wires a PatternRule into Check (simplified matcher).
func RegisterPatternRule(lt *languagetool.JLanguageTool, match func(*languagetool.AnalyzedSentence) ([]*RuleMatch, error), id string) {
	if lt == nil || match == nil {
		return
	}
	if id == "" {
		id = "PATTERN_RULE"
	}
	lt.AddRuleChecker(id, AsSentenceChecker(match))
}
