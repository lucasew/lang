package translation

// Twin of org.languagetool.language.en.translation.BeoLingusTranslatorTest
// Uses official beolingus_test.txt and real DE tagger + EN synthesizer (Java Resources).

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	rulesde "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/de"
	"github.com/stretchr/testify/require"
)

func requireBeolingusTestFile(t *testing.T) string {
	t.Helper()
	p := DiscoverBeolingusTestFile()
	if p == "" {
		t.Skip("beolingus_test.txt not discoverable under inspiration/languagetool")
	}
	return p
}

// ensureGermanPOSDictForTests makes german.dict discoverable like Java's german-pos-dict
// dependency on the classpath. Prefer already-present dict; else extract from a known jar.
func ensureGermanPOSDictForTests(t *testing.T) {
	t.Helper()
	if rulesde.DiscoverGermanPOSDict() != "" {
		return
	}
	// Place next to official DE resources so DiscoverGermanPOSDict / DiscoveredGermanTagger find it.
	root := rulesde.DiscoverGermanResourceDir()
	if root == "" {
		t.Skip("German resource dir not discoverable (need inspiration de resources)")
	}
	destDict := filepath.Join(root, "german.dict")
	destInfo := filepath.Join(root, "german.info")
	if st, err := os.Stat(destDict); err == nil && st.Mode().IsRegular() {
		return
	}
	jar := findGermanPOSDictJar()
	if jar == "" {
		t.Skip("german.dict missing and german-pos-dict jar not found (Java de.danielnaber:german-pos-dict)")
	}
	require.NoError(t, extractZipFiles(jar, map[string]string{
		"org/languagetool/resource/de/german.dict": destDict,
		"org/languagetool/resource/de/german.info": destInfo,
	}))
	// DiscoveredGermanTagger is sync.Once — process must not have opened tagger without dict.
	// Tests construct translator after this helper; order is TestXxx → ensure → GetInstance.
	if rulesde.DiscoverGermanPOSDict() == "" {
		t.Fatal("german.dict extract did not become discoverable")
	}
}

func findGermanPOSDictJar() string {
	// Common local installs (ltex-ls ships german-pos-dict-1.2.4.jar — same artifact LT uses).
	candidates := []string{}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, ".local/share/mise/installs/github-valentjn-ltex-ls/16.0.0/lib/german-pos-dict-1.2.4.jar"),
			filepath.Join(home, ".local/share/workspaced/tools/github-valentjn-ltex-ls/16.0.0/lib/german-pos-dict-1.2.4.jar"),
			filepath.Join(home, ".m2/repository/de/danielnaber/german-pos-dict/1.2.4/german-pos-dict-1.2.4.jar"),
		)
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.Mode().IsRegular() {
			return c
		}
	}
	// walk mise installs
	if home, err := os.UserHomeDir(); err == nil {
		base := filepath.Join(home, ".local/share/mise/installs")
		_ = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil {
				return nil
			}
			if !info.IsDir() && filepath.Base(path) == "german-pos-dict-1.2.4.jar" {
				candidates = append(candidates, path)
				return filepath.SkipDir
			}
			return nil
		})
		for _, c := range candidates {
			if st, err := os.Stat(c); err == nil && st.Mode().IsRegular() {
				return c
			}
		}
	}
	return ""
}

func extractZipFiles(zipPath string, nameToDest map[string]string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		dest, ok := nameToDest[f.Name]
		if !ok {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(dest)
		if err != nil {
			rc.Close()
			return err
		}
		_, copyErr := io.Copy(out, rc)
		closeErr := out.Close()
		rc.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		delete(nameToDest, f.Name)
	}
	if len(nameToDest) > 0 {
		return os.ErrNotExist
	}
	return nil
}

