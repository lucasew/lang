package de

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// Wire helpers for AgreementRule / SubjectVerbAgreementRule / VerbAgreementRule
// using discovered DE resources (same discovery as WireAdaptSuggestionFilter).
// Without resources, rules stay fail-closed (nil hooks).

var (
	agreementStackOnce sync.Once
	wiredAgreement     *AgreementRule
	wiredSubjectVerb   *SubjectVerbAgreementRule
	wiredVerbAgreement *VerbAgreementRule
	// compoundSpeller is optional GERMAN_SPELLER_RULE for Java initLt().check gate.
	compoundSpeller *GermanSpellerRule
)

// WireAgreementRule returns AgreementRule with Synth + CompoundPhraseValid when
// german_synth.dict / german.dict (tagger) are discoverable.
func WireAgreementRule(messages map[string]string) *AgreementRule {
	agreementStackOnce.Do(wireAgreementStack)
	if wiredAgreement == nil {
		return NewAgreementRule(messages)
	}
	cp := *wiredAgreement
	if messages != nil {
		cp.Messages = messages
	}
	return &cp
}

// WireSubjectVerbAgreementRule returns SubjectVerbAgreementRule with
// LookupInfinitive from GermanTagger when available.
func WireSubjectVerbAgreementRule(messages map[string]string) *SubjectVerbAgreementRule {
	agreementStackOnce.Do(wireAgreementStack)
	if wiredSubjectVerb == nil {
		return NewSubjectVerbAgreementRule(messages)
	}
	cp := *wiredSubjectVerb
	if messages != nil {
		cp.Messages = messages
	}
	return &cp
}

// WireVerbAgreementRule returns VerbAgreementRule with Synth when available.
func WireVerbAgreementRule(messages map[string]string) *VerbAgreementRule {
	agreementStackOnce.Do(wireAgreementStack)
	if wiredVerbAgreement == nil {
		return NewVerbAgreementRule(messages)
	}
	cp := *wiredVerbAgreement
	if messages != nil {
		cp.Messages = messages
	}
	return &cp
}

func wireAgreementStack() {
	root := DiscoverGermanResourceDir()
	tagger := openDiscoveredGermanTagger(root)

	wiredAgreement = NewAgreementRule(nil)
	wiredSubjectVerb = NewSubjectVerbAgreementRule(nil)
	// Java: new VerbAgreementRule(messages, language) — synthesizer via language.getSynthesizer().
	wiredVerbAgreement = NewVerbAgreementRule(nil)
	if synth := openDiscoveredGermanSynthesizer(); synth != nil {
		wiredVerbAgreement = wiredVerbAgreement.WithSynth(synth)
	}

	// Java initLt: GERMAN_SPELLER_RULE uses de/hunspell/de_DE.dict.
	if p := DiscoverGermanHunspellDict("DE"); p != "" {
		_ = WireGermanFilterSpeller(p)
		sp := NewGermanSpellerRule(nil)
		if err := sp.InitFromDiscoveredResources(); err == nil {
			compoundSpeller = sp
		}
	}

	if tagger != nil {
		// Java: GermanTagger.lookup(lower).hasPosTagStartingWith("VER:INF")
		wiredSubjectVerb.LookupInfinitive = func(lowerWord string) bool {
			return lookupIsInfinitive(tagger, lowerWord)
		}
		// Java: lt.check(phrase) with only DE_AGREEMENT + GERMAN_SPELLER_RULE enabled.
		wiredAgreement.CompoundPhraseValid = func(phrase string) bool {
			return compoundPhraseValid(tagger, phrase)
		}
	}

	// Java language.getSynthesizer() → GermanSynthesizer.INSTANCE
	if gs := openDiscoveredGermanSynthesizer(); gs != nil {
		wiredAgreement.Synth = gs
		wiredVerbAgreement.Synth = gs
	} else if base := openDiscoveredGermanSynthBase(); base != nil {
		wiredAgreement.Synth = base
		wiredVerbAgreement.Synth = base
	}
	_ = root
}

// WireMissingVerbRule attaches TagFirstLowercased from discovered GermanTagger
// (sentence-start uppercase verb workaround — Java language.getTagger().tag).
// Pass existing rule or nil.
func WireMissingVerbRule(r *MissingVerbRule) *MissingVerbRule {
	if r == nil {
		r = NewMissingVerbRule(nil)
	}
	root := DiscoverGermanResourceDir()
	tagger := openDiscoveredGermanTagger(root)
	if tagger != nil {
		// Java verbAtSentenceStart:
		//   List<AnalyzedTokenReadings> lcReadings = language.getTagger().tag(singleton);
		//   return lcReadings.size() > 0 && lcReadings.get(0).hasPosTagStartingWith("VER");
		// GermanTagger.Tag → TagIgnoreCase(..., true) (Java tag(List) default).
		r.TagFirstLowercased = func(lower string) bool {
			lcReadings := tagger.Tag([]string{lower})
			return len(lcReadings) > 0 && lcReadings[0] != nil &&
				lcReadings[0].HasPosTagStartingWith("VER")
		}
	}
	return r
}

// WireCaseRule attaches GermanTagger.Lookup for CaseRule morph path (Java language.getTagger()).
// Without resources Lookup stays nil (hasNounReading / isNumber limited).
func WireCaseRule(messages map[string]string) *CaseRule {
	r := NewCaseRule(messages)
	root := DiscoverGermanResourceDir()
	tagger := openDiscoveredGermanTagger(root)
	if tagger != nil {
		r.Lookup = func(word string) *languagetool.AnalyzedTokenReadings {
			return tagger.Lookup(word)
		}
	}
	// Java: language.getDefaultSpellingRule().isMisspelled(lcWord)
	// FilterDictIsMisspelled is fail-closed (false) without dict; with Wire filter dict matches speller.
	r.IsMisspelled = FilterDictIsMisspelled
	return r
}

