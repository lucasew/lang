package en

import (
	"os"
	"path/filepath"
	"strings"
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

func findSoftDisambigXML(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	dir := wd
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "testdata", "disambiguation", "en-soft.xml")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("en-soft.xml not found")
	return ""
}

func TestSoftXML_FilterWillRun(t *testing.T) {
	// Soft FILTER keeps only VB reading on "run" after "will"
	p := findEnglishPOSDict(t)
	lt := languagetool.NewJLanguageTool("en")
	require.True(t, RegisterBinaryEnglishTagger(lt, p))
	xmlPath := findSoftDisambigXML(t)
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

func TestSoftXML_FilterShouldRun(t *testing.T) {
	p := findEnglishPOSDict(t)
	lt := languagetool.NewJLanguageTool("en")
	require.True(t, RegisterBinaryEnglishTagger(lt, p))
	RegisterSoftEnglishDisambiguator(lt, "", findSoftDisambigXML(t), "")
	sents := lt.Analyze("You should run more.")
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
	for _, tag := range runPOS {
		require.Equal(t, "VB", tag, "runPOS=%v", runPOS)
	}
}

func TestSoftXML_ImmunizeBTW(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	lt.AddRuleChecker("MORFOLOGIK_RULE_EN_US", languagetool.SimplePredicateSpellerChecker(
		"MORFOLOGIK_RULE_EN_US",
		func(w string) bool { return w != "btw" && w != "irl" },
		nil, nil, nil,
	))
	RegisterSoftEnglishDisambiguator(lt, "", findSoftDisambigXML(t), "")
	for _, text := range []string{
		"Send that btw.",
		"We met irl yesterday.",
	} {
		m := lt.Check(text)
		for _, x := range m {
			require.NotEqual(t, "MORFOLOGIK_RULE_EN_US", x.RuleID, "text=%q matches=%+v", text, m)
		}
	}
}

func TestSplitMultiwordLine(t *testing.T) {
	phrase, tag, ok := splitMultiwordLine("New York\tNNP")
	require.True(t, ok)
	require.Equal(t, "New York", phrase)
	require.Equal(t, "NNP", tag)

	phrase, tag, ok = splitMultiwordLine("status quoNN:UN")
	require.True(t, ok)
	require.Equal(t, "status quo", phrase)
	require.Equal(t, "NN:UN", tag)

	_, _, ok = splitMultiwordLine("singletonNNP")
	require.False(t, ok)
}

func TestLoadUpstreamEnglishMultiwords(t *testing.T) {
	// Prefer vendored upstream multiwords when present.
	root := findRepoRoot(t)
	path := filepath.Join(root, "testdata", "disambiguation", "en-multiwords-upstream.txt")
	if _, err := os.Stat(path); err != nil {
		t.Skip("vendored multiwords missing")
	}
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	lines, err := loadTabSeparatedMultiwords(f)
	require.NoError(t, err)
	require.Greater(t, len(lines), 1000, "expected thousands of upstream multiwords")
	// sanity: known entry present
	found := false
	for _, l := range lines {
		if strings.HasPrefix(l, "New York\t") || strings.HasPrefix(l, "status quo\t") {
			found = true
			break
		}
	}
	require.True(t, found, "expected New York or status quo in loaded multiwords")
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	dir := wd
	for {
		if st, err := os.Stat(filepath.Join(dir, "testdata", "upstream")); err == nil && st.IsDir() {
			return dir
		}
		p := filepath.Dir(dir)
		if p == dir {
			t.Fatal("repo root not found")
		}
		dir = p
	}
}

func TestIgnoreSpellingPaths_MergesUpstreamSpelling(t *testing.T) {
	root := findRepoRoot(t)
	up := filepath.Join(root, "testdata", "disambiguation", "en-spelling-upstream.txt")
	if _, err := os.Stat(up); err != nil {
		t.Skip("en-spelling-upstream.txt missing; run vendor script")
	}
	// primary soft list + auto walk-up upstream
	soft := filepath.Join(root, "testdata", "disambiguation", "en-ignore-spelling.txt")
	paths := ignoreSpellingPaths(soft)
	require.GreaterOrEqual(t, len(paths), 2)
	words := map[string]struct{}{}
	for _, p := range paths {
		loaded, err := loadIgnoreSpellingWords(p)
		require.NoError(t, err)
		for k := range loaded {
			words[k] = struct{}{}
		}
	}
	// soft tech name
	_, okChat := words["ChatGPT"]
	_, okChatLow := words["chatgpt"]
	require.True(t, okChat || okChatLow, "soft list should include ChatGPT")
	// upstream spelling.txt sample
	_, okBash := words["Bash"]
	require.True(t, okBash, "upstream spelling should include Bash")
}
