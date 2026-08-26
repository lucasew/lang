package el

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	elXmlOnce sync.Once
	elXmlInst *disambigrules.XmlRuleDisambiguator
)

// GreekXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(this) from Greek
// (useGlobalDisambiguation default false) over official resource/el/disambiguation.xml.
// Process-cached like IrishXmlRuleDisambiguator / GalicianXmlRuleDisambiguator.
func GreekXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	elXmlOnce.Do(func() {
		elXmlInst = loadELXmlRuleDisambiguator()
	})
	return elXmlInst
}

func loadELXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverGreekDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "el", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverGreekDisambiguationXML finds official el/disambiguation.xml
// (Java resource /el/disambiguation.xml used by XmlRuleDisambiguator(Greek)).
func DiscoverGreekDisambiguationXML() string {
	if p := os.Getenv("LANG_EL_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "el",
		"src", "main", "resources", "org", "languagetool", "resource", "el", "disambiguation.xml")
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
