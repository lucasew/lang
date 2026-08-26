package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/ConsistentApostrophesRuleTest.java
import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	enTok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
	"github.com/stretchr/testify/require"
)

// analyzeLikeJavaEN approximates Java JLanguageTool.analyzeText for English apostrophe tests:
// EnglishWordTokenizer (contraction splits) + EnglishTagger-style ’→' + hasTypographicApostrophe.
// Positions are UTF-16 (Java String indices) via tokenizers.BuildPositions.
func analyzeLikeJavaEN(text string) []*languagetool.AnalyzedSentence {
	return []*languagetool.AnalyzedSentence{analyzeEnglishPlain(text)}
}

func analyzeEnglishPlain(text string) *languagetool.AnalyzedSentence {
	ensureEnglishTokenizerTaggerForTest()
	wt := enTok.NewEnglishWordTokenizer()
	raw := wt.Tokenize(text)
	positions := tokenizers.BuildPositions(raw)
	readings := make([]*languagetool.AnalyzedTokenReadings, 0, len(raw)+1)
	ss := languagetool.SentenceStartTagName
	startTok := languagetool.NewAnalyzedToken("", &ss, nil)
	startR := languagetool.NewAnalyzedTokenReadings(startTok)
	startR.SetStartPos(0)
	readings = append(readings, startR)
	prev := ""
	for i, tok := range raw {
		// Java EnglishTagger.tag: if word.length() > 1 && contains(’) → rewrite + setTypographicApostrophe
		typoApos := false
		if tokenizers.UTF16Len(tok) > 1 && strings.Contains(tok, "’") {
			typoApos = true
			tok = strings.ReplaceAll(tok, "’", "'")
		}
		at := languagetool.NewAnalyzedToken(tok, nil, nil)
		ar := languagetool.NewAnalyzedTokenReadingsAt(at, positions[i])
		if typoApos {
			ar.SetTypographicApostrophe(true)
		}
		if prev != "" {
			ar.SetWhitespaceBeforeToken(prev)
		}
		readings = append(readings, ar)
		prev = tok
	}
	return languagetool.NewAnalyzedSentence(readings)
}

var enTokTaggerMu sync.Mutex

// ensureEnglishTokenizerTaggerForTest wires EnglishTagger for EnglishWordTokenizer.wordsToAdd
// (Java: EnglishTagger.INSTANCE). Needed so "'s"/"n't" stay intact like Java analyzeText.
func ensureEnglishTokenizerTaggerForTest() {
	if enTok.IsTaggedEN != nil {
		return
	}
	enTokTaggerMu.Lock()
	defer enTokTaggerMu.Unlock()
	if enTok.IsTaggedEN != nil {
		return
	}
	p := discoverEnglishPOSDictForTokenizerTest()
	if p == "" {
		return
	}
	lt := languagetool.NewJLanguageTool("en")
	_ = RegisterBinaryEnglishTagger(lt, p)
}

func discoverEnglishPOSDictForTokenizerTest() string {
	if p := os.Getenv("LANG_ENGLISH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english.dict"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
			"src", "main", "resources", "org", "languagetool", "resource", "en", "english.dict"),
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 12; i++ {
		for _, rel := range relPaths {
			cand := filepath.Join(dir, rel)
			if st, e := os.Stat(cand); e == nil && st.Mode().IsRegular() {
				return cand
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func TestConsistentApostrophesRule_Rule(t *testing.T) {
	// Java: ConsistentApostrophesRule rule = new ConsistentApostrophesRule(TestTools.getEnglishMessages());
	//       JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("en"));
	rule := NewConsistentApostrophesRule(nil)

	// Java: rule.match(lt.analyzeText("It's a nice idea. But it doesn’t work."));
	matches := rule.MatchList(analyzeLikeJavaEN("It's a nice idea. But it doesn’t work."))
	require.Equal(t, 2, len(matches))
	require.Equal(t, 2, matches[0].GetFromPos())
	require.Equal(t, 4, matches[0].GetToPos())
	require.Equal(t, "[’s]", formatSuggested(matches[0].GetSuggestedReplacements()))
	require.Equal(t, 29, matches[1].GetFromPos())
	require.Equal(t, 32, matches[1].GetToPos())
	require.Equal(t, "[n't]", formatSuggested(matches[1].GetSuggestedReplacements()))
	// Java messages (not asserted in Java test, but surface of twin match())
	require.Contains(t, matches[0].GetMessage(), "typewriter-style apostrophe")
	require.Contains(t, matches[0].GetMessage(), "Both are correct, but consider using the same type everywhere in your text.")
	require.Contains(t, matches[1].GetMessage(), "typographic apostrophe")
	require.Contains(t, matches[1].GetMessage(), "Both are correct, but consider using the same type everywhere in your text.")

	// Java: assertThat(rule.match(lt.analyzeText("It’s a nice idea. But it doesn't work.")).length, is(2));
	require.Equal(t, 2, len(rule.MatchList(analyzeLikeJavaEN("It’s a nice idea. But it doesn't work."))))
	// Java: assertThat(rule.match(lt.analyzeText("It's a nice idea. But it doesn't work.")).length, is(0));
	require.Equal(t, 0, len(rule.MatchList(analyzeLikeJavaEN("It's a nice idea. But it doesn't work."))))
	// Java: assertThat(rule.match(lt.analyzeText("It’s a nice idea. But it doesn’t work.")).length, is(0));
	require.Equal(t, 0, len(rule.MatchList(analyzeLikeJavaEN("It’s a nice idea. But it doesn’t work."))))
}

// Java ConsistentApostrophesRule ctor: setDefaultTempOff, URL, example pair; minToCheckParagraph -1.
func TestConsistentApostrophesRule_Metadata(t *testing.T) {
	rule := NewConsistentApostrophesRule(nil)
	require.Equal(t, "EN_CONSISTENT_APOS", rule.GetID())
	require.Equal(t, "Checks if the two types of apostrophes (' and ’) are used consistently in a text.", rule.GetDescription())
	require.Equal(t, "https://languagetool.org/insights/post/punctuation-guide/#what-is-an-apostrophe", rule.GetURL())
	require.True(t, rule.IsDefaultTempOff())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "It's nice, but it <marker>doesn’t</marker> work.", inc[0].GetExample())
	require.Equal(t, []string{"doesn't"}, inc[0].GetCorrections())
	require.Equal(t, "It's nice, but it <marker>doesn't</marker> work.", rule.GetCorrectExamples()[0].GetExample())
}

func formatSuggested(s []string) string {
	if len(s) == 0 {
		return "[]"
	}
	return "[" + s[0] + "]"
}
