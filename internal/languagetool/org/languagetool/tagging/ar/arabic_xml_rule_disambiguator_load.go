package ar

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	arXmlOnce sync.Once
	arXmlInst *disambigrules.XmlRuleDisambiguator
)

// ArabicXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Arabic())
// (useGlobalDisambiguation default false) over official resource/ar/disambiguation.xml.
// Process-cached like ItalianXmlRuleDisambiguator / English hybrid loaders.
func ArabicXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	arXmlOnce.Do(func() {
		arXmlInst = loadARXmlRuleDisambiguator()
	})
	return arXmlInst
}

func loadARXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverArabicDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "ar", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverArabicDisambiguationXML finds official ar/disambiguation.xml
// (Java resource /ar/disambiguation.xml).
func DiscoverArabicDisambiguationXML() string {
	if p := os.Getenv("LANG_AR_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ar",
		"src", "main", "resources", "org", "languagetool", "resource", "ar", "disambiguation.xml")
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
