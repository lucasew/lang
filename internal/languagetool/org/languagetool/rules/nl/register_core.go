package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreDutchRules installs Dutch.getRelevantRules surface (class getId parity)
// plus word-repeat helpers used by shared layout languages.
// Soft invent token sequences are not registered (incomplete without grammar.xml).
func RegisterCoreDutchRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Java Dutch.getPriorityForId on Check priorities.
	lt.PriorityForId = language.DutchPriorityForId

	// Shared layout; Dutch GenericUnpairedBracketsRule symbol set replaces simple checker.
	// Dutch SentenceWhitespaceRule messages replace the English shared defaults.
	rules.RegisterSharedLayoutRulesWithOptions(lt, "nl", rules.SharedLayoutOptions{
		SkipUnpairedBrackets:   true,
		SkipSentenceWhitespace: true,
	})
	ubr := NewDutchUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ubr.GetID(), rules.AsTextLevelChecker(ubr.MatchList))
	sw := NewSentenceWhitespaceRule(nil)
	lt.AddTextLevelRuleChecker(sw.GetID(), rules.AsTextLevelChecker(sw.MatchList))

	wr := NewWordRepeatRule(map[string]string{"repetition": "Woordherhaling"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	wrb := NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Drie opeenvolgende zinnen beginnen met hetzelfde woord.",
		"desc_repetition_beginning_adv":  "Drie opeenvolgende zinnen beginnen met hetzelfde bijwoord.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	// Java: new LongSentenceRule(messages, userConfig, 40)
	ls := NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "Deze zin is te lang (%d woorden)",
	}, 40)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))

	// Java: new LongParagraphRule(messages, this, userConfig)
	// → defaultWords=-1, defaultActive=true → maxWords DEFAULT_MAX_WORDS (220), on by default.
	lp := NewLongParagraphRule(map[string]string{
		"long_paragraph_rule_msg": "Deze alinea is te lang (%d woorden)",
	}, 220)
	if lp.LongParagraphRule != nil {
		lp.DefaultOff = false // Java 3-arg ctor: defaultActive=true (not setDefaultOff)
	}
	lt.AddTextLevelRuleChecker(lp.GetID(), rules.AsTextLevelChecker(lp.MatchList))

	// Soft invent token sequences removed (faithful-port): incomplete without grammar.xml, not invented.

	// Official replace + coherency tables (embedded from upstream).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))
	// Official compounds + spaced compound detection.
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	sc := NewSpaceInCompoundRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))

	// Official wrong-word-in-context + check-case.
	ww := NewDutchWrongWordInContextRule(nil)
	lt.AddRuleChecker(ww.GetID(), rules.AsSentenceCheckerSimple(ww.Match))
	cc := NewCheckCaseRule(nil)
	lt.AddRuleChecker(cc.GetID(), rules.AsSentenceCheckerSimple(cc.Match))

	// Java PreferredWordRule
	pw := NewPreferredWordRule(nil)
	lt.AddRuleChecker(pw.GetID(), rules.AsSentenceCheckerSimple(pw.Match))

	// Java createDefaultSpellingRule → MorfologikDutchSpellerRule.
	// Always full Match (compound acceptor + _english_ignore_ SkipTokenFn).
	sp := NewMorfologikDutchSpellerRule()
	// Wire CompoundAcceptor probes (Java DutchTagger + MorfologikDutchSpellerRule).
	// Missing dicts → fail-closed; Accept stays list-only without inventing POS.
	if TryWireDutchFilterSpeller() {
		sp.IsMisspelled = FilterDictIsMisspelled
	}
	_ = TryWireDutchFilterTagger()
	BindDefaultCompoundAcceptorFilters()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceChecker(sp.Match))
}
