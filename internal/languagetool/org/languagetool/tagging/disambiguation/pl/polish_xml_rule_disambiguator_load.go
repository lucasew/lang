package pl

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	plXmlOnce sync.Once
	plXmlInst *disambigrules.XmlRuleDisambiguator
)

// PolishXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Polish())
// (useGlobalDisambiguation default false) over official resource/pl/disambiguation.xml.
// Process-cached like RussianXmlRuleDisambiguator / ArabicXmlRuleDisambiguator.
func PolishXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	plXmlOnce.Do(func() {
		plXmlInst = loadPLXmlRuleDisambiguator()
	})
	return plXmlInst
}

func loadPLXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverPolishDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "pl", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverPolishDisambiguationXML finds official pl/disambiguation.xml
// (Java resource /pl/disambiguation.xml).
func DiscoverPolishDisambiguationXML() string {
	if p := os.Getenv("LANG_PL_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "pl",
		"src", "main", "resources", "org", "languagetool", "resource", "pl", "disambiguation.xml")
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
