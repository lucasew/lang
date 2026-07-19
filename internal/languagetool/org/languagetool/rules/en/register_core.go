package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegisterCoreEnglishLanguageRules installs shared layout + EN-specific word-repeat + a/an + phrases.
func RegisterCoreEnglishLanguageRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Java English / BritishEnglish.getPriorityForId + filterRuleMatches (hooks from language init).
	// en-GB uses BritishEnglish id2prio (Oxford spelling) then English super.
	if languagetool.EnglishPriorityForIdForCodeHook != nil {
		lt.PriorityForId = languagetool.EnglishPriorityForIdForCodeHook(lt.GetLanguageCode())
	} else if languagetool.EnglishPriorityForIdHook != nil {
		lt.PriorityForId = languagetool.EnglishPriorityForIdHook
	}
	// Java English.getDefaultRulePriorityForStyle() = -50
	lt.DefaultRulePriorityForStyle = -50
	if languagetool.FilterEnglishRuleMatchesHook != nil {
		lt.FilterRuleMatches = languagetool.FilterEnglishRuleMatchesHook
	}
	rules.RegisterSharedLayoutRules(lt, "en")
	wr := NewEnglishWordRepeatRule(map[string]string{"repetition": "Word repetition"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	// Faithful AvsAnRule (official determiner lists); DT inject for untagged AnalyzePlain.
	lt.AddRuleChecker("EN_A_VS_AN", AvsAnSentenceChecker())
	// Soft invent PHRASE_REPLACE / token-sequence packs removed (faithful-port policy).
	// Load official grammar.xml when the full pattern loader is ready — do not invent lists.
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

	// Official simple-replace / diacritics data (embedded from LT replace.txt / diacritics.txt).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	di := NewEnglishDiacriticsRule(nil)
	lt.AddRuleChecker(di.GetID(), rules.AsSentenceCheckerSimple(di.Match))
	// Compounds + proper-noun casing from official compounds.txt / specific_case.txt.
	cr := NewCompoundRule(nil)
	lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	sc := NewEnglishSpecificCaseRule(nil)
	lt.AddRuleChecker(sc.GetID(), rules.AsSentenceCheckerSimple(sc.Match))
	// Contractions spelling + wrong word in context (official data files).
	cs := NewContractionSpellingRule(nil)
	lt.AddRuleChecker(cs.GetID(), rules.AsSentenceCheckerSimple(cs.Match))
	ww := NewEnglishWrongWordInContextRule(nil)
	lt.AddRuleChecker(ww.GetID(), rules.AsSentenceCheckerSimple(ww.Match))
	// Dash compounds (en-dash/em-dash vs hyphen) from compounds.txt patterns.
	dr := NewEnglishDashRule(nil)
	lt.AddRuleChecker(dr.GetID(), rules.AsSentenceCheckerSimple(dr.Match))
	// Regional replace tables (British→American and American→British).
	us := NewAmericanReplaceRule(nil)
	lt.AddRuleChecker(us.GetID(), rules.AsSentenceCheckerSimple(us.Match))
	gb := NewBritishReplaceRule(nil)
	lt.AddRuleChecker(gb.GetID(), rules.AsSentenceCheckerSimple(gb.Match))
	// Style tables: redundancies + plain-English/wordiness (official data).
	rd := NewEnglishRedundancyRule(nil)
	lt.AddRuleChecker(rd.GetID(), rules.AsSentenceCheckerSimple(rd.Match))
	pe := NewEnglishPlainEnglishRule(nil)
	lt.AddRuleChecker(pe.GetID(), rules.AsSentenceCheckerSimple(pe.Match))
	// Mixed apostrophe styles across the document (text-level).
	ap := NewConsistentApostrophesRule(nil)
	lt.AddTextLevelRuleChecker(ap.GetID(), rules.AsTextLevelChecker(ap.MatchList))
	// Coherent spelling of dual-admitted variants (official coherency.txt; text-level).
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))
	// New Zealand regional replace table (en-NZ/replace.txt).
	nz := NewNewZealandReplaceRule(nil)
	lt.AddRuleChecker(nz.GetID(), rules.AsSentenceCheckerSimple(nz.Match))
	// Repeated words with synonym suggestions (official synonyms.txt; text-level).
	rw := NewEnglishRepeatedWordsRule(nil)
	lt.AddTextLevelRuleChecker(rw.GetID(), rules.AsTextLevelChecker(rw.MatchList))
	// EN-specific unpaired brackets/quotes (Java English.getRelevantRules; text-level).
	ub := NewEnglishUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))
	uq := NewEnglishUnpairedQuotesRule(nil)
	lt.AddTextLevelRuleChecker(uq.GetID(), rules.AsTextLevelChecker(uq.MatchList))

	// Java English variants createDefaultSpellingRule / Morfologik*SpellerRule.getId.
	// Prefer CFSA2 hunspell/*.dict when present (same files as Java); else empty map
	// Morfologik shell fails closed (no invent misspell flags).
	code := strings.ToLower(lt.GetLanguageCode())
	ruleID, dictFile := EnglishVariantSpellerMeta(code)
	if p := DiscoverEnglishVariantDictFile(dictFile); p != "" {
		if RegisterBinaryEnglishSpellerID(lt, p, ruleID, nil, nil) {
			return
		}
	}
	var esp *MorfologikVariantSpellerRule
	switch {
	case strings.Contains(code, "gb"):
		esp = NewMorfologikBritishSpellerRule()
	case strings.Contains(code, "-ca") || strings.HasSuffix(code, "_ca"):
		esp = NewMorfologikCanadianSpellerRule()
	case strings.Contains(code, "au"):
		esp = NewMorfologikAustralianSpellerRule()
	case strings.Contains(code, "nz"):
		esp = NewMorfologikNewZealandSpellerRule()
	case strings.Contains(code, "za"):
		esp = NewMorfologikSouthAfricanSpellerRule()
	default:
		esp = NewMorfologikAmericanSpellerRule()
	}
	lt.AddRuleChecker(esp.GetID(), rules.AsSentenceChecker(esp.Match))
}

// RegisterPickyEnglishRules installs Java English picky-level rules (official data only).
// Invent token-sequence packs (alot/irregardless/…) are not registered — use grammar.xml when wired.
func RegisterPickyEnglishRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Official profanity list (Tag.picky in Java English.getRelevantRules).
	pf := NewSimpleReplaceProfanityRule(nil)
	lt.AddRuleChecker(pf.GetID(), rules.AsSentenceCheckerSimple(pf.Match))
	// Variant unit conversion (imperial/US messages).
	usU := NewUnitConversionRuleUS(nil)
	lt.AddRuleChecker(usU.GetID(), rules.AsSentenceCheckerSimple(usU.Match))
	imU := NewUnitConversionRuleImperial(nil)
	lt.AddRuleChecker(imU.GetID(), rules.AsSentenceCheckerSimple(imU.Match))
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

	// Metric unit conversion (Java UnitConversionRule via AbstractUnitConversionRule).
	uc := NewUnitConversionRule(nil)
	lt.AddRuleChecker(uc.GetID(), rules.AsSentenceCheckerSimple(uc.Match))
}
