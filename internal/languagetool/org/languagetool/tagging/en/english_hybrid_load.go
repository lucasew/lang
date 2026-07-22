package en

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// official EN multiword / disambiguation resources (Java EnglishHybridDisambiguator).
const (
	enMultiwordsRel     = "en/multiwords.txt"
	enSpellingGlobalRel = "spelling_global.txt"
)

var (
	englishHybridOnce sync.Once
	englishHybridInst *EnglishHybridDisambiguator
)

// DefaultEnglishHybridDisambiguator returns a process singleton matching Java
// EnglishHybridDisambiguator: spelling_global + multiwords chunkers (ignore-spelling)
// + XmlRuleDisambiguator(en disambiguation.xml + disambiguation-global.xml).
func DefaultEnglishHybridDisambiguator() *EnglishHybridDisambiguator {
	englishHybridOnce.Do(func() {
		englishHybridInst = loadEnglishHybridDisambiguator()
	})
	return englishHybridInst
}

func loadEnglishHybridDisambiguator() *EnglishHybridDisambiguator {
	d := NewEnglishHybridDisambiguator()
	// Java: chunkerGlobal first (spelling_global.txt, tagForNotAddingTags)
	if p := spelling.DiscoverSpellingResource(enSpellingGlobalRel); p != "" {
		if c, err := openENMultiWordChunker(p, disambiguation.TagForNotAddingTags); err == nil && c != nil {
			c.SetIgnoreSpelling(true)
			d.GlobalChunker = c
		}
	}
	// Java: MultiWordChunker.getInstance("/en/multiwords.txt", true, true, false)
	if p := spelling.DiscoverSpellingResource(enMultiwordsRel); p != "" {
		if c, err := openENMultiWordChunker(p, ""); err == nil && c != nil {
			c.SetIgnoreSpelling(true)
			c.SetRemovePreviousTags(true)
			d.Chunker = c
		}
	}
	// Java: new XmlRuleDisambiguator(lang, true) — language + global XML
	if xml := loadENXmlRuleDisambiguator(true); xml != nil && len(xml.Rules) > 0 {
		d.RulesDisambiguator = xml
	}
	return d
}

// openENMultiWordChunker loads MultiWordChunker from a multiwords-style file.
// defaultTag empty → phrase\ttag lines; non-empty → phrase-only with fixed tag (global).
func openENMultiWordChunker(path, defaultTag string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	settings := disambiguation.MultiWordChunkerSettings{
		DefaultTag:            defaultTag,
		AllowFirstCapitalized: true,
		AllowAllUppercase:     true,
		// Java getInstance(..., true, true, false) → allowTitlecase false
	}
	c, err := disambiguation.NewMultiWordChunkerFromReader(f, settings)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// EnglishXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(English)
// (useGlobalDisambiguation=false). Used by EnglishDisambiguationRuleTest.setUp.
func EnglishXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	return loadENXmlRuleDisambiguator(false)
}

// loadENXmlRuleDisambiguator ports XmlRuleDisambiguator(English, useGlobalDisambiguation).
func loadENXmlRuleDisambiguator(useGlobal bool) *disambigrules.XmlRuleDisambiguator {
	var all []*disambigrules.DisambiguationPatternRule
	var uni *patterns.UnifierConfiguration
	loader := disambigrules.NewDisambiguationRuleLoader()

	if p := discoverENDisambiguationXML(); p != "" {
		if rules, u, err := loadDisambigFile(loader, p, "en"); err == nil {
			all = append(all, rules...)
			if uni == nil {
				uni = u
			}
		}
	}
	if useGlobal {
		if p := discoverGlobalDisambiguationXML(); p != "" {
			if rules, u, err := loadDisambigFile(loader, p, "global"); err == nil {
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

func loadDisambigFile(loader *disambigrules.DisambiguationRuleLoader, path, langCode string) ([]*disambigrules.DisambiguationPatternRule, *patterns.UnifierConfiguration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	return loader.GetRulesAndUnifierFromReader(f, langCode, path)
}

func discoverENDisambiguationXML() string {
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	return walkUpFind(filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
		"src", "main", "resources", "org", "languagetool", "resource", "en", "disambiguation.xml"))
}

func discoverGlobalDisambiguationXML() string {
	if p := os.Getenv("LANG_DISAMBIGUATION_GLOBAL"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	return walkUpFind(filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
		"org", "languagetool", "resource", "disambiguation-global.xml"))
}

func walkUpFind(rel string) string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
