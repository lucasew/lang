package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreEnglishLanguageRules installs shared layout + EN-specific word-repeat + a/an + phrases.
func RegisterCoreEnglishLanguageRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	rules.RegisterSharedLayoutRules(lt, "en")
	wr := NewEnglishWordRepeatRule(map[string]string{"repetition": "Word repetition"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	lt.AddRuleChecker("EN_A_VS_AN", languagetool.SimpleAvsAnChecker())
	lt.AddRuleChecker("PHRASE_REPLACE", languagetool.SimplePhraseReplaceChecker("PHRASE_REPLACE", map[string]string{
		"tot he": "to the",
	}))
	// Multi-sentence: three successive sentences starting with the same word/adverb.
	wrb := NewEnglishWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_adv":       "Three successive sentences begin with the same adverb.",
		"desc_repetition_beginning_word":      "Three successive sentences begin with the same word.",
		"desc_repetition_beginning_thesaurus": "Consider using a thesaurus to find synonyms.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	ls := rules.NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "This sentence is too long (%d words)",
	}, 40)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))
}