// lookupIsInfinitive ports the tagger part of containsOnlyInfinitivesToTheLeft.
func lookupIsInfinitive(tagger *detag.GermanTagger, lowerWord string) bool {
	if tagger == nil || lowerWord == "" {
		return false
	}
	rd := tagger.Lookup(lowerWord)
	if rd == nil {
		return false
	}
	return rd.HasPosTagStartingWith("VER:INF")
}

// compoundPhraseValid ports the lt.check gate for open-compound suggestions
// (Java initLt: only DE_AGREEMENT + GERMAN_SPELLER_RULE enabled).
// Fail-closed without inventing full grammar-check results.
func compoundPhraseValid(tagger *detag.GermanTagger, phrase string) bool {
	if tagger == nil || phrase == "" {
		return false
	}
	// Java: StringUtils.split / space-separated open compounds (ASCII space only).
	parts := splitASCIISpaceOmitEmptyDE(phrase)
	if len(parts) == 0 {
		return false
	}
	sent := compoundPhraseAnalyzed(tagger, parts)
	if sent == nil {
		return false
	}
	// DE_AGREEMENT: phrase must not trigger agreement errors.
	// Fresh rule: no CompoundPhraseValid → open-compound branch stays fail-closed (no recurse).
	if len(NewAgreementRule(nil).Match(sent)) != 0 {
		return false
	}
	// GERMAN_SPELLER_RULE: Java lt.check(phrase).isEmpty().
	// Prefer GermanSpellerRule.Match (sentence path) over raw dict IsMisspelled —
	// Match accepts productive compounds (Magnesiumanteil, Standardpriorität) that
	// bare FSA lookup may reject. Hyphen forms: Java accepts when segments spell OK
	// even when the full hyphen string is unknown to the FSA (LanguageTool-Account).
	if !compoundPhraseSpellingOK(sent, parts) {
		return false
	}
	return true
}

// compoundPhraseAnalyzed tags phrase tokens like JLanguageTool analysis for lt.check.
// Includes space tokens so GetText / start positions match the original phrase.
func compoundPhraseAnalyzed(tagger *detag.GermanTagger, parts []string) *languagetool.AnalyzedSentence {
	if tagger == nil || len(parts) == 0 {
		return nil
	}
	// surfaces: part0, " ", part1, " ", ... (Java tokenizer leaves spaces as tokens).
	surfaces := make([]string, 0, len(parts)*2-1)
	for i, p := range parts {
		if i > 0 {
			surfaces = append(surfaces, " ")
		}
		surfaces = append(surfaces, p)
	}
	tagged := tagger.Tag(surfaces)
	if len(tagged) == 0 {
		return nil
	}
	ss := languagetool.SentenceStartTagName
	toks := make([]*languagetool.AnalyzedTokenReadings, 0, len(tagged)+1)
	toks = append(toks, languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("", &ss, nil), 0,
	))
	positions := tokenizers.BuildPositions(surfaces)
	for i, tr := range tagged {
		w := ""
		if i < len(surfaces) {
			w = surfaces[i]
		}
		start := 0
		if i < len(positions) {
			start = positions[i]
		}
		if tr == nil {
			// untagged token — AgreementRule fails closed on missing POS (no invent)
			tr = languagetool.NewAnalyzedTokenReadingsAt(
				languagetool.NewAnalyzedToken(w, nil, nil), start,
			)
		} else if tr.GetStartPos() != start {
			tr.SetStartPos(start)
		}
		toks = append(toks, tr)
	}
	return languagetool.NewAnalyzedSentence(toks)
}

// compoundPhraseSpellingOK ports GERMAN_SPELLER_RULE half of lt.check(phrase).isEmpty().
// Without a wired speller dict, fail-closed (cannot invent that forms are spelled OK).
func compoundPhraseSpellingOK(sent *languagetool.AnalyzedSentence, parts []string) bool {
	if len(parts) == 0 {
		return false
	}
	if !FilterDictAvailable() {
		// No speller resource → cannot prove Java GERMAN_SPELLER pass; fail-closed.
		return false
	}
	last := parts[len(parts)-1]
	if strings.Contains(last, "-") {
		// Hyphen rewrite: Java accepts LanguageTool-Account / Standard-Priorität when
		// segments are known even if full-form FSA flags the hyphen token.
		for _, seg := range strings.Split(last, "-") {
			if seg == "" {
				continue
			}
			if FilterDictIsMisspelled(seg) {
				return false
			}
		}
		// Non-last words must be spelled correctly (FSA / Match on those tokens).
		for _, w := range parts[:len(parts)-1] {
			if FilterDictIsMisspelled(w) {
				return false
			}
		}
		return true
	}
	// Closed form: GermanSpellerRule.Match on the phrase (Java GERMAN_SPELLER_RULE).
	// Match (not bare IsMisspelled) accepts productive compounds like Magnesiumanteil.
	if compoundSpeller != nil && sent != nil {
		return len(compoundSpeller.Match(sent)) == 0
	}
	// Fallback: last-word FSA (weaker than Match for productive compounds).
	return !FilterDictIsMisspelled(last)
}

// compoundPhraseAgreementOK kept for tests that call it directly.
func compoundPhraseAgreementOK(tagger *detag.GermanTagger, parts []string) bool {
	sent := compoundPhraseAnalyzed(tagger, parts)
	if sent == nil {
		return false
	}
	return len(NewAgreementRule(nil).Match(sent)) == 0
}
