package de

import (
	"os"
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigde "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/de"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// DiscoverGermanDisambiguationXML finds resource/de/disambiguation.xml.
func DiscoverGermanDisambiguationXML() string {
	const rel = "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/disambiguation.xml"
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, rel)
		if fileExists(cand) {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if root := DiscoverGermanResourceDir(); root != "" {
		p := filepath.Join(root, "disambiguation.xml")
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// DiscoverGlobalDisambiguationXML finds core resource/disambiguation-global.xml when present.
func DiscoverGlobalDisambiguationXML() string {
	candidates := []string{
		"inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/disambiguation-global.xml",
		"inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/disambiguation.xml",
		"testdata/upstream/disambiguation-global.xml",
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		for _, rel := range candidates {
			cand := filepath.Join(dir, rel)
			if fileExists(cand) {
				return cand
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// WireGermanDisambiguator installs GermanRuleDisambiguator on lt
// (Java GermanyGerman.createDefaultDisambiguator):
//
//	multitoken-ignore → spelling_global → multitoken-suggest → XmlRuleDisambiguator
//
// Official resource files only. Returns false if nothing could be wired (fail-closed).
func WireGermanDisambiguator(lt *languagetool.JLanguageTool) bool {
	if lt == nil {
		return false
	}
	h := disambigde.NewGermanRuleDisambiguator()
	tagNone := disambiguation.MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
		AllowAllUppercase:     true,
		AllowTitlecase:        false,
		DefaultTag:            disambiguation.TagForNotAddingTags,
	}

	root := DiscoverGermanResourceDir()
	if root != "" {
		if p := filepath.Join(root, "multitoken-ignore.txt"); fileExists(p) {
			if c := openMultiWordChunkerFile(p, tagNone); c != nil {
				c.AddIgnoreSpelling = true
				h.MultitokenIgnore = c
			}
		}
		if p := filepath.Join(root, "multitoken-suggest.txt"); fileExists(p) {
			if c := openMultiWordChunkerFile(p, tagNone); c != nil {
				c.AddIgnoreSpelling = true
				h.MultitokenSuggest = c
			}
		}
	}
	if p := DiscoverSpellingGlobal(); p != "" {
		if c := openMultiWordChunkerFile(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: false,
			AllowAllUppercase:     true,
			AllowTitlecase:        false,
			DefaultTag:            disambiguation.TagForNotAddingTags,
		}); c != nil {
			c.AddIgnoreSpelling = true
			h.MultitokenGlobal = c
		}
	}

	// XmlRuleDisambiguator(lang, true): language disambiguation.xml + optional global.
	loader := disambigrules.NewDisambiguationRuleLoader()
	var all []*disambigrules.DisambiguationPatternRule
	var uni *patterns.UnifierConfiguration
	if rules, u := loadDisambigXMLFile(loader, DiscoverGermanDisambiguationXML(), "de"); len(rules) > 0 {
		all = append(all, rules...)
		uni = u
	}
	if rules, u := loadDisambigXMLFile(loader, DiscoverGlobalDisambiguationXML(), "global"); len(rules) > 0 {
		all = append(all, rules...)
		if uni == nil {
			uni = u
		}
	}
	if len(all) > 0 {
		x := disambigrules.NewXmlRuleDisambiguator(all)
		x.UnifierConfig = uni
		h.Rules = x
	}

	if h.MultitokenIgnore == nil && h.MultitokenGlobal == nil && h.MultitokenSuggest == nil && h.Rules == nil {
		return false
	}
	lt.Disambiguator = h
	return true
}

func openMultiWordChunkerFile(path string, settings disambiguation.MultiWordChunkerSettings) *disambiguation.MultiWordChunker {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	c, err := disambiguation.NewMultiWordChunkerFromReader(f, settings)
	if err != nil {
		return nil
	}
	return c
}

func loadDisambigXMLFile(loader *disambigrules.DisambiguationRuleLoader, path, languageCode string) ([]*disambigrules.DisambiguationPatternRule, *patterns.UnifierConfiguration) {
	if path == "" || loader == nil {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, nil
	}
	defer f.Close()
	rules, uni, err := loader.GetRulesAndUnifierFromReader(f, languageCode, path)
	if err != nil {
		return nil, nil
	}
	return rules, uni
}
