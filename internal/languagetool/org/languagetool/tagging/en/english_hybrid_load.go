package en

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// official EN disambiguation resources (Java EnglishHybridDisambiguator).
// GlobalChunker: EnglishGlobalChunker / DiscoverEnglishGlobalChunker.
// Multiwords: EnglishMultiWordChunker / DiscoverEnglishMultiwords.

var (
	englishHybridOnce sync.Once
	englishHybridInst *EnglishHybridDisambiguator

	enXmlHybridOnce sync.Once
	enXmlHybridInst *disambigrules.XmlRuleDisambiguator

	enXmlLocalOnce sync.Once
	enXmlLocalInst *disambigrules.XmlRuleDisambiguator
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
	// + setIgnoreSpelling(true); no setRemovePreviousTags — process-cached twin.
	if g := EnglishGlobalChunker(); g != nil {
		d.GlobalChunker = g
	}
	// Java: MultiWordChunker.getInstance("/en/multiwords.txt", true, true, false)
	// + setIgnoreSpelling(true) + setRemovePreviousTags(true) — process-cached twin.
	if c := EnglishMultiWordChunker(); c != nil {
		d.Chunker = c
	}
	// Java: new XmlRuleDisambiguator(lang, true) — language + global XML
	if xml := EnglishHybridXmlRuleDisambiguator(); xml != nil && len(xml.Rules) > 0 {
		d.RulesDisambiguator = xml
	}
	return d
}

// EnglishHybridXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(lang, true)
// as used by EnglishHybridDisambiguator: official en/disambiguation.xml then
// disambiguation-global.xml (useGlobalDisambiguation=true). Process-cached.
func EnglishHybridXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	enXmlHybridOnce.Do(func() {
		enXmlHybridInst = loadENXmlRuleDisambiguator(true)
	})
	return enXmlHybridInst
}

// EnglishXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(English)
// (useGlobalDisambiguation=false). Used by EnglishDisambiguationRuleTest.setUp.
// Process-cached EN-only pack (no disambiguation-global.xml).
func EnglishXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	enXmlLocalOnce.Do(func() {
		enXmlLocalInst = loadENXmlRuleDisambiguator(false)
	})
	return enXmlLocalInst
}

// loadENXmlRuleDisambiguator ports XmlRuleDisambiguator(English, useGlobalDisambiguation).
// Language XML first; when useGlobal, append official disambiguation-global.xml.
func loadENXmlRuleDisambiguator(useGlobal bool) *disambigrules.XmlRuleDisambiguator {
	var all []*disambigrules.DisambiguationPatternRule
	var uni *patterns.UnifierConfiguration
	loader := disambigrules.NewDisambiguationRuleLoader()

	if p := DiscoverEnglishDisambiguationXML(); p != "" {
		if rules, u, err := loadDisambigFile(loader, p, "en"); err == nil {
			all = append(all, rules...)
			if uni == nil {
				uni = u
			}
		}
	}
	if useGlobal {
		if p := DiscoverGlobalDisambiguationXML(); p != "" {
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

// DiscoverEnglishDisambiguationXML finds official en/disambiguation.xml
// (Java resource /en/disambiguation.xml used by XmlRuleDisambiguator(English, …)).
func DiscoverEnglishDisambiguationXML() string {
	if p := os.Getenv("LANG_EN_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	return walkUpFind(filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
		"src", "main", "resources", "org", "languagetool", "resource", "en", "disambiguation.xml"))
}

// DiscoverGlobalDisambiguationXML finds official disambiguation-global.xml
// (Java resource /org/languagetool/resource/disambiguation-global.xml).
// Twin of nl/fr discoverers for EnglishHybrid useGlobal=true.
func DiscoverGlobalDisambiguationXML() string {
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
