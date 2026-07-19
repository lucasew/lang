package ca

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

	caSynthOnce sync.Once
	caSynthBase *synthesis.BaseSynthesizer
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

// CatalanRepeatedWordsRule ports org.languagetool.rules.ca.CatalanRepeatedWordsRule.
// Synthesizer opens catalan_synth.dict when present; else lemma suggestions (fail-closed).
type CatalanRepeatedWordsRule struct {
	*rules.AbstractRepeatedWordsRule
}

func NewCatalanRepeatedWordsRule(messages map[string]string) *CatalanRepeatedWordsRule {
	// Java getMessage / getDescription / getShortMessage (not invent English).
	// Java: setTags(picky); REPETITIONS_STYLE + Style from abstract ctor.
	base := &rules.AbstractRepeatedWordsRule{
		ID:          "CA_REPEATEDWORDS",
		Description: "Sinònims per a paraules repetides.",
		Message: "Aquesta paraula apareix en una de les frases anteriors. " +
			"Podeu substituir-la per un sinònim per a fer més variat el text, llevat que la repetició sigui intencionada.",
		ShortMsg:     "Estil: paraula repetida",
		WordsToCheck: loadSynonyms(),
		LanguageCode: "ca",
		Tags:         []rules.Tag{rules.TagPicky},
		// Java ANTI_PATTERNS → getSentenceWithImmunization (1/1)
		SentenceWithImmunization: catalanRepeatedWordsSentenceWithImmunization,
	}
	rules.InitRepeatedWordsMeta(base, messages)
	base.IsException = catalanRepeatedWordsIsException
	base.AdjustPostag = catalanRepeatedWordsAdjustPostag
	// Java language.getSynthesizer() (fail-closed when dict missing → lemma forms).
	base.SynthesizeRE = catalanRepeatedWordsSynthesizeRE
	return &CatalanRepeatedWordsRule{AbstractRepeatedWordsRule: base}
}

// catalanRepeatedWordsIsException ports CatalanRepeatedWordsRule.isException.
func catalanRepeatedWordsIsException(
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

// catalanRepeatedWordsAdjustPostag ports CatalanRepeatedWordsRule.adjustPostag
// (replaceFirst — first occurrence only). CA differs from ES on CS/CP (MFC vs MC).
func catalanRepeatedWordsAdjustPostag(postag string) string {
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
		return strings.Replace(postag, "CS", "[MFC][SN]", 1)
	case strings.Contains(postag, "CP"):
		return strings.Replace(postag, "CP", "[MFC][PN]", 1)
	case strings.Contains(postag, "MN"):
		return strings.Replace(postag, "MN", "[MC][SPN]", 1)
	case strings.Contains(postag, "FN"):
		return strings.Replace(postag, "FN", "[FC][SPN]", 1)
	default:
		return postag
	}
}

func (r *CatalanRepeatedWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractRepeatedWordsRule.MatchList(sentences)
}

// MinToCheckParagraph ports AbstractRepeatedWordsRule.minToCheckParagraph (Java returns 1).
func (r *CatalanRepeatedWordsRule) MinToCheckParagraph() int { return 1 }


func discoverCatalanSynthDict() string {
	if p := os.Getenv("LANG_CATALAN_SYNTH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ca", "src", "main", "resources", "org", "languagetool", "resource", "ca", "catalan_synth.dict"),
		filepath.Join("third_party", "catalan-pos-dict", "org", "languagetool", "resource", "ca", "catalan_synth.dict"),
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

func openDiscoveredCatalanSynthBase() *synthesis.BaseSynthesizer {
	caSynthOnce.Do(func() {
		if p := discoverCatalanSynthDict(); p != "" {
			caSynthBase = synthesis.OpenBaseSynthesizerFromDictPath("ca", p)
		}
	})
	return caSynthBase
}

func catalanRepeatedWordsSynthesizeRE(token *languagetool.AnalyzedToken, posTag string) []string {
	if token == nil {
		return nil
	}
	if base := openDiscoveredCatalanSynthBase(); base != nil {
		forms, err := base.SynthesizeRE(token, posTag, true)
		if err != nil || len(forms) == 0 {
			return nil
		}
		return forms
	}
	return nil
}