func newTestBeoLingusTranslator(t *testing.T) *BeoLingusTranslator {
	t.Helper()
	ensureGermanPOSDictForTests(t)
	// Inflector needs english_synth for plural forms in inflected tests
	if discoverEnglishSynthDict() == "" || defaultEnglishSynthesizer() == nil {
		t.Skip("english_synth.dict not discoverable — Inflector needs Languages en synthesizer like Java")
	}
	if rulesde.DiscoveredGermanTagger() == nil {
		t.Skip("German tagger unavailable (german.dict / added.txt)")
	}
	path := requireBeolingusTestFile(t)
	ResetBeoLingusInstanceForTest()
	cfg := &languagetool.GlobalConfig{}
	cfg.SetBeolingusFile(path)
	tr, err := GetInstance(cfg)
	require.NoError(t, err)
	require.NotNil(t, tr)
	return tr
}

// Port of BeoLingusTranslatorTest.testTranslateInflectedForm
func TestBeoLingusTranslator_TranslateInflectedForm(t *testing.T) {
	translator := newTestBeoLingusTranslator(t)
	result1, err := translator.Translate("Luftpumpen", "de", "en")
	require.NoError(t, err)
	require.Equal(t, 1, len(result1))
	require.Contains(t, result1[0].GetL2(), "tyre pumps [Br.]")
	require.Contains(t, result1[0].GetL2(), "tire pumps [Am.]")
	require.Contains(t, result1[0].GetL2(), "tyre inflators [Br.]")
	require.Contains(t, result1[0].GetL2(), "tire inflators [Am.]")
}

// Port of BeoLingusTranslatorTest.testTranslate
func TestBeoLingusTranslator_Translate(t *testing.T) {
	translator := newTestBeoLingusTranslator(t)

	result1, err := translator.Translate("Haus", "de", "en")
	require.NoError(t, err)
	require.Equal(t, 3, len(result1))
	require.Contains(t, result1[0].GetL2(), "house")
	require.Contains(t, result1[1].GetL2(), "home")
	require.Contains(t, result1[2].GetL2(), "volta bracket (sheet music)")

	result2, err := translator.Translate("haus", "de", "en")
	require.NoError(t, err)
	require.Equal(t, 3, len(result2))
	require.Contains(t, result2[0].GetL2(), "house")
	require.Contains(t, result2[1].GetL2(), "home")
	require.Contains(t, result2[2].GetL2(), "volta bracket (sheet music)")

	result4, err := translator.Translate("suchwort", "de", "en")
	require.NoError(t, err)
	require.Equal(t, 1, len(result4))
	require.Contains(t, result4[0].GetL2(), "search word {one; to}")
	require.Contains(t, result4[0].GetL2(), "another item")
}

// Port of BeoLingusTranslatorTest.testSplit
func TestBeoLingusTranslator_Split(t *testing.T) {
	translator := newTestBeoLingusTranslator(t)
	// Java assertThat(...toString(), is("[foo]")) — List.toString
	require.Equal(t, "[foo]", listString(translator.Split("foo")))
	require.Equal(t, "[foo { bar } foo]", listString(translator.Split("foo { bar } foo")))
	require.Equal(t, "[foo { bar }, foo]", listString(translator.Split("foo { bar }; foo")))
	require.Equal(t, "[foo, bar, foo]", listString(translator.Split("foo; bar; foo")))
	require.Equal(t, "[foo, bar { blah }, foo]", listString(translator.Split("foo; bar { blah }; foo")))
	require.Equal(t, "[foo, bar { blah; blubb }, foo]", listString(translator.Split("foo; bar { blah; blubb }; foo")))
	require.Equal(t, "[foo, bar { blah; blubb; three four }, foo]", listString(translator.Split("foo; bar { blah; blubb; three four }; foo")))
}

