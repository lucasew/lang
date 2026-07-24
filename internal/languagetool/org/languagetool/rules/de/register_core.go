package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// germanVariant returns AT/CH/DE from language code (e.g. de-AT → AT).
func germanVariant(langCode string) string {
	lc := strings.ToLower(langCode)
	switch {
	case strings.Contains(lc, "-at") || strings.HasSuffix(lc, "_at"):
		return "AT"
	case strings.Contains(lc, "-ch") || strings.HasSuffix(lc, "_ch"):
		return "CH"
	default:
		return "DE"
	}
}

// RegisterCoreGermanRules ports German.getRelevantRules (+ variant speller/compounds).
// Variant (de-AT / de-CH) selects speller, compound rule, and old-spelling like Java GermanyGerman /
// AustrianGerman / SwissGerman. No invent SharedLayoutRules helper.
func RegisterCoreGermanRules(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	variant := "DE"
	if lt != nil {
		variant = germanVariant(lt.GetLanguageCode())
	}
	// Java German / SimpleGerman.getPriorityForId + filterRuleMatches on Check output.
	// de-DE-x-simple-language uses SimpleGerman overrides then German super.
	lt.PriorityForId = language.GermanPriorityForIdForCode(lt.GetLanguageCode())
	if variant == "CH" {
		// SwissGerman.filterRuleMatches: super (German) then ß→ss on suggestions.
		lt.FilterRuleMatches = language.FilterSwissGermanRuleMatches
	} else {
		lt.FilterRuleMatches = language.FilterGermanRuleMatches
	}
	// Java German.getIgnoredCharactersRegex → soft hyphen U+00AD stripped per token.
	lt.IgnoredCharacters = languagetool.GermanIgnoredCharactersRegex
	// Process-wide GermanTagger.INSTANCE for InsertComma / UppercaseNoun filters.
	WireGermanTaggerDefaults()
	// Spelling dict + multitoken + synthesizer + disambiguator for grammar filters /
	// pattern match (Java Language defaults). Fail-closed without resources.
	WireGermanRuntimeResourcesFor(lt, variant)
	// Official grammar.xml / style.xml when UseUpstreamGrammar (default on).
	WireGermanUpstreamGrammar(lt)
	// Java RemoteRuleFilters.load(de) — pattern XML when present (fail-closed if missing).
	WireGermanRemoteRuleFilters()

	// Java German.getRelevantRules layout only (no invent SharedLayout helper / no
	// WHITESPACE_PUNCTUATION / no core DOUBLE_PUNCTUATION / SENTENCE_WHITESPACE /
	// UNPAIRED_BRACKETS / PARAGRAPH_REPEAT_BEGINNING — DE-specific rules below).
	deComma := NewGermanCommaWhitespaceRule(nil)
	lt.AddRuleChecker(deComma.GetID(), rules.AsSentenceCheckerSimple(deComma.Match))

	// Unpaired brackets/quotes registered later with DE-specific rules (same Java list).

	deUpper := NewUppercaseSentenceStartRule(nil)
	lt.AddRuleChecker(deUpper.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return deUpper.MatchList([]*languagetool.AnalyzedSentence{s})
	}))

	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))

	wpe := rules.NewWhiteSpaceBeforeParagraphEnd(map[string]string{
		"whitespace_before_parapgraph_end_msg": "Don't end a paragraph with whitespace",
	})
	lt.AddTextLevelRuleChecker(wpe.GetID(), rules.AsTextLevelChecker(wpe.MatchList))
	if wpe.IsDefaultOff() {
		lt.MarkDefaultOff(wpe.GetID())
	}

	wpb := rules.NewWhiteSpaceAtBeginOfParagraph(map[string]string{
		"whitespace_at_begin_parapgraph_msg": "Don't start a paragraph with whitespace",
	})
	lt.AddRuleChecker(wpb.GetID(), rules.AsSentenceCheckerSimple(wpb.Match))
	if wpb.IsDefaultOff() {
		lt.MarkDefaultOff(wpb.GetID())
	}

	el := rules.NewEmptyLineRule(map[string]string{"empty_line_rule_msg": "Empty line"})
	lt.AddTextLevelRuleChecker(el.GetID(), rules.AsTextLevelChecker(el.MatchList))
	if el.IsDefaultOff() {
		lt.MarkDefaultOff(el.GetID())
	}

	// Java LongParagraphRule(messages, this, userConfig) → maxWords 220, always setDefaultOff.
	lp := rules.NewLongParagraphRule(map[string]string{
		"long_paragraph_rule_msg": "This paragraph is too long (%d words)",
	}, 220)
	lt.AddTextLevelRuleChecker(lp.GetID(), rules.AsTextLevelChecker(lp.MatchList))
	if lp.IsDefaultOff() {
		lt.MarkDefaultOff(lp.GetID())
	}

	// Java PunctuationMarkAtParagraphEnd(messages, this) → defaultActive true (on).
	ppe := rules.NewPunctuationMarkAtParagraphEnd(map[string]string{
		"punctuation_mark_paragraph_end_msg": "Add a punctuation mark at paragraph end",
	})
	ppe.DefaultOff = false
	lt.AddTextLevelRuleChecker(ppe.GetID(), rules.AsTextLevelChecker(ppe.MatchList))

	wr := NewGermanWordRepeatRule(map[string]string{"repetition": "Wortwiederholung"})
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))

	wrb := NewGermanWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_adv":       "Drei aufeinanderfolgende Sätze beginnen mit demselben Adverb.",
		"desc_repetition_beginning_word":      "Drei aufeinanderfolgende Sätze beginnen mit demselben Wort.",
		"desc_repetition_beginning_thesaurus": "Erwägen Sie ein Synonym.",
	})
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))

	ls := NewLongSentenceRule(map[string]string{
		"long_sentence_rule_msg2": "Dieser Satz ist zu lang (%d Wörter)",
	}, 40)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))

	// No soft invent token sequences (e.g. DE_WEGEN_DEM). Java covers wegen/trotz via
	// official grammar.xml pattern rules when grammar is loaded — incomplete without XML, not invented.

	// Official replace.txt / replace_custom.txt + coherency.txt (vendored/embedded).
	sr := NewSimpleReplaceRule(nil)
	lt.AddRuleChecker(sr.GetID(), rules.AsSentenceCheckerSimple(sr.Match))
	wc := NewWordCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(wc.GetID(), rules.AsTextLevelChecker(wc.MatchList))
	// Compounds: Java NonSwissGerman/GermanyGerman/AT → GermanCompoundRule; CH → SwissCompoundRule only.
	if variant == "CH" {
		sw := NewSwissCompoundRule(nil)
		lt.AddRuleChecker(sw.GetID(), rules.AsSentenceCheckerSimple(sw.Match))
	} else {
		cr := NewGermanCompoundRule(nil)
		lt.AddRuleChecker(cr.GetID(), rules.AsSentenceCheckerSimple(cr.Match))
	}
	rw := NewGermanRepeatedWordsRule(nil)
	lt.AddTextLevelRuleChecker(rw.GetID(), rules.AsTextLevelChecker(rw.MatchList))
	cc := NewCompoundCoherencyRule(nil)
	lt.AddTextLevelRuleChecker(cc.GetID(), rules.AsTextLevelChecker(cc.MatchList))

	// Official wrong-word-in-context + dash compounds.
	ww := NewGermanWrongWordInContextRule(nil)
	lt.AddRuleChecker(ww.GetID(), rules.AsSentenceCheckerSimple(ww.Match))
	dr := NewDashRule(nil)
	lt.AddRuleChecker(dr.GetID(), rules.AsSentenceCheckerSimple(dr.Match))

	// Case rule (Wire attaches tagger lookup when resources present), metric units.
	// Morph/POS only; untagged Check fails closed (no surface invent).
	cas := WireCaseRule(nil)
	lt.AddRuleChecker(cas.GetID(), rules.AsSentenceCheckerSimple(cas.Match))
	uc := NewUnitConversionRule(nil)
	lt.AddRuleChecker(uc.GetID(), rules.AsSentenceCheckerSimple(uc.Match))

	// Speller: Java GermanyGerman / AustrianGerman / SwissGerman variants.
	// Match silent empty without resources (Java hunspell null).
	var sp *GermanSpellerRule
	switch variant {
	case "AT":
		sp = NewAustrianGermanSpellerRule(nil).GermanSpellerRule
	case "CH":
		sp = NewSwissGermanSpellerRule(nil).GermanSpellerRule
	default:
		sp = NewGermanSpellerRule(nil)
	}
	_ = sp.InitFromDiscoveredResources()
	lt.AddRuleChecker(sp.GetID(), rules.AsSentenceCheckerSimple(sp.Match))

	// Agreement stack (Wire* attaches synth/tagger hooks when resources exist).
	agr := WireAgreementRule(nil)
	lt.AddRuleChecker(agr.GetID(), rules.AsSentenceCheckerSimple(agr.Match))
	agr2 := NewAgreementRule2(nil)
	if agr.Synth != nil {
		agr2 = agr2.WithSynth(agr.Synth)
	}
	lt.AddRuleChecker(agr2.GetID(), rules.AsSentenceCheckerSimple(agr2.Match))
	sva := WireSubjectVerbAgreementRule(nil)
	lt.AddRuleChecker(sva.GetID(), rules.AsSentenceCheckerSimple(sva.Match))
	va := WireVerbAgreementRule(nil)
	// Java VerbAgreementRule is a text-level rule (match(List)); register once under DE_VERBAGREEMENT.
	lt.AddTextLevelRuleChecker(va.GetID(), rules.AsTextLevelChecker(va.MatchList))

	// MissingVerbRule (Java default off).
	mv := WireMissingVerbRule(NewMissingVerbRule(nil))
	lt.AddRuleChecker(mv.GetID(), rules.AsSentenceCheckerSimple(mv.Match))
	if mv.DefaultOff {
		lt.MarkDefaultOff(mv.GetID())
	}

	// Style statistic rules (Java default off; text-level).
	// Match with MinPercent 0 for all hits when enabled; default limits via *WithDefaultLimit.
	ps := NewPassiveSentenceRuleWithDefaultLimit(nil)
	lt.AddTextLevelRuleChecker(ps.GetID(), rules.AsTextLevelChecker(ps.MatchList))
	lt.MarkDefaultOff(ps.GetID())
	ns := NewNonSignificantVerbsRuleWithDefaultLimit(nil)
	lt.AddTextLevelRuleChecker(ns.GetID(), rules.AsTextLevelChecker(ns.MatchList))
	lt.MarkDefaultOff(ns.GetID())
	man := NewSentenceWithManRuleWithDefaultLimit(nil)
	lt.AddTextLevelRuleChecker(man.GetID(), rules.AsTextLevelChecker(man.MatchList))
	lt.MarkDefaultOff(man.GetID())
	modalStyle := NewSentenceWithModalVerbRuleWithDefaultLimit(nil)
	lt.AddTextLevelRuleChecker(modalStyle.GetID(), rules.AsTextLevelChecker(modalStyle.MatchList))
	lt.MarkDefaultOff(modalStyle.GetID())
	conjBegin := NewConjunctionAtBeginOfSentenceRuleWithDefaultLimit(nil)
	lt.AddTextLevelRuleChecker(conjBegin.GetID(), rules.AsTextLevelChecker(conjBegin.MatchList))
	lt.MarkDefaultOff(conjBegin.GetID())

	// Unpaired brackets/quotes (Java German.getRelevantRules; text-level).
	ub := NewGermanUnpairedBracketsRule(nil)
	lt.AddTextLevelRuleChecker(ub.GetID(), rules.AsTextLevelChecker(ub.MatchList))
	uq := NewGermanUnpairedQuotesRule(nil)
	lt.AddTextLevelRuleChecker(uq.GetID(), rules.AsTextLevelChecker(uq.MatchList))

	// Redundant modal/aux (Java default off).
	rm := NewRedundantModalOrAuxiliaryVerb(nil)
	lt.AddRuleChecker(rm.GetID(), rules.AsSentenceCheckerSimple(rm.Match))
	if rm.DefaultOff {
		lt.MarkDefaultOff(rm.GetID())
	}

	// Old spelling (alt_neu.csv). Austria: Geschoß remains acceptable (NewOldSpellingRuleAT).
	var osr *OldSpellingRule
	if variant == "AT" {
		osr = NewOldSpellingRuleAT(nil)
	} else {
		osr = NewOldSpellingRule(nil)
	}
	lt.AddRuleChecker(osr.GetID(), rules.AsSentenceCheckerSimple(osr.Match))

	// Style repeated words across sentences (Java default off).
	styleRep := NewGermanStyleRepeatedWordRule(nil)
	lt.AddTextLevelRuleChecker(styleRep.GetID(), rules.AsTextLevelChecker(styleRep.MatchList))
	lt.MarkDefaultOff(styleRep.GetID())

	// Style too-often-used (Java default off; default 5% / 100 words).
	tooNoun := NewStyleTooOftenUsedNounRuleWithDefaultLimit(nil)
	lt.AddTextLevelRuleChecker(tooNoun.GetID(), rules.AsTextLevelChecker(tooNoun.MatchList))
	lt.MarkDefaultOff(tooNoun.GetID())
	tooVerb := NewStyleTooOftenUsedVerbRuleWithDefaultLimit(nil)
	lt.AddTextLevelRuleChecker(tooVerb.GetID(), rules.AsTextLevelChecker(tooVerb.MatchList))
	lt.MarkDefaultOff(tooVerb.GetID())
	tooAdj := NewStyleTooOftenUsedAdjectiveRuleWithDefaultLimit(nil)
	lt.AddTextLevelRuleChecker(tooAdj.GetID(), rules.AsTextLevelChecker(tooAdj.MatchList))
	lt.MarkDefaultOff(tooAdj.GetID())
	repBegin := NewStyleRepeatedSentenceBeginning(nil)
	lt.AddTextLevelRuleChecker(repBegin.GetID(), rules.AsTextLevelChecker(repBegin.MatchList))
	lt.MarkDefaultOff(repBegin.GetID())

	// ProhibitedCompoundRule is getRelevantLanguageModelRules (needs LM) — not invent here.

	// Du/du casing consistency (text-level).
	duRule := NewDuUpperLowerCaseRule(nil)
	lt.AddTextLevelRuleChecker(duRule.GetID(), rules.AsTextLevelChecker(duRule.MatchList))

	// Wieder vs wider (spiegeln).
	wvw := NewWiederVsWiderRule(nil)
	lt.AddRuleChecker(wvw.GetID(), rules.AsSentenceCheckerSimple(wvw.Match))

	// Similar names (Java default off).
	sn := NewSimilarNameRule(nil)
	lt.AddTextLevelRuleChecker(sn.GetID(), rules.AsTextLevelChecker(sn.MatchList))
	lt.MarkDefaultOff(sn.GetID())

	// Unnecessary phrases (Java default off).
	up := NewUnnecessaryPhraseRuleWithDefaultLimit(nil)
	lt.AddTextLevelRuleChecker(up.GetID(), rules.AsTextLevelChecker(up.MatchList))
	lt.MarkDefaultOff(up.GetID())

	// Paragraph beginning repeat (Java ParagraphRepeatBeginningRule setDefaultOff).
	pbr := NewGermanParagraphRepeatBeginningRule(nil)
	lt.AddTextLevelRuleChecker(pbr.GetID(), rules.AsTextLevelChecker(pbr.MatchList))
	if pbr.ParagraphRepeatBeginningRule != nil && pbr.IsDefaultOff() {
		lt.MarkDefaultOff(pbr.GetID())
	}

	// DE-specific sentence whitespace / double punctuation (Java IDs only — not core twins).
	deSW := NewSentenceWhitespaceRule(nil)
	lt.AddTextLevelRuleChecker(deSW.GetID(), rules.AsTextLevelChecker(deSW.MatchList))
	deDP := NewGermanDoublePunctuationRule(nil)
	lt.AddRuleChecker(deDP.GetID(), rules.AsSentenceCheckerSimple(deDP.Match))
	// COMMA_WHITESPACE uses shared registration with GermanCommaWhitespaceRule.IsException.

	// Filler words (Java AbstractStatisticStyleRule default off; DEFAULT_MIN_PERCENT=8).
	filler := NewGermanFillerWordsRuleWithDefaultLimit(nil)
	lt.AddTextLevelRuleChecker(filler.GetID(), rules.AsTextLevelChecker(filler.MatchList))
	lt.MarkDefaultOff(filler.GetID())

	// Staccato short sentences (Java default off).
	shortSents := NewStyleRepeatedVeryShortSentences(nil)
	lt.AddTextLevelRuleChecker(shortSents.GetID(), rules.AsTextLevelChecker(shortSents.MatchList))
	lt.MarkDefaultOff(shortSents.GetID())

	// GermanConfusionProbabilityRule / UpperCaseNgramRule are getRelevantLanguageModel* — not invent here.

	// Readability (Java default off): easy + difficult variants.
	readEasy := NewGermanReadabilityRule(nil, true)
	lt.AddTextLevelRuleChecker(readEasy.GetID(), rules.AsTextLevelChecker(readEasy.MatchList))
	lt.MarkDefaultOff(readEasy.GetID())
	readDiff := NewGermanReadabilityRule(nil, false)
	lt.AddTextLevelRuleChecker(readDiff.GetID(), rules.AsTextLevelChecker(readDiff.MatchList))
	lt.MarkDefaultOff(readDiff.GetID())

	// CompoundInfinitivRule (Java German.getRelevantRules).
	// Match needs ZUS + VER:INF; untagged input fails closed (no surface invent).
	ci := WireCompoundInfinitivRule(nil)
	lt.AddRuleChecker(ci.GetID(), rules.AsSentenceCheckerSimple(ci.Match))

	// MissingCommaRelativeClauseRule: front + behind (Java both registered).
	// Morph/POS only; without VER tags skip so untagged Check does not invent false positives.
	mcFront := NewMissingCommaRelativeClauseRule(nil)
	lt.AddRuleChecker(mcFront.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		if s == nil || !hasAnyVERTags(s.GetTokensWithoutWhitespace()) {
			return nil
		}
		return mcFront.Match(s)
	}))
	mcBehind := NewMissingCommaRelativeClauseRuleBehind(nil)
	lt.AddRuleChecker(mcBehind.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		if s == nil || !hasAnyVERTags(s.GetTokensWithoutWhitespace()) {
			return nil
		}
		return mcBehind.Match(s)
	}))
}
