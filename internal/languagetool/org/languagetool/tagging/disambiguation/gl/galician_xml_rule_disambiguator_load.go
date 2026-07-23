package gl

import (
	"os"
	"path/filepath"
	"sync"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

var (
	glXmlOnce sync.Once
	glXmlInst *disambigrules.XmlRuleDisambiguator
)

// GalicianXmlRuleDisambiguator ports Java new XmlRuleDisambiguator(new Galician())
// (useGlobalDisambiguation default false) over official resource/gl/disambiguation.xml.
// Process-cached like PolishXmlRuleDisambiguator / RussianXmlRuleDisambiguator.
func GalicianXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	glXmlOnce.Do(func() {
		glXmlInst = loadGLXmlRuleDisambiguator()
	})
	return glXmlInst
}

func loadGLXmlRuleDisambiguator() *disambigrules.XmlRuleDisambiguator {
	p := DiscoverGalicianDisambiguationXML()
	if p == "" {
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	// Java XmlRuleDisambiguator(language) → useGlobalDisambiguation=false: language XML only.
	rules, uni, err := disambigrules.NewDisambiguationRuleLoader().GetRulesAndUnifierFromReader(f, "gl", p)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	x.UnifierConfig = uni
	return x
}

// DiscoverGalicianDisambiguationXML finds official gl/disambiguation.xml
// (Java resource /gl/disambiguation.xml).
func DiscoverGalicianDisambiguationXML() string {
	if p := os.Getenv("LANG_GL_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if p := os.Getenv("LANG_DISAMBIGUATION_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "gl",
		"src", "main", "resources", "org", "languagetool", "resource", "gl", "disambiguation.xml")
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
