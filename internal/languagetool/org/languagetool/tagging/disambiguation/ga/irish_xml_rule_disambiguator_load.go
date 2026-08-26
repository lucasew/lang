package ga

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	gaXmlOnce sync.Once
	gaXmlInst *disambigrules.XmlRuleDisambiguator
)

// IrishXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(Irish.getInstance())
// (useGlobalDisambiguation default false) over official resource/ga/disambiguation.xml.
// Process-cached like GalicianXmlRuleDisambiguator / PolishXmlRuleDisambiguator.
func IrishXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	gaXmlOnce.Do(func() {
		gaXmlInst = loadGAXmlRuleDisambiguator()
	})
	return gaXmlInst
}

func loadGAXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverIrishDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "ga", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverIrishDisambiguationXML finds official ga/disambiguation.xml
// (Java resource /ga/disambiguation.xml used by XmlRuleDisambiguator(Irish)).
func DiscoverIrishDisambiguationXML() string {
	if p := os.Getenv("LANG_GA_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ga",
		"src", "main", "resources", "org", "languagetool", "resource", "ga", "disambiguation.xml")
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
