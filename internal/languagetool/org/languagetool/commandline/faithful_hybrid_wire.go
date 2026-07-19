package commandline

import (
	"os"
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/ca"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/de"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/es"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/nl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/fr"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/pt"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	ukdis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/uk"
	entag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
)

// RegisterEnglishHybridDisambiguator installs Java EnglishHybridDisambiguator on lt:
// spelling_global MultiWordChunker → /en/multiwords.txt chunker → XmlRuleDisambiguator(lang, true).
// Resources must be official LT files (inspiration / vendored upstream), not soft extracts.
//
// Java: org.languagetool.tagging.en.EnglishHybridDisambiguator
func RegisterEnglishHybridDisambiguator(lt *languagetool.JLanguageTool, opts *CommandLineOptions) bool {
	if lt == nil {
		return false
	}
	hybrid := entag.NewEnglishHybridDisambiguator()

	// Java: MultiWordChunker.getInstance("/spelling_global.txt", true, true, false, tagForNotAddingTags)
	if p := DiscoverSpellingGlobal(opts); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: true,
			AllowAllUppercase:     true,
			AllowTitlecase:        false,
			DefaultTag:            disambiguation.TagForNotAddingTags,
		}); err == nil && c != nil {
			c.AddIgnoreSpelling = true
			hybrid.GlobalChunker = c
		}
	}

	// Java: MultiWordChunker.getInstance("/en/multiwords.txt", true, true, false)
	if p := DiscoverEnglishMultiwords(opts); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: true,
			AllowAllUppercase:     true,
			AllowTitlecase:        false,
		}); err == nil && c != nil {
			c.AddIgnoreSpelling = true
			c.SetRemovePreviousTags(true)
			hybrid.Chunker = c
		}
	}

	// Java: new XmlRuleDisambiguator(lang, true) after multiword chunkers
	// (language disambiguation.xml then disambiguation-global.xml).
	if xml := loadXmlRuleDisambiguator("en", opts, true); xml != nil && len(xml.Rules) > 0 {
		hybrid.RulesDisambiguator = xml
	}

	if hybrid.GlobalChunker == nil && hybrid.Chunker == nil && hybrid.RulesDisambiguator == nil {
		return false
	}
	lt.Disambiguator = hybrid
	return true
}

// RegisterHybridDisambiguator installs the Java hybrid for supported languages.
// Official multiwords + spelling_global + disambiguation.xml only (no soft extracts).
func RegisterHybridDisambiguator(lt *languagetool.JLanguageTool, lang string, opts *CommandLineOptions) bool {
	if lt == nil {
		return false
	}
	base := languageBaseCode(lang)
	switch base {
	case "en":
		return RegisterEnglishHybridDisambiguator(lt, opts)
	case "fr":
		return registerFrenchHybrid(lt, opts)
	case "es":
		return registerSpanishHybrid(lt, opts)
	case "pt":
		return registerPortugueseHybrid(lt, opts)
	case "de":
		return registerGermanHybrid(lt, opts)
	case "ca":
		return registerCatalanHybrid(lt, opts)
	case "nl":
		return registerDutchHybrid(lt, opts)
	case "uk":
		return registerUkrainianHybrid(lt, opts)
	default:
		return false
	}
}

// registerUkrainianHybrid ports UkrainianHybridDisambiguator wiring.
// Java: preDisambiguate (SimpleDisambiguator) → multiwords (/uk/multiwords.txt, allowFirstCap)
// → XmlRuleDisambiguator(Ukrainian) → hybrid context filters (in package uk).
func registerUkrainianHybrid(lt *languagetool.JLanguageTool, opts *CommandLineOptions) bool {
	h := ukdis.NewUkrainianHybridDisambiguator()
	// Simple maps are loaded by NewUkrainianHybridDisambiguator (official disambig_remove/dups).

	// Java: new UkrainianMultiwordChunker("/uk/multiwords.txt", true)
	if p := DiscoverLanguageMultiwords(opts, "uk"); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: true,
			AllowAllUppercase:     false,
			AllowTitlecase:        false,
		}); err == nil && c != nil {
			h.Chunker = c
		}
	}
	// Java: new XmlRuleDisambiguator(Ukrainian.DEFAULT_VARIANT) — language XML (global optional).
	if xml := loadXmlRuleDisambiguator("uk", opts, true); xml != nil && len(xml.Rules) > 0 {
		h.Inner = xml
	}
	// Always install when we have simple maps (always) or multiwords/XML.
	// Simple alone is useful; fail only if hybrid is completely empty of work.
	if h.Simple == nil && h.Chunker == nil && h.Inner == nil {
		return false
	}
	lt.Disambiguator = h
	return true
}

