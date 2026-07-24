package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
)

// Wire helpers for AgreementRule / SubjectVerbAgreementRule / VerbAgreementRule
// using discovered DE resources (same discovery as WireAdaptSuggestionFilter).
// Without resources, rules stay fail-closed (nil hooks).

var (
	agreementStackOnce sync.Once
	wiredAgreement     *AgreementRule
	wiredSubjectVerb   *SubjectVerbAgreementRule
	wiredVerbAgreement *VerbAgreementRule
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
	wiredVerbAgreement = NewVerbAgreementRule(nil)

	if tagger != nil {
		// Java: GermanTagger.lookup(lower).hasPosTagStartingWith("VER:INF")
		wiredSubjectVerb.LookupInfinitive = func(lowerWord string) bool {
			return lookupIsInfinitive(tagger, lowerWord)
		}
		// Partial lt.check stand-in for compound phrases (Java: only DE_AGREEMENT +
		// GERMAN_SPELLER active). We require last word tags as SUB and, when a
		// filter dict is wired, that last word is not misspelled.
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
// (sentence-start uppercase verb workaround). Pass existing rule or nil.
func WireMissingVerbRule(r *MissingVerbRule) *MissingVerbRule {
	if r == nil {
		r = NewMissingVerbRule(nil)
	}
	root := DiscoverGermanResourceDir()
	tagger := openDiscoveredGermanTagger(root)
	if tagger != nil {
		r.TagFirstLowercased = func(lower string) bool {
			rd := tagger.Lookup(lower)
			return rd != nil && rd.HasPosTagStartingWith("VER")
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
	// GERMAN_SPELLER_RULE stand-in: reject when wired filter dict marks last word misspelled.
	last := parts[len(parts)-1]
	if FilterDictAvailable() && FilterDictIsMisspelled(last) {
		// Hyphen compounds: accept if each segment is known (Java speller on full form
		// may differ; still fail-closed rather than invent).
		if !strings.Contains(last, "-") {
			return false
		}
		for _, seg := range strings.Split(last, "-") {
			if seg == "" {
				continue
			}
			if FilterDictIsMisspelled(seg) {
				return false
			}
		}
	}
	// Last content word must be a noun (closed compound or hyphen segment) — structural
	// prerequisite for open-compound rewrite; without SUB reading, fail-closed.
	if !compoundLastWordIsNoun(tagger, phrase) {
		return false
	}
	// DE_AGREEMENT stand-in: tag phrase and require no AgreementRule match.
	// (Java: lt.check(phrase).isEmpty() with DE_AGREEMENT still active.)
	if !compoundPhraseAgreementOK(tagger, parts) {
		return false
	}
	return true
}

// compoundPhraseAgreementOK tags tokens with the German tagger and runs AgreementRule
// without CompoundPhraseValid (no recursive open-compound invent).
func compoundPhraseAgreementOK(tagger *detag.GermanTagger, parts []string) bool {
	if tagger == nil || len(parts) == 0 {
		return false
	}
	tagged := tagger.Tag(parts)
	if len(tagged) == 0 {
		return false
	}
	ss := languagetool.SentenceStartTagName
	toks := make([]*languagetool.AnalyzedTokenReadings, 0, len(tagged)+1)
	toks = append(toks, languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("", &ss, nil), 0,
	))
	pos := 0
	for i, tr := range tagged {
		if tr == nil {
			// untagged token — AgreementRule fails closed on missing POS (no invent)
			w := parts[i]
			if i < len(parts) {
				tr = languagetool.NewAnalyzedTokenReadingsAt(
					languagetool.NewAnalyzedToken(w, nil, nil), pos,
				)
			}
		}
		if tr != nil {
			tr.SetStartPos(pos)
			toks = append(toks, tr)
		}
		if i < len(parts) {
			pos += tokenizers.UTF16Len(parts[i]) + 1
		}
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	// Fresh rule: no CompoundPhraseValid → open-compound branch stays fail-closed.
	agr := NewAgreementRule(nil)
	return len(agr.Match(sent)) == 0
}

// compoundLastWordIsNoun is a structural proxy: last word (or hyphen segment) tags as SUB.
func compoundLastWordIsNoun(tagger *detag.GermanTagger, phrase string) bool {
	if tagger == nil || phrase == "" {
		return false
	}
	parts := splitASCIISpaceOmitEmptyDE(phrase)
	if len(parts) == 0 {
		return false
	}
	last := parts[len(parts)-1]
	// closed form "Originalmail" — lookup as-is first
	if rd := tagger.Lookup(last); rd != nil && rd.HasPosTagStartingWith("SUB") {
		return true
	}
	// hyphen form: Original-Mail → check last segment
	seg := last
	if i := strings.LastIndex(last, "-"); i >= 0 && i+1 < len(last) {
		seg = last[i+1:]
	}
	rd := tagger.Lookup(seg)
	if rd == nil {
		// try lowercase first for closed compounds built with LowercaseFirstChar
		rd = tagger.Lookup(strings.ToLower(seg))
	}
	if rd == nil {
		// try full closed form lowercased first char (Java potentialCompound)
		if len(last) > 0 {
			rd = tagger.Lookup(last)
			if rd == nil {
				// Lowercase first rune of last for weird casing
				r := []rune(last)
				if len(r) > 0 {
					r[0] = []rune(strings.ToLower(string(r[0])))[0]
					rd = tagger.Lookup(string(r))
				}
			}
		}
	}
	if rd == nil {
		return false
	}
	return rd.HasPosTagStartingWith("SUB")
}

// Ensure unused import of languagetool for AnalyzePlain callers if needed later.
var _ = languagetool.SentenceStartTagName
