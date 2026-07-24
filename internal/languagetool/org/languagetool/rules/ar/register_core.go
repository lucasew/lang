package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"
	_ "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ar/filters" // RuleFilter init
)

// RegisterCoreArabicRules ports Arabic.getRelevantRules.
// Java list only — no invent SharedLayout extras (no uppercase / empty-line /
// long-paragraph / invent AR_UNPAIRED_BRACKETS / invent TOO_LONG_SENTENCE_AR).
func RegisterCoreArabicRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	// Java SentenceWhitespaceRule → SENTENCE_WHITESPACE (not invent AR_).
	sw := rules.NewSentenceWhitespaceRule(nil)
	lt.AddTextLevelRuleChecker(sw.GetID(), rules.AsTextLevelChecker(sw.MatchList))

	ub := NewUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))

	// Java CommaWhitespaceRule(messages, true) — quotesWhitespaceCheck true (default).
	// Wait: Arabic uses new CommaWhitespaceRule(messages, true) → quotesWhitespace = true.
	// Actually Java: new CommaWhitespaceRule(messages, true) sets quotesWhitespaceCheck = true.
	cw := rules.NewCommaWhitespaceRule(nil)
	cw.QuotesWhitespaceCheck = true
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))

	// Java LongSentenceRule(messages, userConfig, 50) → TOO_LONG_SENTENCE.
	ls := rules.NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "This sentence is too long (%d words)",
	}, 50)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))

	// specific to Arabic (Java order)
	sp := NewArabicHunspellSpellerRule(hunspell.TryOpenFromClasspath(ArabicHunspellDictPath))
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))

	acw := NewArabicCommaWhitespaceRule(nil)
	lt.AddRuleChecker(acw.GetID(), rules.AsSentenceCheckerSimple(acw.Match))

	aqm := NewArabicQuestionMarkWhitespaceRule(nil)
	lt.AddRuleChecker(aqm.GetID(), rules.AsSentenceCheckerSimple(aqm.Match))

	asc := NewArabicSemiColonWhitespaceRule(nil)
	lt.AddRuleChecker(asc.GetID(), rules.AsSentenceCheckerSimple(asc.Match))

	adp := NewArabicDoublePunctuationRule(nil)
	lt.AddRuleChecker(adp.GetID(), rules.AsSentenceCheckerSimple(adp.Match))

	wr := NewArabicWordRepeatRule(map[string]string{"repetition": "تكرار كلمة"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	sr := NewArabicSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))

	adi := NewArabicDiacriticsRule(nil)
	lt.AddRuleChecker(adi.GetID(), rules.AsSentenceCheckerSimple(adi.Match))

	ad := NewArabicDarjaRule(nil)
	lt.AddRuleChecker(ad.GetID(), rules.AsSentenceCheckerSimple(ad.Match))

	ah := NewArabicHomophonesRule(nil)
	lt.AddRuleChecker(ah.GetID(), rules.AsSentenceCheckerSimple(ah.Match))

	ard := NewArabicRedundancyRule(nil)
	lt.AddRuleChecker(ard.GetID(), rules.AsSentenceCheckerSimple(ard.Match))

	wc := NewArabicWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	aw := NewArabicWordinessRule(nil)
	lt.AddRuleChecker(aw.GetID(), rules.AsSentenceCheckerSimple(aw.Match))

	aww := NewArabicWrongWordInContextRule(nil)
	lt.AddRuleChecker(aww.GetID(), rules.AsSentenceCheckerSimple(aww.Match))

	// Java ArabicTransVerbRule — was missing from prior invent registration.
	atv := NewArabicTransVerbRule(nil)
	lt.AddRuleChecker(atv.GetID(), rules.AsSentenceCheckerSimple(atv.Match))

	ai := NewArabicInflectedOneWordReplaceRule(nil)
	lt.AddRuleChecker(ai.GetID(), rules.AsSentenceCheckerSimple(ai.Match))
}
