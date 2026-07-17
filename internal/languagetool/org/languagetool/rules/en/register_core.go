package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
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

	// Soft grammar patterns (token sequences) until full grammar.xml load is wired.
	patterns.RegisterTokenSequences(lt, "en", []patterns.TokenSequenceSpec{
		{ID: "EN_COULD_OF", Tokens: []string{"could", "of"}, Message: "Did you mean 'could have'?", Suggestion: "could have"},
		{ID: "EN_SHOULD_OF", Tokens: []string{"should", "of"}, Message: "Did you mean 'should have'?", Suggestion: "should have"},
		{ID: "EN_WOULD_OF", Tokens: []string{"would", "of"}, Message: "Did you mean 'would have'?", Suggestion: "would have"},
		{ID: "EN_MUST_OF", Tokens: []string{"must", "of"}, Message: "Did you mean 'must have'?", Suggestion: "must have"},
	})
}

// RegisterDemoEnglishSpeller installs a map-backed MORFOLOGIK_RULE_EN_US inject.
// known may be nil (no-op). Soft stand-in until binary dictionaries are ported.
func RegisterDemoEnglishSpeller(lt *languagetool.JLanguageTool, known map[string]struct{}, suggestions map[string][]string) {
	if lt == nil || known == nil {
		return
	}
	lt.AddRuleChecker("MORFOLOGIK_RULE_EN_US", languagetool.SimpleMapSpellerChecker("MORFOLOGIK_RULE_EN_US", known, suggestions))
}

// DemoEnglishKnownWords is a tiny inject dictionary for smoke/demo checks.
func DemoEnglishKnownWords() map[string]struct{} {
	words := []string{
		"I", "you", "he", "she", "it", "we", "they", "a", "an", "the", "is", "are", "was", "were",
		"to", "of", "and", "in", "on", "for", "with", "this", "that", "have", "has", "had",
		"could", "should", "would", "must", "done", "better", "test", "hello", "world",
		"LanguageTool", "English", "sentence", "word", "Galaxy", "Guide", "like", "so",
	}
	m := make(map[string]struct{}, len(words)*2)
	for _, w := range words {
		m[w] = struct{}{}
		m[strings.ToLower(w)] = struct{}{}
	}
	return m
}
