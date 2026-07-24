package fa

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCorePersianRules ports Persian.getRelevantRules.
// Java list only — no invent SharedLayout extras (no uppercase/unpaired/empty-line/etc.).
func RegisterCorePersianRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Java: generic CommaWhitespace + DoublePunctuation + MultipleWhitespace.
	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	// Java LongSentenceRule(messages, userConfig, 50) → id TOO_LONG_SENTENCE (not invent FA_).
	ls := rules.NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "This sentence is too long (%d words)",
	}, 50)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))

	// Persian-specific punctuation / repeat / replace.
	pcw := NewPersianCommaWhitespaceRule(nil)
	lt.AddRuleChecker(pcw.GetID(), rules.AsSentenceCheckerSimple(pcw.Match))

	pdp := NewPersianDoublePunctuationRule(nil)
	lt.AddRuleChecker(pdp.GetID(), rules.AsSentenceCheckerSimple(pdp.Match))

	wrb := NewPersianWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "سه جمله پیاپی با یک کلمه آغاز می‌شوند.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	wr := NewPersianWordRepeatRule(map[string]string{"repetition": "تکرار"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	sb := NewPersianSpaceBeforeRule(nil)
	lt.AddRuleChecker(sb.GetID(), rules.AsSentenceCheckerSimple(sb.Match))

	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))
}