// Port of BeoLingusTranslatorTest.testAmericanBritishVariants
func TestBeoLingusTranslator_AmericanBritishVariants(t *testing.T) {
	translator := newTestBeoLingusTranslator(t)

	result1, err := translator.Translate("Luftpumpe", "de", "en")
	require.NoError(t, err)
	require.Equal(t, 1, len(result1))
	require.Contains(t, result1[0].GetL2(), "tyre pump [Br.]")
	require.Contains(t, result1[0].GetL2(), "tire pump [Am.]")
	require.Contains(t, result1[0].GetL2(), "tyre inflator [Br.]")
	require.Contains(t, result1[0].GetL2(), "tire inflator [Am.]")

	result2, err := translator.Translate("Testluftpumpe", "de", "en")
	require.NoError(t, err)
	require.Equal(t, 1, len(result2))
	require.Contains(t, result2[0].GetL2(), "tyre pump [Br.]")
	require.Contains(t, result2[0].GetL2(), "tire pump [Am.]")

	result3, err := translator.Translate("Reifen wechseln", "de", "en")
	require.NoError(t, err)
	require.Equal(t, 1, len(result3))
	require.Contains(t, result3[0].GetL2(), "to change the tyres [Br.]")
	require.Contains(t, result3[0].GetL2(), "to change the tires [Am.]")
}

// Port of BeoLingusTranslatorTest.testCleanTranslationForReplace
func TestBeoLingusTranslator_CleanTranslationForReplace(t *testing.T) {
	translator := newTestBeoLingusTranslator(t)
	// Java null prevWord → Go ""
	require.Equal(t, "", translator.CleanTranslationForReplace("", ""))
	require.Equal(t, "go", translator.CleanTranslationForReplace("to go", ""))
	require.Equal(t, "to go", translator.CleanTranslationForReplace("to go", "need"))
	require.Equal(t, "go", translator.CleanTranslationForReplace("to go", "will"))
	require.Equal(t, "go", translator.CleanTranslationForReplace("to go", "foo"))
	require.Equal(t, "go", translator.CleanTranslationForReplace("to go", "to"))
	require.Equal(t, "foo", translator.CleanTranslationForReplace("foo (bar) {mus}", ""))
	require.Equal(t, "some thing , something", translator.CleanTranslationForReplace("some thing [Br.], something", ""))
	require.Equal(t, "Friday", translator.CleanTranslationForReplace("Friday /Fri/", ""))
	require.Equal(t, "demise", translator.CleanTranslationForReplace("demise [poet.] <death>", ""))
}

// Port of BeoLingusTranslatorTest.testGetTranslationSuffix
func TestBeoLingusTranslator_GetTranslationSuffix(t *testing.T) {
	translator := newTestBeoLingusTranslator(t)
	require.Equal(t, "", translator.GetTranslationSuffix(""))
	require.Equal(t, "", translator.GetTranslationSuffix(" "))
	require.Equal(t, "", translator.GetTranslationSuffix("foo bar"))
	require.Equal(t, "[Br.]", translator.GetTranslationSuffix("foo bar [Br.]"))
	require.Equal(t, "{ugs} [Br.]", translator.GetTranslationSuffix("foo bar {ugs} [Br.]"))
	require.Equal(t, "{ugs} [Br.] (Blah)", translator.GetTranslationSuffix("foo bar {ugs} [Br.] (Blah)"))
	require.Equal(t, "<blah>", translator.GetTranslationSuffix("foo bar <blah>"))
}

// DataSource / Message twin of Java getters
func TestBeoLingusTranslator_MessageAndDataSource(t *testing.T) {
	translator := newTestBeoLingusTranslator(t)
	require.Equal(t, "Translate to English?", translator.GetMessage())
	ds := translator.GetDataSource()
	require.Equal(t, "https://www.gnu.org/licenses/old-licenses/gpl-2.0.en.html", ds.LicenseURL)
	require.Equal(t, "BEOLINGUS", ds.SourceName)
	require.Equal(t, "http://dict.tu-chemnitz.de", ds.SourceURL)
}

// listString ports Java List.toString for string lists.
func listString(parts []string) string {
	if parts == nil {
		return "null"
	}
	// Java ArrayList.toString: [a, b, c]
	return "[" + joinCommaSpace(parts) + "]"
}

func joinCommaSpace(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += ", " + parts[i]
	}
	return out
}
