package sv

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	svXmlOnce sync.Once
	svXmlInst *disambigrules.XmlRuleDisambiguator
)

// SwedishXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Swedish())
// (useGlobalDisambiguation default false) over official resource/sv/disambiguation.xml.
// Process-cached like PolishXmlRuleDisambiguator / GalicianXmlRuleDisambiguator.
func SwedishXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	svXmlOnce.Do(func() {
		svXmlInst = loadSVXmlRuleDisambiguator()
	})
	return svXmlInst
}

func loadSVXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverSwedishDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "sv", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverSwedishDisambiguationXML finds official sv/disambiguation.xml
// (Java resource /sv/disambiguation.xml).
func DiscoverSwedishDisambiguationXML() string {
	if p := os.Getenv("LANG_SV_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "sv",
		"src", "main", "resources", "org", "languagetool", "resource", "sv", "disambiguation.xml")
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