// registerFrenchHybrid ports FrenchHybridDisambiguator wiring.
// Java: multiwords true,true,false + removePreviousTags; global false,true,false tagForNotAddingTags.
func registerFrenchHybrid(lt *languagetool.JLanguageTool, opts *CommandLineOptions) bool {
	h := fr.NewFrenchHybridDisambiguator()
	// Java chunkerGlobal first
	if p := DiscoverSpellingGlobal(opts); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: false,
			AllowAllUppercase:     true,
			AllowTitlecase:        false,
			DefaultTag:            disambiguation.TagForNotAddingTags,
		}); err == nil && c != nil {
			c.AddIgnoreSpelling = true
			h.GlobalChunker = c
		}
	}
	if p := DiscoverLanguageMultiwords(opts, "fr"); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: true,
			AllowAllUppercase:     true,
			AllowTitlecase:        false,
		}); err == nil && c != nil {
			c.SetRemovePreviousTags(true)
			h.Chunker = c
		}
	}
	if xml := loadXmlRuleDisambiguator("fr", opts, true); xml != nil && len(xml.Rules) > 0 {
		h.Rules = xml
	}
	if h.GlobalChunker == nil && h.Chunker == nil && h.Rules == nil {
		return false
	}
	lt.Disambiguator = h
	return true
}

// registerSpanishHybrid ports SpanishHybridDisambiguator.
// Java global DefaultTag "NPCN000"; multiwords removePreviousTags.
func registerSpanishHybrid(lt *languagetool.JLanguageTool, opts *CommandLineOptions) bool {
	h := es.NewSpanishHybridDisambiguator()
	if p := DiscoverSpellingGlobal(opts); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: false,
			AllowAllUppercase:     true,
			AllowTitlecase:        false,
			DefaultTag:            "NPCN000",
		}); err == nil && c != nil {
			h.GlobalChunker = c
		}
	}
	if p := DiscoverLanguageMultiwords(opts, "es"); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: true,
			AllowAllUppercase:     true,
			AllowTitlecase:        false,
		}); err == nil && c != nil {
			c.SetRemovePreviousTags(true)
			h.Chunker = c
		}
	}
	if xml := loadXmlRuleDisambiguator("es", opts, true); xml != nil && len(xml.Rules) > 0 {
		h.Rules = xml
	}
	if h.GlobalChunker == nil && h.Chunker == nil && h.Rules == nil {
		return false
	}
	lt.Disambiguator = h
	return true
}

// registerPortugueseHybrid ports PortugueseHybridDisambiguator.
// Java multiwords true,true,true; global false,true,true "NPCN000"; ignoreSpelling on both.
func registerPortugueseHybrid(lt *languagetool.JLanguageTool, opts *CommandLineOptions) bool {
	h := pt.NewPortugueseHybridDisambiguator()
	if p := DiscoverSpellingGlobal(opts); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: false,
			AllowAllUppercase:     true,
			AllowTitlecase:        true,
			DefaultTag:            "NPCN000",
		}); err == nil && c != nil {
			c.AddIgnoreSpelling = true
			h.GlobalChunker = c
		}
	}
	if p := DiscoverLanguageMultiwords(opts, "pt"); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: true,
			AllowAllUppercase:     true,
			AllowTitlecase:        true,
		}); err == nil && c != nil {
			c.SetRemovePreviousTags(true)
			c.AddIgnoreSpelling = true
			h.Chunker = c
		}
	}
	if xml := loadXmlRuleDisambiguator("pt", opts, true); xml != nil && len(xml.Rules) > 0 {
		h.Rules = xml
	}
	if h.GlobalChunker == nil && h.Chunker == nil && h.Rules == nil {
		return false
	}
	lt.Disambiguator = h
	return true
}

