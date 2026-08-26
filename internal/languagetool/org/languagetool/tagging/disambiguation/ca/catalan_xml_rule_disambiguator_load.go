package ca

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	caXmlOnce sync.Once
	caXmlInst *disambigrules.XmlRuleDisambiguator
)

// CatalanXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(lang, true)
// as used by CatalanHybridDisambiguator: official ca/disambiguation.xml then
// disambiguation-global.xml (useGlobalDisambiguation=true). Process-cached.
func CatalanXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	caXmlOnce.Do(func() {
		caXmlInst = loadCAXmlRuleDisambiguator()
	})
	return caXmlInst
}

// loadCAXmlRuleDisambiguator ports XmlRuleDisambiguator(Catalan, true):
// language XML first, then global, UnifierConfig from ca pack (or first non-nil).
func loadCAXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	var all []*disambigrules.DisambiguationPatternRule
	var uni *patterns.UnifierConfiguration
	loader := disambigrules.NewDisambiguationRuleLoader()

	if p := DiscoverCatalanDisambiguationXML(); p != "" {
		if rules, u, err := loadCADisambigFile(loader, p, "ca"); err == nil {
			all = append(all, rules...)
			if uni == nil {
				uni = u
			}
		}
	}
	// Java useGlobalDisambiguation=true: append disambiguation-global.xml
	if p := DiscoverGlobalDisambiguationXML(); p != "" {
		if rules, u, err := loadCADisambigFile(loader, p, "global"); err == nil {
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

func loadCADisambigFile(loader *disambigrules.DisambiguationRuleLoader, path, langCode string) ([]*disambigrules.DisambiguationPatternRule, *patterns.UnifierConfiguration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	return loader.GetRulesAndUnifierFromReader(f, langCode, path)
}

// DiscoverCatalanDisambiguationXML finds official ca/disambiguation.xml
// (Java resource /ca/disambiguation.xml used by XmlRuleDisambiguator(Catalan, …)).
func DiscoverCatalanDisambiguationXML() string {
	if p := os.Getenv("LANG_CA_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ca",
		"src", "main", "resources", "org", "languagetool", "resource", "ca", "disambiguation.xml")
	return walkUpFindCA(rel)
}

// DiscoverGlobalDisambiguationXML finds official disambiguation-global.xml
// (Java resource /org/languagetool/resource/disambiguation-global.xml).
// Twin of pt/es discoverers for tagging/disambiguation/ca (CatalanHybrid useGlobal=true).
func DiscoverGlobalDisambiguationXML() string {
	if p := os.Getenv("LANG_DISAMBIGUATION_GLOBAL"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
		"org", "languagetool", "resource", "disambiguation-global.xml")
	return walkUpFindCA(rel)
}

func walkUpFindCA(rel string) string {
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
