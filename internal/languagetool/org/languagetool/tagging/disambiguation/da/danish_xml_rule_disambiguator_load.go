package da

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	daXmlOnce sync.Once
	daXmlInst *disambigrules.XmlRuleDisambiguator
)

// DanishXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(this) from Danish
// (useGlobalDisambiguation default false) over official resource/da/disambiguation.xml.
// Process-cached like GreekXmlRuleDisambiguator / RomanianXmlRuleDisambiguator.
func DanishXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	daXmlOnce.Do(func() {
		daXmlInst = loadDAXmlRuleDisambiguator()
	})
	return daXmlInst
}

func loadDAXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverDanishDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "da", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverDanishDisambiguationXML finds official da/disambiguation.xml
// (Java resource /da/disambiguation.xml used by XmlRuleDisambiguator(Danish)).
func DiscoverDanishDisambiguationXML() string {
	if p := os.Getenv("LANG_DA_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "da",
		"src", "main", "resources", "org", "languagetool", "resource", "da", "disambiguation.xml")
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