// registerGermanHybrid ports GermanRuleDisambiguator.
// Java: multitoken-ignore → spelling_global → multitoken-suggest → XmlRuleDisambiguator(true).
func registerGermanHybrid(lt *languagetool.JLanguageTool, opts *CommandLineOptions) bool {
	h := de.NewGermanRuleDisambiguator()
	tagNone := disambiguation.MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
		AllowAllUppercase:     true,
		AllowTitlecase:        false,
		DefaultTag:            disambiguation.TagForNotAddingTags,
	}
	if p := DiscoverGermanMultitokenIgnore(opts); p != "" {
		if c, err := openMultiWordChunker(p, tagNone); err == nil && c != nil {
			c.AddIgnoreSpelling = true
			h.MultitokenIgnore = c
		}
	}
	if p := DiscoverSpellingGlobal(opts); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: false,
			AllowAllUppercase:     true,
			AllowTitlecase:        false,
			DefaultTag:            disambiguation.TagForNotAddingTags,
		}); err == nil && c != nil {
			c.AddIgnoreSpelling = true
			h.MultitokenGlobal = c
		}
	}
	if p := DiscoverGermanMultitokenSuggest(opts); p != "" {
		if c, err := openMultiWordChunker(p, tagNone); err == nil && c != nil {
			c.AddIgnoreSpelling = true
			h.MultitokenSuggest = c
		}
	}
	if xml := loadXmlRuleDisambiguator("de", opts, true); xml != nil && len(xml.Rules) > 0 {
		h.Rules = xml
	}
	if h.MultitokenIgnore == nil && h.MultitokenGlobal == nil && h.MultitokenSuggest == nil && h.Rules == nil {
		return false
	}
	lt.Disambiguator = h
	return true
}

// registerCatalanHybrid ports CatalanHybridDisambiguator.
// Java: global NPCN000 → multiwords removePrevious → XML → CatalanMultitokenDisambiguator.
func registerCatalanHybrid(lt *languagetool.JLanguageTool, opts *CommandLineOptions) bool {
	h := ca.NewCatalanHybridDisambiguator()
	if p := DiscoverSpellingGlobal(opts); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: false,
			AllowAllUppercase:     true,
			AllowTitlecase:        false,
			DefaultTag:            "NPCN000",
		}); err == nil && c != nil {
			h.GlobalChunker = c
		}
	}
	if p := DiscoverLanguageMultiwords(opts, "ca"); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: true,
			AllowAllUppercase:     true,
			AllowTitlecase:        false,
		}); err == nil && c != nil {
			c.SetRemovePreviousTags(true)
			h.Chunker = c
		}
	}
	if xml := loadXmlRuleDisambiguator("ca", opts, true); xml != nil && len(xml.Rules) > 0 {
		h.Rules = xml
	}
	// Java CatalanMultitokenDisambiguator after XML.
	// Without Morfologik multitoken speller, IsMisspelled stays nil (no invent list).
	h.Multitoken = ca.NewCatalanMultitokenDisambiguator()
	if h.GlobalChunker == nil && h.Chunker == nil && h.Rules == nil {
		return false
	}
	lt.Disambiguator = h
	return true
}

// registerDutchHybrid ports DutchHybridDisambiguator.
// Java: global + multiwords both tagForNotAddingTags, ignoreSpelling; then XML.
func registerDutchHybrid(lt *languagetool.JLanguageTool, opts *CommandLineOptions) bool {
	h := nl.NewDutchHybridDisambiguator()
	tagNone := disambiguation.MultiWordChunkerSettings{
		AllowFirstCapitalized: false,
		AllowAllUppercase:     true,
		AllowTitlecase:        false,
		DefaultTag:            disambiguation.TagForNotAddingTags,
	}
	if p := DiscoverSpellingGlobal(opts); p != "" {
		if c, err := openMultiWordChunker(p, tagNone); err == nil && c != nil {
			c.AddIgnoreSpelling = true
			h.GlobalChunker = c
		}
	}
	if p := DiscoverLanguageMultiwords(opts, "nl"); p != "" {
		if c, err := openMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			AllowFirstCapitalized: true,
			AllowAllUppercase:     true,
			AllowTitlecase:        false,
			DefaultTag:            disambiguation.TagForNotAddingTags,
		}); err == nil && c != nil {
			c.AddIgnoreSpelling = true
			h.Chunker = c
		}
	}
	if xml := loadXmlRuleDisambiguator("nl", opts, true); xml != nil && len(xml.Rules) > 0 {
		h.Rules = xml
	}
	if h.GlobalChunker == nil && h.Chunker == nil && h.Rules == nil {
		return false
	}
	lt.Disambiguator = h
	return true
}

