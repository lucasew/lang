package en

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterSoftEnglishDisambiguator_Multiword(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	// tagger optional
	if p := findEnglishPOSDict(t); p != "" {
		require.True(t, RegisterBinaryEnglishTagger(lt, p))
	}
	RegisterSoftEnglishDisambiguator(lt, "", "", "")
	require.NotNil(t, lt.Disambiguator)
	sents := lt.Analyze("I live in New York.")
	require.NotEmpty(t, sents)
	// find New or York readings with NNP multiword tag
	var sawNNP bool
	for _, s := range sents {
		for _, tok := range s.GetTokensWithoutWhitespace() {
			if tok == nil {
				continue
			}
			w := tok.GetToken()
			if w != "New" && w != "York" {
				continue
			}
			for i := 0; i < tok.GetReadingsLength(); i++ {
				at := tok.GetAnalyzedToken(i)
				if at == nil || at.GetPOSTag() == nil {
					continue
				}
				if *at.GetPOSTag() == "NNP" || *at.GetPOSTag() == "B-NNP" || *at.GetPOSTag() == "E-NNP" {
					sawNNP = true
				}
			}
		}
	}
	// MultiWordChunker may tag with B-/E- or full NNP depending on path
	require.True(t, sawNNP || multiwordTagged(sents), "expected multiword tags on New York, sents=%s", dumpSents(sents))
}

func multiwordTagged(sents []*languagetool.AnalyzedSentence) bool {
	for _, s := range sents {
		for _, tok := range s.GetTokens() {
			if tok == nil {
				continue
			}
			for i := 0; i < tok.GetReadingsLength(); i++ {
				at := tok.GetAnalyzedToken(i)
				if at == nil || at.GetPOSTag() == nil {
					continue
				}
				p := *at.GetPOSTag()
				if p == "NNP" || len(p) > 2 && (p[0] == 'B' || p[0] == 'E') && (containsNNP(p)) {
					return true
				}
			}
		}
	}
	return false
}

func containsNNP(p string) bool {
	return len(p) >= 3 && (p == "NNP" || p[len(p)-3:] == "NNP" || p == "B-NP" || p == "E-NP" || p == "B-NNP" || p == "E-NNP")
}

func dumpSents(sents []*languagetool.AnalyzedSentence) string {
	var b string
	for _, s := range sents {
		for _, tok := range s.GetTokensWithoutWhitespace() {
			if tok == nil {
				continue
			}
			b += tok.GetToken() + ":"
			for i := 0; i < tok.GetReadingsLength(); i++ {
				at := tok.GetAnalyzedToken(i)
				if at != nil && at.GetPOSTag() != nil {
					b += *at.GetPOSTag() + ","
				}
			}
			b += " "
		}
	}
	return b
}

func TestRegisterSoftEnglishDisambiguator_SoftXMLIgnoreSpelling(t *testing.T) {
	// locate testdata/disambiguation/en-soft.xml
	wd, _ := os.Getwd()
	path := ""
	dir := wd
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "testdata", "disambiguation", "en-soft.xml")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			path = cand
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if path == "" {
		t.Skip("en-soft.xml not found")
	}
	lt := languagetool.NewJLanguageTool("en")
	// mark everything unknown for spelling
	lt.AddRuleChecker("MORFOLOGIK_RULE_EN_US", languagetool.SimplePredicateSpellerChecker(
		"MORFOLOGIK_RULE_EN_US",
		func(w string) bool { return w != "iPhone" && w != "GitHub" },
		nil, nil, nil,
	))
	// pass ignore list too when present
	ignPath := ""
	dir = filepath.Dir(path)
	if st, err := os.Stat(filepath.Join(dir, "en-ignore-spelling.txt")); err == nil && st.Mode().IsRegular() {
		ignPath = filepath.Join(dir, "en-ignore-spelling.txt")
	}
	RegisterSoftEnglishDisambiguator(lt, "", path, ignPath)
	// without soft ignore list/XML, iPhone would flag; with ignore-spelling it should not
	m := lt.Check("I use an iPhone.")
	for _, x := range m {
		require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", x.RuleID, "%+v", m)
	}
}

func TestSoftIgnoreSpellingList_TechNames(t *testing.T) {
	wd, _ := os.Getwd()
	path := ""
	dir := wd
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "testdata", "disambiguation", "en-ignore-spelling.txt")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			path = cand
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if path == "" {
		t.Skip("en-ignore-spelling.txt not found")
	}
	lt := languagetool.NewJLanguageTool("en")
	lt.AddRuleChecker("MORFOLOGIK_RULE_EN_US", languagetool.SimplePredicateSpellerChecker(
		"MORFOLOGIK_RULE_EN_US",
		func(w string) bool {
			// treat listed tech names as "unknown" to prove ignore list works
			switch w {
			case "Kubernetes", "TypeScript", "Docker":
				return false
			default:
				return true
			}
		},
		nil, nil, nil,
	))
	RegisterSoftEnglishDisambiguator(lt, "", "", path)
	for _, text := range []string{
		"We deploy with Kubernetes.",
		"I prefer TypeScript.",
		"Docker is useful.",
	} {
		m := lt.Check(text)
		for _, x := range m {
			require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", x.RuleID, "text=%q matches=%+v", text, m)
		}
	}
}

func TestSoftXML_FilterWillRun(t *testing.T) {
	// Soft FILTER keeps only VB reading on "run" after "will"
	p := findEnglishPOSDict(t)
	lt := languagetool.NewJLanguageTool("en")
	require.True(t, RegisterBinaryEnglishTagger(lt, p))
	// locate en-soft.xml
	wd, _ := os.Getwd()
	xmlPath := ""
	dir := wd
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "testdata", "disambiguation", "en-soft.xml")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			xmlPath = cand
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if xmlPath == "" {
		t.Skip("en-soft.xml not found")
	}
	RegisterSoftEnglishDisambiguator(lt, "", xmlPath, "")
	sents := lt.Analyze("I will run tomorrow.")
	require.NotEmpty(t, sents)
	// find "run" token: primary POS should be VB after filter
	var runPOS []string
	for _, s := range sents {
		for _, tok := range s.GetTokensWithoutWhitespace() {
			if tok == nil || tok.GetToken() != "run" {
				continue
			}
			for i := 0; i < tok.GetReadingsLength(); i++ {
				at := tok.GetAnalyzedToken(i)
				if at != nil && at.GetPOSTag() != nil {
					runPOS = append(runPOS, *at.GetPOSTag())
				}
			}
		}
	}
	require.NotEmpty(t, runPOS)
	// FILTER keeps VB; other readings (NN/VBN/…) should be gone when filter applies
	for _, p := range runPOS {
		require.Equal(t, "VB", p, "runPOS=%v", runPOS)
	}
}
