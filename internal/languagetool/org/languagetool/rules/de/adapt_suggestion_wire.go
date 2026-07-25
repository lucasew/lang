package de

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	detok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/de"
)

var (
	adaptWireOnce sync.Once
	adaptWired    *AdaptSuggestionFilter

	// agreementFilterOnce builds SuggestionFilter(AgreementRule, German) once
	// (Java constructs per acceptRuleMatch; caching is behavior-equivalent).
	agreementFilterOnce sync.Once
	agreementFilterFn   func(suggs []string, template string) []string
)

// WireAdaptSuggestionFilter returns an AdaptSuggestionFilter with GenderOf +
// Synthesize + FilterSuggestions hooked from discovered DE resources when present
// (Java GermanTagger.INSTANCE + GermanSynthesizer.INSTANCE +
// SuggestionFilter(AgreementRule, German)).
func WireAdaptSuggestionFilter() *AdaptSuggestionFilter {
	adaptWireOnce.Do(func() {
		adaptWired = NewAdaptSuggestionFilter()
		root := DiscoverGermanResourceDir()
		if tagger := openDiscoveredGermanTagger(root); tagger != nil {
			adaptWired.GenderOf = func(word string) string {
				return nounGenderFromTagger(tagger, word)
			}
		}
		if openDiscoveredGermanSynthesizer() != nil || openDiscoveredGermanSynthBase() != nil {
			adaptWired.Synthesize = synthesizeGermanRE
		}
		adaptWired.FilterSuggestions = defaultAgreementSuggestionFilter()
	})
	if adaptWired == nil {
		return NewAdaptSuggestionFilter()
	}
	// shallow copy so callers can still override hooks in tests
	cp := *adaptWired
	return &cp
}

// defaultAgreementSuggestionFilter ports:
//
//	AgreementRule agreementRule = new AgreementRule(..., german);
//	SuggestionFilter suggestionFilter = new SuggestionFilter(agreementRule, german);
//	suggestionFilter.filter(newSugg, template)
//
// Java uses lang.createDefaultJLanguageTool().analyzeText for each candidate.
func defaultAgreementSuggestionFilter() func(suggs []string, template string) []string {
	agreementFilterOnce.Do(func() {
		lt := languagetool.NewJLanguageTool("de-DE")
		if !WireGermanTagWord(lt) {
			// fail-closed: no tags → AgreementRule cannot reject; keep identity
			// only if tagger missing (no invent filter)
			agreementFilterFn = nil
			return
		}
		WireGermanDisambiguator(lt)
		agr := NewAgreementRule(nil)
		sf := rules.NewSuggestionFilter(func(filled string) bool {
			sent := lt.GetAnalyzedSentence(filled)
			if sent == nil {
				return false
			}
			return len(agr.Match(sent)) > 0
		})
		agreementFilterFn = sf.Filter
	})
	return agreementFilterFn
}

// nounGenderFromGermanTagger ports AdaptSuggestionFilter.getNounGender via
// GermanTagger.INSTANCE (DiscoveredGermanTagger).
func nounGenderFromGermanTagger(word string) string {
	return nounGenderFromTagger(DiscoveredGermanTagger(), word)
}

// nounGenderFromTagger ports AdaptSuggestionFilter.getNounGender:
//
//	List<AnalyzedTokenReadings> readings = GermanTagger.INSTANCE.tag(singletonList(word));
//	first reading POS starts with SUB: → parts[3] (MAS|FEM|NEU)
func nounGenderFromTagger(tagger *detag.GermanTagger, word string) string {
	if tagger == nil || word == "" {
		return ""
	}
	// Java: tag(Collections.singletonList(word)) → ignoreCase=true
	readings := tagger.Tag([]string{word})
	for _, atr := range readings {
		if atr == nil {
			continue
		}
		rds := atr.GetReadings()
		if len(rds) == 0 {
			continue
		}
		// Java: only first reading
		r0 := rds[0]
		if r0 == nil || r0.GetPOSTag() == nil {
			continue
		}
		pos := *r0.GetPOSTag()
		if !strings.HasPrefix(pos, "SUB:") {
			continue
		}
		parts := strings.Split(pos, ":")
		if len(parts) >= 4 {
			return parts[3]
		}
	}
	return ""
}

