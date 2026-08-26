package es

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

//go:embed data/synonyms.txt
var synonymsFS embed.FS

var (
	synOnce sync.Once
	synMap  map[string]*rules.SynonymsData

	esSynthOnce sync.Once
	esSynthBase *synthesis.BaseSynthesizer
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

// SpanishRepeatedWordsRule ports org.languagetool.rules.es.SpanishRepeatedWordsRule.
// Synthesizer opens spanish_synth.dict when present; else lemma suggestions (fail-closed).
type SpanishRepeatedWordsRule struct {
	*rules.AbstractRepeatedWordsRule
}

func NewSpanishRepeatedWordsRule(messages map[string]string) *SpanishRepeatedWordsRule {
	// Java getMessage / getDescription / getShortMessage (not invent English).
	// Java: setTags(picky); REPETITIONS_STYLE + Style from abstract ctor.
	base := &rules.AbstractRepeatedWordsRule{
		ID:          "ES_REPEATEDWORDS",
		Description: "Sinónimos para palabras repetidas.",
		Message: "Esta palabra ya ha aparecido en una de las frases inmediatamente anteriores. " +
			"Puede usar un sinónimo para hacer más interesante el texto, excepto si la repetición es intencionada.",
		ShortMsg:     "Estilo: palabra repetida",
		WordsToCheck: loadSynonyms(),
		LanguageCode: "es",
		Tags:         []rules.Tag{rules.TagPicky},
		// Java ANTI_PATTERNS → getSentenceWithImmunization (5/5)
		SentenceWithImmunization: spanishRepeatedWordsSentenceWithImmunization,
	}
	rules.InitRepeatedWordsMeta(base, messages)
	base.IsException = spanishRepeatedWordsIsException
	base.AdjustPostag = spanishRepeatedWordsAdjustPostag
	// Java SpanishSynthesizer.INSTANCE (fail-closed when dict missing → lemma forms).
	base.SynthesizeRE = spanishRepeatedWordsSynthesizeRE
	return &SpanishRepeatedWordsRule{AbstractRepeatedWordsRule: base}
}

// spanishRepeatedWordsIsException ports SpanishRepeatedWordsRule.isException.
func spanishRepeatedWordsIsException(
	tokens []*languagetool.AnalyzedTokenReadings,
	i int,
	sentStart, isCapitalized, isAllUppercase bool,
) bool {
	if isAllUppercase || (isCapitalized && !sentStart) {
		return true
	}
	if i >= 0 && i < len(tokens) && tokens[i] != nil {
		if tokens[i].HasPosTagStartingWith("NP") || tokens[i].HasPosTag("_english_ignore_") {
			return true
		}
	}
	return false
}

// spanishRepeatedWordsAdjustPostag ports SpanishRepeatedWordsRule.adjustPostag
// (StringUtils.replaceOnce — first occurrence only).
func spanishRepeatedWordsAdjustPostag(postag string) string {
	switch {
	case strings.Contains(postag, "CN"):
		return strings.Replace(postag, "CN", "..", 1)
	case strings.Contains(postag, "MS"):
		return strings.Replace(postag, "MS", "[MC][SN]", 1)
	case strings.Contains(postag, "FS"):
		return strings.Replace(postag, "FS", "[FC][SN]", 1)
	case strings.Contains(postag, "MP"):
		return strings.Replace(postag, "MP", "[MC][PN]", 1)
	case strings.Contains(postag, "FP"):
		return strings.Replace(postag, "FP", "[FC][PN]", 1)
	case strings.Contains(postag, "CS"):
		return strings.Replace(postag, "CS", "[MC][SN]", 1)
	case strings.Contains(postag, "CP"):
		return strings.Replace(postag, "CP", "[MC][PN]", 1)
	case strings.Contains(postag, "MN"):
		return strings.Replace(postag, "MN", "[MC][SPN]", 1)
	case strings.Contains(postag, "FN"):
		return strings.Replace(postag, "FN", "[FC][SPN]", 1)
	default:
		return postag
	}
}

func (r *SpanishRepeatedWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractRepeatedWordsRule.MatchList(sentences)
}

// MinToCheckParagraph ports AbstractRepeatedWordsRule.minToCheckParagraph (Java returns 1).
func (r *SpanishRepeatedWordsRule) MinToCheckParagraph() int { return 1 }


func discoverSpanishSynthDict() string {
	if p := os.Getenv("LANG_SPANISH_SYNTH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "es", "src", "main", "resources", "org", "languagetool", "resource", "es", "spanish_synth.dict"),
		filepath.Join("third_party", "spanish-pos-dict", "org", "languagetool", "resource", "es", "spanish_synth.dict"),
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

func openDiscoveredSpanishSynthBase() *synthesis.BaseSynthesizer {
	esSynthOnce.Do(func() {
		if p := discoverSpanishSynthDict(); p != "" {
			esSynthBase = synthesis.OpenBaseSynthesizerFromDictPath("es", p)
		}
	})
	return esSynthBase
}

func spanishRepeatedWordsSynthesizeRE(token *languagetool.AnalyzedToken, posTag string) []string {
	if token == nil {
		return nil
	}
	if base := openDiscoveredSpanishSynthBase(); base != nil {
		forms, err := base.SynthesizeRE(token, posTag, true)
		if err != nil || len(forms) == 0 {
			return nil
		}
		return forms
	}
	return nil
}
