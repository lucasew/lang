package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// ToLocalMatches converts RuleMatch list to cycle-free LocalMatch for JLanguageTool.Check.
func ToLocalMatches(ms []*RuleMatch) []languagetool.LocalMatch {
	if len(ms) == 0 {
		return nil
	}
	out := make([]languagetool.LocalMatch, 0, len(ms))
	for _, m := range ms {
		if m == nil {
			continue
		}
		lm := languagetool.LocalMatch{
			FromPos:     m.GetFromPos(),
			ToPos:       m.GetToPos(),
			Message:     m.GetMessage(),
			ShortMessage: m.GetShortMessage(),
			Suggestions: append([]string(nil), m.GetSuggestedReplacements()...),
		}
		if g, ok := m.Rule.(interface{ GetID() string }); ok {
			lm.RuleID = g.GetID()
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
func AsTextLevelChecker(matchList func([]*languagetool.AnalyzedSentence) []*RuleMatch) languagetool.TextLevelChecker {
	return func(sents []*languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		if matchList == nil {
			return nil
		}
		return ToLocalMatches(matchList(sents))
	}
}

// RegisterSharedLayoutRules installs cross-language layout/punctuation rules.
func RegisterSharedLayoutRules(lt *languagetool.JLanguageTool, uppercaseLang string) {
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
	lt.AddRuleChecker(cw.GetID(), AsSentenceCheckerSimple(cw.Match))

	dp := NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), AsSentenceCheckerSimple(dp.Match))

	wbp := NewWhitespaceBeforePunctuationRule(map[string]string{
		"no_space_before_colon":     "Don't put a space before the colon",
		"no_space_before_semicolon": "Don't put a space before the semicolon",
	})
	lt.AddRuleChecker(wbp.GetID(), AsSentenceCheckerSimple(wbp.Match))

	wpb := NewWhiteSpaceAtBeginOfParagraph(map[string]string{
		"whitespace_at_begin_parapgraph_msg": "Don't start a paragraph with whitespace",
	})
	lt.AddRuleChecker(wpb.GetID(), AsSentenceCheckerSimple(wpb.Match))

	up := NewUppercaseSentenceStartRule(nil, uppercaseLang)
	lt.AddRuleChecker(up.GetID(), AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*RuleMatch {
		return up.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	// inject unpaired brackets (GenericUnpairedBracketsRule is multi-sentence)
	lt.AddRuleChecker("UNPAIRED_BRACKETS", languagetool.SimpleUnpairedBracketsChecker())

	// text-level: missing space between sentences ("end.Start")
	sw := NewSentenceWhitespaceRule(map[string]string{
		"addSpaceBetweenSentences": "Add a space between sentences",
	})
	lt.AddTextLevelRuleChecker(sw.GetID(), AsTextLevelChecker(sw.MatchList))

	// text-level: long paragraph (default 150 words, Java-ish)
	lp := NewLongParagraphRule(map[string]string{
		"long_paragraph_rule_msg": "This paragraph is too long (%d words)",
	}, 150)
	lt.AddTextLevelRuleChecker(lp.GetID(), AsTextLevelChecker(lp.MatchList))

	// text-level: successive paragraphs starting with the same word
	prb := NewParagraphRepeatBeginningRule(map[string]string{
		"repetition_paragraph_beginning_last_msg": "Paragraphs should not begin with the same words",
	})
	lt.AddTextLevelRuleChecker(prb.GetID(), AsTextLevelChecker(prb.MatchList))

	// text-level: trailing whitespace before paragraph end
	wpe := NewWhiteSpaceBeforeParagraphEnd(map[string]string{
		"whitespace_before_parapgraph_end_msg": "Don't end a paragraph with whitespace",
	})
	lt.AddTextLevelRuleChecker(wpe.GetID(), AsTextLevelChecker(wpe.MatchList))

	// text-level: empty line (extra blank line between paragraphs)
	el := NewEmptyLineRule(map[string]string{
		"empty_line_rule_msg": "Empty line",
	})
	lt.AddTextLevelRuleChecker(el.GetID(), AsTextLevelChecker(el.MatchList))

	// text-level: missing punctuation at paragraph end
	ppe := NewPunctuationMarkAtParagraphEnd(map[string]string{
		"punctuation_mark_paragraph_end_msg": "Add a punctuation mark at paragraph end",
	})
	lt.AddTextLevelRuleChecker(ppe.GetID(), AsTextLevelChecker(ppe.MatchList))
}

// RegisterCoreEnglishRules installs shared layout + EN a/an + word-repeat (real rule).
func RegisterCoreEnglishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	RegisterSharedLayoutRules(lt, "en")

	wr := NewWordRepeatRule(map[string]string{"repetition": "Word repetition"})
	lt.AddRuleChecker(wr.GetID(), AsSentenceCheckerSimple(wr.Match))

	lt.AddRuleChecker("EN_A_VS_AN", languagetool.SimpleAvsAnChecker())
	lt.AddRuleChecker("PHRASE_REPLACE", languagetool.SimplePhraseReplaceChecker("PHRASE_REPLACE", map[string]string{
		"tot he": "to the",
	}))
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
