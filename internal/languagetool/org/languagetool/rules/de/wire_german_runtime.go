package de

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// WireGermanRuntimeResources ports Java Language module defaults used by grammar
// filters and pattern suggestions:
//   - getDefaultSpellingRule → WireGermanFilterSpeller (de_DE/AT/CH.dict)
//   - GermanMultitokenSpeller.INSTANCE → SetDefaultMultitokenSpeller
//   - getSynthesizer → RegisterLanguageSynthesizer("de", GermanSynthesizer)
// Missing resources stay fail-closed (no invent).
//
// Pass lt to also wire createDefaultDisambiguator (GermanRuleDisambiguator).
// lt may be nil when only process-wide filter/synth hooks are needed.
func WireGermanRuntimeResources(variant string) {
	WireGermanRuntimeResourcesFor(nil, variant)
}

// WireGermanRuntimeResourcesFor is WireGermanRuntimeResources plus optional lt.Disambiguator.
func WireGermanRuntimeResourcesFor(lt *languagetool.JLanguageTool, variant string) {
	// Spelling rule for filters (Java GermanyGerman.getDefaultSpellingRule).
	if !FilterDictAvailable() {
		if p := DiscoverGermanHunspellDict(variant); p != "" {
			_ = WireGermanFilterSpeller(p)
		}
	}

	// Multitoken speller for MultitokenSpellerFilter (official resource files only).
	sp := DiscoverAndLoadGermanMultitokenSpeller()
	if sp != nil && sp.MultitokenSpeller != nil {
		var isMiss func(string) bool
		if FilterDictAvailable() {
			isMiss = FilterDictIsMisspelled
		}
		patterns.SetDefaultMultitokenSpeller(sp.MultitokenSpeller, isMiss)
	}

	// Synthesizer for pattern <match postag="…"/> (Java GermanSynthesizer.INSTANCE).
	if gs := openDiscoveredGermanSynthesizer(); gs != nil {
		patterns.RegisterLanguageSynthesizer("de", gs)
	} else if base := openDiscoveredGermanSynthBase(); base != nil {
		patterns.RegisterLanguageSynthesizer("de", base)
	}

	if lt != nil {
		// createDefaultTagger → GermanTagger / SwissGermanTagger for CH
		_ = WireGermanTagWordFor(lt, variant)
		// createDefaultDisambiguator → GermanRuleDisambiguator
		_ = WireGermanDisambiguator(lt)
		// createDefaultPostDisambiguationChunker → GermanChunker
		WireGermanChunker(lt)
	}
}

// germanRemoteFiltersOnce ports Java RemoteRuleFilters LoadingCache (load once).
var germanRemoteFiltersOnce sync.Once

// WireGermanRemoteRuleFilters loads official de/remote-rule-filters.xml into
// GlobalRemoteRuleFilters when the file is present (Java RemoteRuleFilters.load).
// Fail-closed when missing — no invent. Loads at most once per process.
func WireGermanRemoteRuleFilters() {
	germanRemoteFiltersOnce.Do(func() {
		p := DiscoverGermanRemoteRuleFiltersXML()
		if p == "" {
			return
		}
		_, _ = patterns.LoadRemoteRuleFiltersFile(p, "de")
	})
}

// WireGermanUpstreamGrammar loads official grammar.xml / style.xml when
// UseUpstreamGrammar (default on; LANG_USE_UPSTREAM_GRAMMAR=0 to skip).
// Same gate as CLI core_rules_checker. Do not invent soft token packs instead.
// Variant from lt language code: DE/AT also load de-DE-AT/grammar.xml (Java
// GermanyGerman / AustrianGerman / NonSwissGerman); CH does not.
func WireGermanUpstreamGrammar(lt *languagetool.JLanguageTool) {
	if lt == nil || !languagetool.UseUpstreamGrammar() {
		return
	}
	lang := lt.GetLanguageCode()
	if lang == "" {
		lang = "de"
	}
	// Java German base: /de/grammar.xml (+ style when present).
	if p := DiscoverGermanGrammarXML(); p != "" {
		_, _ = patterns.RegisterGrammarFile(lt, p, lang)
	}
	if p := DiscoverGermanStyleXML(); p != "" {
		_, _ = patterns.RegisterGrammarFile(lt, p, lang)
	}
	// Variant pattern files (Java Language.getRuleFileNames + GermanyGerman/AT overrides):
	//   de-DE / de-AT: + /de/de-DE-AT/grammar.xml (NonSwiss; not Swiss)
	//   de-CH: + /de/de-CH/grammar.xml (shortCodeWithCountryAndVariant path)
	variant := germanVariant(lang)
	switch variant {
	case "CH":
		if p := DiscoverGermanCHGrammarXML(); p != "" {
			_, _ = patterns.RegisterGrammarFile(lt, p, lang)
		}
	default:
		if p := DiscoverGermanDEATGrammarXML(); p != "" {
			_, _ = patterns.RegisterGrammarFile(lt, p, lang)
		}
	}
}
