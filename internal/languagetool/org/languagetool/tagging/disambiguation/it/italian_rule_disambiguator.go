package it

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// ItalianRuleDisambiguator ports
// org.languagetool.tagging.disambiguation.rules.it.ItalianRuleDisambiguator.
// Java: private final Disambiguator disambiguator = new XmlRuleDisambiguator(new Italian());
// (useGlobalDisambiguation default false).
type ItalianRuleDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Rules is the XmlRuleDisambiguator stage (Java final field).
	// Optional override for tests; NewItalianRuleDisambiguator loads official XML.
	Rules disambiguation.Disambiguator
}

var (
	itXmlOnce sync.Once
	itXmlInst *disambigrules.XmlRuleDisambiguator
)

// NewItalianRuleDisambiguator builds the Java default: XmlRuleDisambiguator(Italian)
// over official resource/it/disambiguation.xml (useGlobalDisambiguation=false).
func NewItalianRuleDisambiguator() *ItalianRuleDisambiguator {
	d := &ItalianRuleDisambiguator{}
	if xml := ItalianXmlRuleDisambiguator(); xml != nil {
		d.Rules = xml
	}
	return d
}

// ItalianXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Italian())
// (useGlobalDisambiguation default false). Process-cached like other language loaders.
func ItalianXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	itXmlOnce.Do(func() {
		itXmlInst = loadITXmlRuleDisambiguator()
	})
	return itXmlInst
}

func loadITXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverItalianDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "it", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverItalianDisambiguationXML finds official it/disambiguation.xml
// (Java resource /it/disambiguation.xml).
func DiscoverItalianDisambiguationXML() string {
	if p := os.Getenv("LANG_IT_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "it",
		"src", "main", "resources", "org", "languagetool", "resource", "it", "disambiguation.xml")
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

// Disambiguate ports ItalianRuleDisambiguator.disambiguate → XmlRuleDisambiguator.
func (d *ItalianRuleDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	if d != nil && d.Rules != nil {
		return d.Rules.Disambiguate(input)
	}
	return input
}

var _ disambiguation.Disambiguator = (*ItalianRuleDisambiguator)(nil)
