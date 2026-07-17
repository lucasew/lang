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
	lt.AddRuleChecker("PHRASE_REPLACE", languagetool.SimplePhraseReplaceChecker("PHRASE_REPLACE", SoftEnglishPhraseReplacements()))
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
	patterns.RegisterTokenSequences(lt, "en", SoftEnglishTokenSequences())

	// Official simple-replace / diacritics data (embedded from LT replace.txt / diacritics.txt).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	di := NewEnglishDiacriticsRule(nil)
	lt.AddRuleChecker(di.GetID(), rules.AsSentenceCheckerSimple(di.Match))
}

// SoftEnglishPhraseReplacements is the soft PHRASE_REPLACE map (wrong → fix).
func SoftEnglishPhraseReplacements() map[string]string {
	return map[string]string{
		"tot he":                     "to the",
		"for all intensive purposes": "for all intents and purposes",
		"nip it in the butt":         "nip it in the bud",
		"on accident":                "by accident",
		"could care less":            "couldn't care less",
		"one in the same":            "one and the same",
		"case and point":             "case in point",
		"deep-seeded":                "deep-seated",
		"baited breath":              "bated breath",
		"free reign":                 "free rein",
		"based off of":               "based on",
		"based off":                  "based on",
		"eachother":                  "each other",
		"never the less":             "nevertheless",
		"in regards to":              "with regard to",
		"with regards to":            "with regard to",
		"all of the sudden":          "all of a sudden",
		"by in large":                "by and large",
		"first come first serve":     "first come, first served",
		"I could care less":          "I couldn't care less",
		"mute point":                 "moot point",
		"slight of hand":             "sleight of hand",
		"tow the line":               "toe the line",
		"wait with baited breath":    "wait with bated breath",
		"piece of mind":              "peace of mind",
		"make due":                   "make do",
		"pass mustard":               "pass muster",
		"extract revenge":            "exact revenge",
		"hone in on":                 "home in on",
		"fall by the waste side":     "fall by the wayside",
		"wet your appetite":          "whet your appetite",
		"for all intensive purpose":  "for all intents and purposes",
		"in the same vane":           "in the same vein",
		"statue of limitations":      "statute of limitations",
		"escape goat":                "scapegoat",
		"few and far in between":     "few and far between",
		"chuck full":                 "chock-full",
		"do diligence":               "due diligence",
		"peak my interest":           "pique my interest",
		"reign in":                   "rein in",
		"shoe in":                    "shoo-in",
		"nerve wracking":             "nerve-racking",
		"wait listed":                "wait-listed",
		"second hand smoke":          "secondhand smoke",
		"day to day basis":           "day-to-day basis",
	}
}

// SoftEnglishTokenSequences is the soft modal-of / fixed phrase token pack.
func SoftEnglishTokenSequences() []patterns.TokenSequenceSpec {
	return []patterns.TokenSequenceSpec{
		{ID: "EN_COULD_OF", Tokens: []string{"could", "of"}, Message: "Did you mean 'could have'?", Suggestion: "could have"},
		{ID: "EN_SHOULD_OF", Tokens: []string{"should", "of"}, Message: "Did you mean 'should have'?", Suggestion: "should have"},
		{ID: "EN_WOULD_OF", Tokens: []string{"would", "of"}, Message: "Did you mean 'would have'?", Suggestion: "would have"},
		{ID: "EN_MUST_OF", Tokens: []string{"must", "of"}, Message: "Did you mean 'must have'?", Suggestion: "must have"},
		{ID: "EN_MIGHT_OF", Tokens: []string{"might", "of"}, Message: "Did you mean 'might have'?", Suggestion: "might have"},
		{ID: "EN_TRY_AND", Tokens: []string{"try", "and"}, Message: "Did you mean 'try to'?", Suggestion: "try to"},
		{ID: "EN_SUPPOSE_TO", Tokens: []string{"suppose", "to"}, Message: "Did you mean 'supposed to'?", Suggestion: "supposed to"},
		{ID: "EN_USED_TO_GO", Tokens: []string{"use", "to", "go"}, Message: "Did you mean 'used to go'?", Suggestion: "used to go"},
		{ID: "EN_INTENTS_PURPOSE", Tokens: []string{"intensive", "purposes"}, Message: "Did you mean 'intents and purposes'?", Suggestion: "intents and purposes"},
		{ID: "EN_COULD_CARE", Tokens: []string{"could", "care", "less"}, Message: "Did you mean 'couldn't care less'?", Suggestion: "couldn't care less"},
		{ID: "EN_FOR_ALL_INTENSIVE", Tokens: []string{"for", "all", "intensive", "purposes"}, Message: "Did you mean 'for all intents and purposes'?", Suggestion: "for all intents and purposes"},
	}
}

