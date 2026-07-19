package en

import (
	"embed"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	synthen "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/en"
)

//go:embed data/synonyms.txt
var synonymsFS embed.FS

var (
	synOnce sync.Once
	synMap  map[string]*rules.SynonymsData

	enSynthOnce sync.Once
	enSynth     *synthen.EnglishSynthesizer
)

func loadSynonyms() map[string]*rules.SynonymsData {
	synOnce.Do(func() {
		f, err := synonymsFS.Open("data/synonyms.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		m, err := rules.LoadSynonymsWords(f)
		if err != nil {
			panic(err)
		}
		synMap = m
	})
	return synMap
}

// discoverEnglishSynthDict finds english_synth.dict (Java /en/english_synth.dict).
// Empty when missing (fail-closed; no invent forms).
func discoverEnglishSynthDict() string {
	if p := os.Getenv("LANG_ENGLISH_SYNTH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english_synth.dict"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en", "src", "main", "resources", "org", "languagetool", "resource", "en", "english_synth.dict"),
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		for _, rel := range relPaths {
			cand := filepath.Join(dir, rel)
			if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
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

func openDiscoveredEnglishSynthesizer() *synthen.EnglishSynthesizer {
	enSynthOnce.Do(func() {
		if p := discoverEnglishSynthDict(); p != "" {
			enSynth = synthen.OpenEnglishSynthesizerFromDictPath(p)
		}
	})
	return enSynth
}

// EnglishRepeatedWordsRule ports org.languagetool.rules.en.EnglishRepeatedWordsRule.
type EnglishRepeatedWordsRule struct {
	*rules.AbstractRepeatedWordsRule
}

func NewEnglishRepeatedWordsRule(messages map[string]string) *EnglishRepeatedWordsRule {
	// Java getMessage / getDescription / getShortMessage
	// Java: setTags(picky); REPETITIONS_STYLE + Style from abstract ctor.
	base := &rules.AbstractRepeatedWordsRule{
		ID:          "EN_REPEATEDWORDS",
		Description: "Suggest synonyms for repeated words.",
		Message: "This word has been used in one of the immediately preceding sentences. " +
			"Using a synonym could make your text more interesting to read, unless the repetition is intentional.",
		ShortMsg:     "Style: repeated word",
		WordsToCheck: loadSynonyms(),
		LanguageCode: "en",
		Tags:         []rules.Tag{rules.TagPicky},
		// Java ANTI_PATTERNS → getSentenceWithImmunization (24/24)
		SentenceWithImmunization: englishRepeatedWordsSentenceWithImmunization,
	}
	rules.InitRepeatedWordsMeta(base, messages)
	// Java isException: all-upper / mid-sentence capitalized / NNP*
	base.IsException = englishRepeatedWordsIsException
	// Java EnglishSynthesizer.INSTANCE (fail-closed when dict missing → lemma suggestions)
	base.SynthesizeRE = englishRepeatedWordsSynthesizeRE
	return &EnglishRepeatedWordsRule{AbstractRepeatedWordsRule: base}
}

func englishRepeatedWordsIsException(
	tokens []*languagetool.AnalyzedTokenReadings,
	i int,
	sentStart, isCapitalized, isAllUppercase bool,
) bool {
	if isAllUppercase || (isCapitalized && !sentStart) {
		return true
	}
	if i >= 0 && i < len(tokens) && tokens[i] != nil && tokens[i].HasPosTagStartingWith("NNP") {
		return true
	}
	return false
}

func englishRepeatedWordsSynthesizeRE(token *languagetool.AnalyzedToken, posTag string) []string {
	if token == nil {
		return nil
	}
	if s := openDiscoveredEnglishSynthesizer(); s != nil {
		forms, err := s.SynthesizeRE(token, posTag, true)
		if err != nil || len(forms) == 0 {
			return nil
		}
		return forms
	}
	return nil
}

func (r *EnglishRepeatedWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractRepeatedWordsRule.MatchList(sentences)
}

// MinToCheckParagraph ports AbstractRepeatedWordsRule.minToCheckParagraph (Java returns 1).
func (r *EnglishRepeatedWordsRule) MinToCheckParagraph() int { return 1 }
