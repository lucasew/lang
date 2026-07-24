package pt

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	ptXmlOnce sync.Once
	ptXmlInst *disambigrules.XmlRuleDisambiguator
)

// PortugueseXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(lang, true)
// as used by PortugueseHybridDisambiguator: official pt/disambiguation.xml then
// disambiguation-global.xml (useGlobalDisambiguation=true). Process-cached.
func PortugueseXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	ptXmlOnce.Do(func() {
		ptXmlInst = loadPTXmlRuleDisambiguator()
	})
	return ptXmlInst
}

// loadPTXmlRuleDisambiguator ports XmlRuleDisambiguator(Portuguese, true):
// language XML first, then global, UnifierConfig from pt pack (or first non-nil).
func loadPTXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	var all []*disambigrules.DisambiguationPatternRule
	var uni *patterns.UnifierConfiguration
	loader := disambigrules.NewDisambiguationRuleLoader()

	if p := DiscoverPortugueseDisambiguationXML(); p != "" {
		if rules, u, err := loadPTDisambigFile(loader, p, "pt"); err == nil {
			all = append(all, rules...)
			if uni == nil {
				uni = u
			}
		}
	}
	// Java useGlobalDisambiguation=true: append disambiguation-global.xml
	if p := DiscoverGlobalDisambiguationXML(); p != "" {
		if rules, u, err := loadPTDisambigFile(loader, p, "global"); err == nil {
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

func loadPTDisambigFile(loader *disambigrules.DisambiguationRuleLoader, path, langCode string) ([]*disambigrules.DisambiguationPatternRule, *patterns.UnifierConfiguration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	return loader.GetRulesAndUnifierFromReader(f, langCode, path)
}

// DiscoverPortugueseDisambiguationXML finds official pt/disambiguation.xml
// (Java resource /pt/disambiguation.xml used by XmlRuleDisambiguator(Portuguese, …)).
func DiscoverPortugueseDisambiguationXML() string {
	if p := os.Getenv("LANG_PT_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "pt",
		"src", "main", "resources", "org", "languagetool", "resource", "pt", "disambiguation.xml")
	return walkUpFindPT(rel)
}

// DiscoverGlobalDisambiguationXML finds official disambiguation-global.xml
// (Java resource /org/languagetool/resource/disambiguation-global.xml).
// Twin of es/de/nl discoverers for tagging/disambiguation/pt (PortugueseHybrid useGlobal=true).
func DiscoverGlobalDisambiguationXML() string {
	if p := os.Getenv("LANG_DISAMBIGUATION_GLOBAL"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
		"org", "languagetool", "resource", "disambiguation-global.xml")
	return walkUpFindPT(rel)
}

func walkUpFindPT(rel string) string {
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
