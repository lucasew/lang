package km

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	kmXmlOnce sync.Once
	kmXmlInst *disambigrules.XmlRuleDisambiguator
)

// KhmerXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Khmer())
// (useGlobalDisambiguation default false) over official resource/km/disambiguation.xml.
// Process-cached like EsperantoXmlRuleDisambiguator / BretonXmlRuleDisambiguator.
func KhmerXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	kmXmlOnce.Do(func() {
		kmXmlInst = loadKMXmlRuleDisambiguator()
	})
	return kmXmlInst
}

func loadKMXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverKhmerDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "km", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverKhmerDisambiguationXML finds official km/disambiguation.xml
// (Java resource /km/disambiguation.xml).
func DiscoverKhmerDisambiguationXML() string {
	if p := os.Getenv("LANG_KM_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "km",
		"src", "main", "resources", "org", "languagetool", "resource", "km", "disambiguation.xml")
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