// openDiscoveredGermanTagger builds a GermanTagger from resource dir (Java
// GermanTagger.INSTANCE): morfologik + added/removed manuals + ExpansionInfos
// from spelling.txt + GermanCompoundTokenizer.getStrictInstance for unknown compounds.
func openDiscoveredGermanTagger(resourceRoot string) *detag.GermanTagger {
	var morfo tagging.WordTagger
	if p := DiscoverGermanPOSDict(); p != "" {
		if mt := tagging.OpenMorfologikTagger(p); mt != nil {
			morfo = mt
		}
	} else if resourceRoot != "" {
		if p := filepath.Join(resourceRoot, "german.dict"); fileExists(p) {
			if mt := tagging.OpenMorfologikTagger(p); mt != nil {
				morfo = mt
			}
		}
	}
	loadManual := func(name string) tagging.WordTagger {
		if resourceRoot == "" {
			return nil
		}
		p := filepath.Join(resourceRoot, name)
		if !fileExists(p) {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()
		mt, err := tagging.NewManualTagger(f)
		if err != nil {
			return nil
		}
		return mt
	}
	var manuals tagging.WordTagger
	added := loadManual("added.txt")
	addedCustom := loadManual("added_custom.txt")
	switch {
	case added != nil && addedCustom != nil:
		manuals = tagging.NewCombiningTagger(added, addedCustom, false)
	case added != nil:
		manuals = added
	case addedCustom != nil:
		manuals = addedCustom
	}
	var adjEx *detag.SpellingAdjExpansion
	var verbEx *detag.SpellingVerbExpansion
	if resourceRoot != "" {
		sp := filepath.Join(resourceRoot, "hunspell", "spelling.txt")
		if fileExists(sp) {
			if ex, err := detag.LoadSpellingAdjExpansionFromFile(sp); err == nil && ex != nil {
				adjEx = ex
				if manuals != nil {
					manuals = tagging.NewCombiningTagger(manuals, ex, false)
				} else {
					manuals = ex
				}
			}
			if vx, err := detag.LoadSpellingVerbExpansionFromFile(sp); err == nil && vx != nil {
				verbEx = vx
			}
		}
	}
	tagger1 := tagging.WordTagger(tagging.MapWordTagger{})
	if morfo != nil {
		tagger1 = morfo
	}
	tagger2 := tagging.WordTagger(tagging.MapWordTagger{})
	if manuals != nil {
		tagger2 = manuals
	}
	if morfo == nil && manuals == nil {
		return nil
	}
	removal := loadManual("removed.txt")
	var wt tagging.WordTagger
	if removal != nil {
		wt = tagging.NewCombiningTaggerWithRemoval(tagger1, tagger2, removal, false)
	} else {
		wt = tagging.NewCombiningTagger(tagger1, tagger2, false)
	}
	gt := detag.NewGermanTagger(wt)
	if adjEx != nil {
		gt.SetSpellingAdjExpansion(adjEx)
	}
	if verbEx != nil {
		gt.SetSpellingVerbExpansion(verbEx)
	}
	gt.SplitCompound = strictCompoundSplitFromResourceDir(resourceRoot)
	return gt
}

// strictCompoundSplitFromResourceDir ports GermanCompoundTokenizer.getStrictInstance
// used by GermanTagger (Java: EmbeddedGermanDictionary + extras only — not hunspell .dic).
func strictCompoundSplitFromResourceDir(resourceRoot string) func(word string) []string {
	_ = resourceRoot
	return func(word string) []string {
		if word == "" {
			return nil
		}
		return detok.GetStrictInstance().Tokenize(word)
	}
}
