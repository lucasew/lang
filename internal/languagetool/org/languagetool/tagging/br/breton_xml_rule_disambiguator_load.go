package br

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	brXmlOnce sync.Once
	brXmlInst *disambigrules.XmlRuleDisambiguator
)

// BretonXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Breton())
// (useGlobalDisambiguation default false) over official resource/br/disambiguation.xml.
// Process-cached like ArabicXmlRuleDisambiguator / PolishXmlRuleDisambiguator.
func BretonXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	brXmlOnce.Do(func() {
		brXmlInst = loadBRXmlRuleDisambiguator()
	})
	return brXmlInst
}

func loadBRXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverBretonDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "br", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverBretonDisambiguationXML finds official br/disambiguation.xml
// (Java resource /br/disambiguation.xml).
func DiscoverBretonDisambiguationXML() string {
	if p := os.Getenv("LANG_BR_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "br",
		"src", "main", "resources", "org", "languagetool", "resource", "br", "disambiguation.xml")
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
