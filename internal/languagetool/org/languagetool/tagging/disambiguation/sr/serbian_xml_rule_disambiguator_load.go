package sr

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	srXmlOnce sync.Once
	srXmlInst *disambigrules.XmlRuleDisambiguator
)

// SerbianXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Serbian())
// (useGlobalDisambiguation default false) over official resource/sr/disambiguation.xml.
// Process-cached like SwedishXmlRuleDisambiguator / PolishXmlRuleDisambiguator.
func SerbianXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	srXmlOnce.Do(func() {
		srXmlInst = loadSRXmlRuleDisambiguator()
	})
	return srXmlInst
}

func loadSRXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverSerbianDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "sr", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverSerbianDisambiguationXML finds official sr/disambiguation.xml
// (Java resource /sr/disambiguation.xml).
func DiscoverSerbianDisambiguationXML() string {
	if p := os.Getenv("LANG_SR_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "sr",
		"src", "main", "resources", "org", "languagetool", "resource", "sr", "disambiguation.xml")
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
