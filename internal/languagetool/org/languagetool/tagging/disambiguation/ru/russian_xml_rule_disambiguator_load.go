package ru

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	// Register RU RuleFilters used by resource/ru/disambiguation.xml (VERB-KA etc.).
	_ "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ru"
)

var (
	ruXmlOnce sync.Once
	ruXmlInst *disambigrules.XmlRuleDisambiguator
)

// RussianXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(Russian.getInstance())
// (useGlobalDisambiguation default false) over official resource/ru/disambiguation.xml.
// Process-cached like ArabicXmlRuleDisambiguator / ItalianXmlRuleDisambiguator.
func RussianXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	ruXmlOnce.Do(func() {
		ruXmlInst = loadRUXmlRuleDisambiguator()
	})
	return ruXmlInst
}

func loadRUXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverRussianDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "ru", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverRussianDisambiguationXML finds official ru/disambiguation.xml
// (Java resource /ru/disambiguation.xml).
func DiscoverRussianDisambiguationXML() string {
	if p := os.Getenv("LANG_RU_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ru",
		"src", "main", "resources", "org", "languagetool", "resource", "ru", "disambiguation.xml")
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
