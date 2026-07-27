package ro

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	roXmlOnce sync.Once
	roXmlInst *disambigrules.XmlRuleDisambiguator
)

// RomanianXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(this) from Romanian
// (useGlobalDisambiguation default false) over official resource/ro/disambiguation.xml.
// Process-cached like GreekXmlRuleDisambiguator / IrishXmlRuleDisambiguator.
func RomanianXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	roXmlOnce.Do(func() {
		roXmlInst = loadROXmlRuleDisambiguator()
	})
	return roXmlInst
}

func loadROXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverRomanianDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "ro", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverRomanianDisambiguationXML finds official ro/disambiguation.xml
// (Java resource /ro/disambiguation.xml used by XmlRuleDisambiguator(Romanian)).
func DiscoverRomanianDisambiguationXML() string {
	if p := os.Getenv("LANG_RO_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ro",
		"src", "main", "resources", "org", "languagetool", "resource", "ro", "disambiguation.xml")
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
