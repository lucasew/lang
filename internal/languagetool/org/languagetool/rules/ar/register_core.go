package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// RegisterCoreArabicRules installs shared layout + Arabic word-repeat + beginning.
func RegisterCoreArabicRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "ar")
	wr := NewArabicWordRepeatRule(map[string]string{"repetition": "تكرار كلمة"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "ثلاث جمل متتالية تبدأ بنفس الكلمة.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
	patterns.RegisterTokenSequences(lt, "ar", []patterns.TokenSequenceSpec{
		{ID: "AR_FI_FI", Tokens: []string{"في", "في"}, Message: "تكرار محتمل لحرف الجر «في».", Suggestion: "في"},
	})

	// Official replace + coherency tables (embedded from upstream).
	sr := NewArabicSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewArabicWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))

	// Official AR data tables (diacritics, style, confusion, dialect, inflected).
	adi := NewArabicDiacriticsRule(nil)
	lt.AddRuleChecker(adi.GetID(), rules.AsSentenceCheckerSimple(adi.Match))
	ard := NewArabicRedundancyRule(nil)
	lt.AddRuleChecker(ard.GetID(), rules.AsSentenceCheckerSimple(ard.Match))
	aw := NewArabicWordinessRule(nil)
	lt.AddRuleChecker(aw.GetID(), rules.AsSentenceCheckerSimple(aw.Match))
	aww := NewArabicWrongWordInContextRule(nil)
	lt.AddRuleChecker(aww.GetID(), rules.AsSentenceCheckerSimple(aww.Match))
	ah := NewArabicHomophonesRule(nil)
	lt.AddRuleChecker(ah.GetID(), rules.AsSentenceCheckerSimple(ah.Match))
	ad := NewArabicDarjaRule(nil)
	lt.AddRuleChecker(ad.GetID(), rules.AsSentenceCheckerSimple(ad.Match))
	ai := NewArabicInflectedOneWordReplaceRule(nil)
	lt.AddRuleChecker(ai.GetID(), rules.AsSentenceCheckerSimple(ai.Match))
}