func openMultiWordChunker(path string, settings disambiguation.MultiWordChunkerSettings) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return disambiguation.NewMultiWordChunkerFromReader(f, settings)
}

// loadXmlRuleDisambiguator loads official disambiguation.xml (+ optional global).
// Java XmlRuleDisambiguator(language, useGlobalDisambiguation).
func loadXmlRuleDisambiguator(lang string, opts *CommandLineOptions, useGlobal bool) *disambigrules.XmlRuleDisambiguator {
	base := languageBaseCode(lang)
	var all []*disambigrules.DisambiguationPatternRule
	var uni *patterns.UnifierConfiguration

	if p := DiscoverLanguageDisambiguationXML(opts, base); p != "" {
		rules, u, err := loadDisambigRulesFile(p, base)
		if err == nil {
			all = append(all, rules...)
			if uni == nil {
				uni = u
			}
		}
	}
	if useGlobal {
		if p := DiscoverGlobalDisambiguationXML(opts); p != "" {
			rules, u, err := loadDisambigRulesFile(p, "global")
			if err == nil {
				all = append(all, rules...)
				if uni == nil {
					uni = u
				}
			}
		}
	}
	if len(all) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(all)
	x.UnifierConfig = uni
	return x
}

func loadDisambigRulesFile(path, languageCode string) ([]*disambigrules.DisambiguationPatternRule, *patterns.UnifierConfiguration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	loader := disambigrules.NewDisambiguationRuleLoader()
	return loader.GetRulesAndUnifierFromReader(f, languageCode, path)
}

// DiscoverSpellingGlobal finds official spelling_global.txt.
// Java: /org/languagetool/resource/spelling_global.txt
func DiscoverSpellingGlobal(opts *CommandLineOptions) string {
	if p := os.Getenv("LANG_SPELLING_GLOBAL"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "spelling_global.txt"),
			filepath.Join(opts.GetDataDir(), "resource", "spelling_global.txt"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	for _, rel := range []string{
		filepath.Join("testdata", "upstream", "spelling_global.txt"),
		filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
			"org", "languagetool", "resource", "spelling_global.txt"),
	} {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// DiscoverLanguageDisambiguationXML finds official disambiguation.xml for lang.
// Not soft extracts. CLI --disambiguation-file / LANG_DISAMBIGUATION_FILE override.
func DiscoverLanguageDisambiguationXML(opts *CommandLineOptions, lang string) string {
	base := languageBaseCode(lang)
	if base == "" {
		return ""
	}
	if opts != nil {
		if p := opts.GetDisambiguationFile(); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), base, "disambiguation.xml"),
			filepath.Join(opts.GetDataDir(), "resource", base, "disambiguation.xml"),
			filepath.Join(opts.GetDataDir(), "upstream", base, "resource", "disambiguation.xml"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	for _, rel := range []string{
		filepath.Join("testdata", "upstream", base, "resource", "disambiguation.xml"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", base,
			"src", "main", "resources", "org", "languagetool", "resource", base, "disambiguation.xml"),
	} {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// DiscoverGlobalDisambiguationXML finds official disambiguation-global.xml.
// Java: org/languagetool/resource/disambiguation-global.xml
func DiscoverGlobalDisambiguationXML(opts *CommandLineOptions) string {
	if p := os.Getenv("LANG_DISAMBIGUATION_GLOBAL"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "disambiguation-global.xml"),
			filepath.Join(opts.GetDataDir(), "resource", "disambiguation-global.xml"),
			filepath.Join(opts.GetDataDir(), "upstream", "resource", "disambiguation-global.xml"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	for _, rel := range []string{
		filepath.Join("testdata", "upstream", "resource", "disambiguation-global.xml"),
		filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
			"org", "languagetool", "resource", "disambiguation-global.xml"),
	} {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}
