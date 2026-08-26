package nl

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	nlXmlOnce sync.Once
	nlXmlInst *disambigrules.XmlRuleDisambiguator
)

// DutchXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(lang, true)
// as used by DutchHybridDisambiguator: official nl/disambiguation.xml then
// disambiguation-global.xml (useGlobalDisambiguation=true). Process-cached.
func DutchXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	nlXmlOnce.Do(func() {
		nlXmlInst = loadNLXmlRuleDisambiguator()
	})
	return nlXmlInst
}

// loadNLXmlRuleDisambiguator ports XmlRuleDisambiguator(Dutch, true):
// language XML first, then global, UnifierConfig from nl pack (or first non-nil).
func loadNLXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	var all []*disambigrules.DisambiguationPatternRule
	var uni *patterns.UnifierConfiguration
	loader := disambigrules.NewDisambiguationRuleLoader()

	if p := DiscoverDutchDisambiguationXML(); p != "" {
		if rules, u, err := loadNLDisambigFile(loader, p, "nl"); err == nil {
			all = append(all, rules...)
			if uni == nil {
				uni = u
			}
		}
	}
	// Java useGlobalDisambiguation=true: append disambiguation-global.xml
	if p := DiscoverGlobalDisambiguationXML(); p != "" {
		if rules, u, err := loadNLDisambigFile(loader, p, "global"); err == nil {
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

func loadNLDisambigFile(loader *disambigrules.DisambiguationRuleLoader, path, langCode string) ([]*disambigrules.DisambiguationPatternRule, *patterns.UnifierConfiguration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	return loader.GetRulesAndUnifierFromReader(f, langCode, path)
}

// DiscoverDutchDisambiguationXML finds official nl/disambiguation.xml
// (Java resource /nl/disambiguation.xml used by XmlRuleDisambiguator(Dutch, …)).
func DiscoverDutchDisambiguationXML() string {
	if p := os.Getenv("LANG_NL_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "nl",
		"src", "main", "resources", "org", "languagetool", "resource", "nl", "disambiguation.xml")
	return walkUpFindNL(rel)
}

// DiscoverGlobalDisambiguationXML finds official disambiguation-global.xml
// (Java resource /org/languagetool/resource/disambiguation-global.xml).
// Twin of de package discoverer for tagging/nl (DutchHybrid useGlobal=true).
func DiscoverGlobalDisambiguationXML() string {
	if p := os.Getenv("LANG_DISAMBIGUATION_GLOBAL"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
		"org", "languagetool", "resource", "disambiguation-global.xml")
	return walkUpFindNL(rel)
}

func walkUpFindNL(rel string) string {
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
