package de

import (
	"os"
	"path/filepath"
	"regexp"
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
)

var masFemNeuFind = regexp.MustCompile(`MAS|FEM|NEU`)

// WireAdaptSuggestionFilter returns an AdaptSuggestionFilter with GenderOf +
// Synthesize hooked from discovered DE resources when present (Java
// GermanTagger.INSTANCE + GermanSynthesizer.INSTANCE). Without resources,
// hooks stay nil (fail-closed).
func WireAdaptSuggestionFilter() *AdaptSuggestionFilter {
	adaptWireOnce.Do(func() {
		adaptWired = NewAdaptSuggestionFilter()
		root := DiscoverGermanResourceDir()
		// GenderOf via tagger (manual + optional german.dict + spelling expansions)
		if tagger := openDiscoveredGermanTagger(root); tagger != nil {
			adaptWired.GenderOf = func(word string) string {
				return nounGenderFromTagger(tagger, word)
			}
		}
		// Synthesize via GermanSynthesizer when german_synth.dict present
		// (Java: GermanSynthesizer.INSTANCE — case/REMOVE/compound filters).
		if openDiscoveredGermanSynthesizer() != nil || openDiscoveredGermanSynthBase() != nil {
			adaptWired.Synthesize = synthesizeGermanRE
		}
		// SuggestionFilter via AgreementRule (Java AgreementRule + template)
		agr := NewAgreementRule(nil)
		sf := rules.NewSuggestionFilter(func(filled string) bool {
			sent := languagetool.AnalyzePlain(filled)
			return len(agr.Match(sent)) > 0
		})
		adaptWired.FilterSuggestions = sf.Filter
	})
	if adaptWired == nil {
		return NewAdaptSuggestionFilter()
	}
	// return a shallow copy so callers can still override hooks in tests
	cp := *adaptWired
	return &cp
}

// nounGenderFromTagger ports AdaptSuggestionFilter.getNounGender.
func nounGenderFromTagger(tagger *detag.GermanTagger, word string) string {
	if tagger == nil || word == "" {
		return ""
	}
	rd := tagger.Lookup(word)
	if rd == nil {
		return ""
	}
	for _, reading := range rd.GetReadings() {
		if reading == nil || reading.GetPOSTag() == nil {
			continue
		}
		pos := *reading.GetPOSTag()
		if !strings.HasPrefix(pos, "SUB:") {
			continue
		}
		if m := masFemNeuFind.FindString(pos); m != "" {
			return m
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
			// Java ExpansionInfos: /A /P adj forms + underscore verb lines from same file.
			if ex, err := detag.LoadSpellingAdjExpansionFromFile(sp); err == nil && ex != nil {
				adjEx = ex
				// Also available as WordTagger for exact surface hits (optional; Java uses adjInfos map only on unknown path).
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
	// Java: GermanCompoundTokenizer.getStrictInstance().tokenize for unknown compounds
	gt.SplitCompound = strictCompoundSplitFromResourceDir(resourceRoot)
	return gt
}

// strictCompoundSplitFromResourceDir ports GermanCompoundTokenizer.getStrictInstance
// used by GermanTagger (lexicon-backed when hunspell .dic present; else exceptions only).
func strictCompoundSplitFromResourceDir(resourceRoot string) func(word string) []string {
	tok := detok.NewGermanCompoundTokenizer(true)
	if resourceRoot != "" {
		for _, name := range []string{"de_DE.dic", "de_AT.dic", "de_CH.dic"} {
			p := filepath.Join(resourceRoot, "hunspell", name)
			f, err := os.Open(p)
			if err != nil {
				continue
			}
			_ = tok.LoadHunspellDic(f)
			_ = f.Close()
			break
		}
	}
	return func(word string) []string {
		if word == "" {
			return nil
		}
		return tok.Tokenize(word)
	}
}
