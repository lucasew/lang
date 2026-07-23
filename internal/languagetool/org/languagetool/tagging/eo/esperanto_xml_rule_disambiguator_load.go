package eo

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	eoXmlOnce sync.Once
	eoXmlInst *disambigrules.XmlRuleDisambiguator
)

// EsperantoXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Esperanto())
// (useGlobalDisambiguation default false) over official resource/eo/disambiguation.xml.
// Process-cached like BretonXmlRuleDisambiguator / ArabicXmlRuleDisambiguator.
func EsperantoXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	eoXmlOnce.Do(func() {
		eoXmlInst = loadEOXmlRuleDisambiguator()
	})
	return eoXmlInst
}

func loadEOXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverEsperantoDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "eo", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverEsperantoDisambiguationXML finds official eo/disambiguation.xml
// (Java resource /eo/disambiguation.xml).
func DiscoverEsperantoDisambiguationXML() string {
	if p := os.Getenv("LANG_EO_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "eo",
		"src", "main", "resources", "org", "languagetool", "resource", "eo", "disambiguation.xml")
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