// RegisterPickyEnglishRules installs extra style/grammar patterns for Level PICKY.
func RegisterPickyEnglishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	patterns.RegisterTokenSequences(lt, "en", []patterns.TokenSequenceSpec{
		{ID: "EN_A_LOT", Tokens: []string{"alot"}, Message: "Did you mean 'a lot'?", Suggestion: "a lot"},
		{ID: "EN_IRREGARDLESS", Tokens: []string{"irregardless"}, Message: "Prefer 'regardless'.", Suggestion: "regardless"},
		{ID: "EN_SUPPOSABLY", Tokens: []string{"supposably"}, Message: "Did you mean 'supposedly'?", Suggestion: "supposedly"},
		{ID: "EN_EXPRESSO", Tokens: []string{"expresso"}, Message: "Did you mean 'espresso'?", Suggestion: "espresso"},
		{ID: "EN_EXCAPE", Tokens: []string{"excape"}, Message: "Did you mean 'escape'?", Suggestion: "escape"},
		{ID: "EN_NUKEULAR", Tokens: []string{"nukeular"}, Message: "Did you mean 'nuclear'?", Suggestion: "nuclear"},
		{ID: "EN_LIBARY", Tokens: []string{"libary"}, Message: "Did you mean 'library'?", Suggestion: "library"},
		{ID: "EN_MISCHIEVOUS", Tokens: []string{"mischievious"}, Message: "Did you mean 'mischievous'?", Suggestion: "mischievous"},
		{ID: "EN_ORIENTATE", Tokens: []string{"orientate"}, Message: "Prefer 'orient' in American English.", Suggestion: "orient"},
		{ID: "EN_PREVENTATIVE", Tokens: []string{"preventative"}, Message: "Prefer 'preventive' in many style guides.", Suggestion: "preventive"},
	})
}

// RegisterDemoEnglishSpeller installs a map-backed MORFOLOGIK_RULE_EN_US inject.
// known may be nil (no-op). Soft stand-in until binary dictionaries are ported.
func RegisterDemoEnglishSpeller(lt *languagetool.JLanguageTool, known map[string]struct{}, suggestions map[string][]string) {
	if lt == nil || known == nil {
		return
	}
	if suggestions == nil {
		suggestions = CommonDemoSpellerSuggestions
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
		// common correction targets for demo edit-distance suggestions
		"receive", "separate", "book", "message", "doctor", "great", "ability", "gift",
		"actual", "library", "eventual", "become", "known", "before", "after", "because",
	}
	m := make(map[string]struct{}, len(words)*2)
	for _, w := range words {
		m[w] = struct{}{}
		m[strings.ToLower(w)] = struct{}{}
	}
	return m
}

// DemoEnglishTagWord returns a tiny closed-class POS inject for smoke tests.
func DemoEnglishTagWord() func(token string) []languagetool.TokenTag {
	m := map[string]languagetool.TokenTag{
		"the": {POS: "DT", Lemma: "the"}, "a": {POS: "DT", Lemma: "a"}, "an": {POS: "DT", Lemma: "an"},
		"is": {POS: "VBZ", Lemma: "be"}, "are": {POS: "VBP", Lemma: "be"}, "was": {POS: "VBD", Lemma: "be"},
		"and": {POS: "CC", Lemma: "and"}, "of": {POS: "IN", Lemma: "of"}, "to": {POS: "TO", Lemma: "to"},
		"I": {POS: "PRP", Lemma: "I"}, "you": {POS: "PRP", Lemma: "you"}, "he": {POS: "PRP", Lemma: "he"},
		"she": {POS: "PRP", Lemma: "she"}, "it": {POS: "PRP", Lemma: "it"}, "we": {POS: "PRP", Lemma: "we"},
		"they": {POS: "PRP", Lemma: "they"},
	}
	return func(token string) []languagetool.TokenTag {
		if tg, ok := m[token]; ok {
			return []languagetool.TokenTag{tg}
		}
		low := strings.ToLower(token)
		if tg, ok := m[low]; ok {
			return []languagetool.TokenTag{tg}
		}
		return nil
	}
}

// RegisterDemoEnglishTagger installs DemoEnglishTagWord on lt.TagWord.
func RegisterDemoEnglishTagger(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.TagWord = DemoEnglishTagWord()
}
