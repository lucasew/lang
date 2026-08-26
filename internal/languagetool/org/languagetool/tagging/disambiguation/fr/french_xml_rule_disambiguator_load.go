package fr

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	frXmlOnce sync.Once
	frXmlInst *disambigrules.XmlRuleDisambiguator
)

// FrenchXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(lang, true)
// as used by FrenchHybridDisambiguator: official fr/disambiguation.xml then
// disambiguation-global.xml (useGlobalDisambiguation=true). Process-cached.
func FrenchXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	frXmlOnce.Do(func() {
		frXmlInst = loadFRXmlRuleDisambiguator()
	})
	return frXmlInst
}

// loadFRXmlRuleDisambiguator ports XmlRuleDisambiguator(French, true):
// language XML first, then global, UnifierConfig from fr pack (or first non-nil).
func loadFRXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	var all []*disambigrules.DisambiguationPatternRule
	var uni *patterns.UnifierConfiguration
	loader := disambigrules.NewDisambiguationRuleLoader()

	if p := DiscoverFrenchDisambiguationXML(); p != "" {
		if rules, u, err := loadFRDisambigFile(loader, p, "fr"); err == nil {
			all = append(all, rules...)
			if uni == nil {
				uni = u
			}
		}
	}
	// Java useGlobalDisambiguation=true: append disambiguation-global.xml
	if p := DiscoverGlobalDisambiguationXML(); p != "" {
		if rules, u, err := loadFRDisambigFile(loader, p, "global"); err == nil {
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

func loadFRDisambigFile(loader *disambigrules.DisambiguationRuleLoader, path, langCode string) ([]*disambigrules.DisambiguationPatternRule, *patterns.UnifierConfiguration, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	return loader.GetRulesAndUnifierFromReader(f, langCode, path)
}

// DiscoverFrenchDisambiguationXML finds official fr/disambiguation.xml
// (Java resource /fr/disambiguation.xml used by XmlRuleDisambiguator(French, …)).
func DiscoverFrenchDisambiguationXML() string {
	if p := os.Getenv("LANG_FR_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "fr",
		"src", "main", "resources", "org", "languagetool", "resource", "fr", "disambiguation.xml")
	return walkUpFindFR(rel)
}

// DiscoverGlobalDisambiguationXML finds official disambiguation-global.xml
// (Java resource /org/languagetool/resource/disambiguation-global.xml).
// Twin of ca/pt/nl discoverers for tagging/disambiguation/fr (FrenchHybrid useGlobal=true).
func DiscoverGlobalDisambiguationXML() string {
	if p := os.Getenv("LANG_DISAMBIGUATION_GLOBAL"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
		"org", "languagetool", "resource", "disambiguation-global.xml")
	return walkUpFindFR(rel)
}

func walkUpFindFR(rel string) string {
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
