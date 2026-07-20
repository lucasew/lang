package uk

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	ukXmlOnce sync.Once
	ukXml     *disambigrules.XmlRuleDisambiguator
)

// LoadDefaultUkrainianXmlDisambiguator loads official /uk/disambiguation.xml
// (and optional global if present beside inspiration tree). Java:
// new XmlRuleDisambiguator(Ukrainian.DEFAULT_VARIANT).
func LoadDefaultUkrainianXmlDisambiguator() disambiguation.Disambiguator {
	ukXmlOnce.Do(func() {
		path := discoverUKDisambiguationXML()
		if path == "" {
			return
		}
		f, err := os.Open(path)
		if err != nil {
			return
		}
		defer f.Close()
		loader := disambigrules.NewDisambiguationRuleLoader()
		rules, uni, err := loader.GetRulesAndUnifierFromReader(f, "uk", path)
		if err != nil || len(rules) == 0 {
			return
		}
		// optional global disambiguation.xml
		if g := discoverGlobalDisambiguationXML(); g != "" {
			if gf, err := os.Open(g); err == nil {
				gr, _, err2 := loader.GetRulesAndUnifierFromReader(gf, "global", g)
				_ = gf.Close()
				if err2 == nil {
					rules = append(rules, gr...)
				}
			}
		}
		x := disambigrules.NewXmlRuleDisambiguator(rules)
		x.UnifierConfig = uni
		ukXml = x
	})
	if ukXml == nil {
		return nil
	}
	return ukXml
}

func discoverUKDisambiguationXML() string {
	return discoverResourcePath(
		"inspiration/languagetool/languagetool-language-modules/uk/src/main/resources/org/languagetool/resource/uk/disambiguation.xml",
	)
}

func discoverGlobalDisambiguationXML() string {
	// Java disambiguation-global.xml under resource/
	return discoverResourcePath(
		"inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/disambiguation-global.xml",
	)
}

func discoverResourcePath(rel string) string {
	_, file, _, ok := runtime.Caller(0)
	if ok {
		// this file is …/tagging/disambiguation/uk/ → walk up to module root
		// uk → disambiguation → tagging → languagetool → org → languagetool → internal → repo root
		root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../../../../"))
		p := filepath.Join(root, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, rel)
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
