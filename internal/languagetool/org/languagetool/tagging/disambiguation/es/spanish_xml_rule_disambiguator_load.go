package es

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	esXmlOnce sync.Once
	esXmlInst *disambigrules.XmlRuleDisambiguator
)

// SpanishXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(lang, true)
// as used by SpanishHybridDisambiguator: official es/disambiguation.xml then
// disambiguation-global.xml (useGlobalDisambiguation=true). Process-cached.
func SpanishXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	esXmlOnce.Do(func() {
		esXmlInst = loadESXmlRuleDisambiguator()
	})
	return esXmlInst
}

// loadESXmlRuleDisambiguator ports XmlRuleDisambiguator(Spanish, true):
// language XML first, then global, UnifierConfig from es pack (or first non-nil).
func loadESXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	var all []*disambigrules.DisambiguationPatternRule
	var uni *patterns.UnifierConfiguration
	loader := disambigrules.NewDisambiguationRuleLoader()

	if p := DiscoverSpanishDisambiguationXML(); p != "" {
		if rules, u, err := loadESDisambigFile(loader, p, "es"); err == nil {
			all = append(all, rules...)
			if uni == nil {
				uni = u
			}
		}
	}
	// Java useGlobalDisambiguation=true: append disambiguation-global.xml
	if p := DiscoverGlobalDisambiguationXML(); p != "" {
		if rules, u, err := loadESDisambigFile(loader, p, "global"); err == nil {
			all = append(all, rules...)
			if uni == nil {
				uni = u
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

func loadESDisambigFile(loader *disambigrules.DisambiguationRuleLoader, path, langCode string) ([]*disambigrules.DisambiguationPatternRule, *patterns.UnifierConfiguration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	return loader.GetRulesAndUnifierFromReader(f, langCode, path)
}

// DiscoverSpanishDisambiguationXML finds official es/disambiguation.xml
// (Java resource /es/disambiguation.xml used by XmlRuleDisambiguator(Spanish, …)).
func DiscoverSpanishDisambiguationXML() string {
	if p := os.Getenv("LANG_ES_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "es",
		"src", "main", "resources", "org", "languagetool", "resource", "es", "disambiguation.xml")
	return walkUpFindES(rel)
}

// DiscoverGlobalDisambiguationXML finds official disambiguation-global.xml
// (Java resource /org/languagetool/resource/disambiguation-global.xml).
// Twin of de/nl discoverers for tagging/disambiguation/es (SpanishHybrid useGlobal=true).
func DiscoverGlobalDisambiguationXML() string {
	if p := os.Getenv("LANG_DISAMBIGUATION_GLOBAL"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
		"org", "languagetool", "resource", "disambiguation-global.xml")
	return walkUpFindES(rel)
}

func walkUpFindES(rel string) string {
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
