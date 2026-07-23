package uk

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	ukXmlOnce sync.Once
	ukXmlInst *disambigrules.XmlRuleDisambiguator
)

// UkrainianXmlRuleDisambiguator ports Java
// new XmlRuleDisambiguator(Ukrainian.DEFAULT_VARIANT)
// (useGlobalDisambiguation default false) over official resource/uk/disambiguation.xml.
// Process-cached like IrishXmlRuleDisambiguator / PolishXmlRuleDisambiguator.
// Does NOT append disambiguation-global.xml (Java useGlobalDisambiguation=false).
func UkrainianXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	ukXmlOnce.Do(func() {
		ukXmlInst = loadUKXmlRuleDisambiguator()
	})
	return ukXmlInst
}

func loadUKXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverUkrainianDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "uk", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverUkrainianDisambiguationXML finds official uk/disambiguation.xml
// (Java resource /uk/disambiguation.xml used by XmlRuleDisambiguator(Ukrainian.DEFAULT_VARIANT)).
func DiscoverUkrainianDisambiguationXML() string {
	if p := os.Getenv("LANG_UK_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "uk",
		"src", "main", "resources", "org", "languagetool", "resource", "uk", "disambiguation.xml")
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
