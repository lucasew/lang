package fr

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

	frSynthOnce sync.Once
	frSynthBase *synthesis.BaseSynthesizer
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

// FrenchRepeatedWordsRule ports org.languagetool.rules.fr.FrenchRepeatedWordsRule.
// Synthesizer opens french_synth.dict when present; else lemma suggestions (fail-closed).
type FrenchRepeatedWordsRule struct {
	*rules.AbstractRepeatedWordsRule
}

func NewFrenchRepeatedWordsRule(messages map[string]string) *FrenchRepeatedWordsRule {
	// Java getMessage / getDescription / getShortMessage
	// Java: REPETITIONS_STYLE + Style (no picky tag on FR).
	base := &rules.AbstractRepeatedWordsRule{
		ID:          "FR_REPEATEDWORDS",
		Description: "Synonymes de mots répétés.",
		Message: "Ce mot apparaît déjà dans l'une des phrases précédant immédiatement celle-ci. " +
			"Utilisez un synonyme pour apporter plus de variété à votre texte, excepté si la répétition est intentionnelle.",
		ShortMsg:     "Style : Mot répété",
		WordsToCheck: loadSynonyms(),
		LanguageCode: "fr",
	}
	rules.InitRepeatedWordsMeta(base, messages)
	base.IsException = frenchRepeatedWordsIsException
	base.AdjustPostag = frenchRepeatedWordsAdjustPostag
	// FrenchSynthesizer.INSTANCE when french_synth.dict discoverable (fail-closed lemma forms).
	base.SynthesizeRE = frenchRepeatedWordsSynthesizeRE
	return &FrenchRepeatedWordsRule{AbstractRepeatedWordsRule: base}
}

// frenchRepeatedWordsIsException ports FrenchRepeatedWordsRule.isException.
func frenchRepeatedWordsIsException(
	tokens []*languagetool.AnalyzedTokenReadings,
	i int,
	sentStart, isCapitalized, isAllUppercase bool,
) bool {
	if isAllUppercase || (isCapitalized && !sentStart) {
		return true
	}
	if i >= 0 && i < len(tokens) && tokens[i] != nil && tokens[i].HasPosTagStartingWith("Z") {
		return true
	}
	return false
}

// frenchRepeatedWordsAdjustPostag ports FrenchRepeatedWordsRule.adjustPostag
// (StringUtils.replaceOnce on suffix-matched postag).
func frenchRepeatedWordsAdjustPostag(postag string) string {
	switch {
	case strings.HasSuffix(postag, "e sp"):
		return strings.Replace(postag, "e sp", ". .*", 1)
	case strings.HasSuffix(postag, "m s"):
		return strings.Replace(postag, "m s", "[me] sp?", 1)
	case strings.HasSuffix(postag, "f s"):
		return strings.Replace(postag, "f s", "[fe] sp?", 1)
	case strings.HasSuffix(postag, "m p"):
		return strings.Replace(postag, "m p", "[me] s?p", 1)
	case strings.HasSuffix(postag, "f p"):
		return strings.Replace(postag, "f p", "[fe] s?p", 1)
	case strings.HasSuffix(postag, "e s"):
		return strings.Replace(postag, "e s", "[me] sp?", 1)
	case strings.HasSuffix(postag, "e p"):
		return strings.Replace(postag, "e p", "[me] s?p", 1)
	case strings.HasSuffix(postag, "m sp"):
		return strings.Replace(postag, "m sp", "[me] s?p?", 1)
	case strings.HasSuffix(postag, "f sp"):
		return strings.Replace(postag, "f sp", "[fe] s?p?", 1)
	default:
		return postag
	}
}

func (r *FrenchRepeatedWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractRepeatedWordsRule.MatchList(sentences)
}

// MinToCheckParagraph ports AbstractRepeatedWordsRule.minToCheckParagraph (Java returns 1).
func (r *FrenchRepeatedWordsRule) MinToCheckParagraph() int { return 1 }

// discoverFrenchSynthDict finds french_synth.dict (Java /fr/french_synth.dict). Empty = fail-closed.
func discoverFrenchSynthDict() string {
	if p := os.Getenv("LANG_FRENCH_SYNTH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "fr", "src", "main", "resources", "org", "languagetool", "resource", "fr", "french_synth.dict"),
		filepath.Join("third_party", "french-pos-dict", "org", "languagetool", "resource", "fr", "french_synth.dict"),
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

func openDiscoveredFrenchSynthBase() *synthesis.BaseSynthesizer {
	frSynthOnce.Do(func() {
		if p := discoverFrenchSynthDict(); p != "" {
			// Java FrenchSynthesizer.INSTANCE → BaseSynthesizer resources; fail-closed if missing.
			frSynthBase = synthesis.OpenBaseSynthesizerFromDictPath("fr", p)
		}
	})
	return frSynthBase
}

// frenchRepeatedWordsSynthesizeRE ports FrenchSynthesizer.INSTANCE.synthesize(token, postag, true).
func frenchRepeatedWordsSynthesizeRE(token *languagetool.AnalyzedToken, posTag string) []string {
	if token == nil {
		return nil
	}
	if base := openDiscoveredFrenchSynthBase(); base != nil {
		forms, err := base.SynthesizeRE(token, posTag, true)
		if err != nil || len(forms) == 0 {
			return nil
		}
		return forms
	}
	return nil
}
