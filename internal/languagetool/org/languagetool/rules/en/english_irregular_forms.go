package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// IrregularForms ports AbstractEnglishSpellerRule.IrregularForms.
type IrregularForms struct {
	BaseForm string
	PosName  string // "verb", "noun", "adjective"
	FormName string // "past tense", "plural", …
	Forms    []string
}

// SynthesizeFn ports language.getSynthesizer().synthesize(token, posTag).
// lemma is base form; token surface is the misspelled word (Java AnalyzedToken).
type SynthesizeFn func(surface, lemma, posTag string) []string

// EnglishIrregularForms ports getIrregularFormsOrNull(word) chain.
// synthesizer nil → always nil (fail-closed, no invent forms).
func EnglishIrregularForms(word string, isMisspelled func(string) bool, synthesize SynthesizeFn) *IrregularForms {
	if word == "" || synthesize == nil {
		return nil
	}
	// order matches Java getIrregularFormsOrNull overload chain
	type arm struct {
		wordSuffix string
		suffixes   []string
		posTag     string
		posName    string
		formName   string
	}
	arms := []arm{
		{"ed", []string{"ed"}, "VBD", "verb", "past tense"},
		{"ed", []string{"d"}, "VBD", "verb", "past tense"},
		{"s", []string{"s"}, "NNS", "noun", "plural"},
		{"es", []string{"es"}, "NNS", "noun", "plural"},
		{"er", []string{"er"}, "JJR", "adjective", "comparative"},
		{"est", []string{"est"}, "JJS", "adjective", "superlative"},
	}
	for _, a := range arms {
		if f := irregularFormsForSuffix(word, a.wordSuffix, a.suffixes, a.posTag, a.posName, a.formName, isMisspelled, synthesize); f != nil {
			return f
		}
	}
	return nil
}

func irregularFormsForSuffix(
	word, wordSuffix string,
	suffixes []string,
	posTag, posName, formName string,
	isMisspelled func(string) bool,
	synthesize SynthesizeFn,
) *IrregularForms {
	if !strings.HasSuffix(word, wordSuffix) {
		return nil
	}
	for _, suffix := range suffixes {
		if !strings.HasSuffix(word, suffix) {
			continue
		}
		// Java: baseForm = word.substring(0, word.length() - suffix.length())
		// with word.endsWith(wordSuffix) already true; suffix is from list (ed/d/s/es/er/est)
		if len(suffix) > len(word) {
			continue
		}
		baseForm := word[:len(word)-len(suffix)]
		forms := synthesize(word, baseForm, posTag)
		var result []string
		for _, form := range forms {
			if isMisspelled != nil && isMisspelled(form) {
				continue
			}
			result = append(result, form)
		}
		// remove self and non-standard
		filtered := result[:0]
		for _, form := range result {
			if form == word || form == "badder" || form == "baddest" || form == "spake" {
				continue
			}
			filtered = append(filtered, form)
		}
		if len(filtered) > 0 {
			return &IrregularForms{
				BaseForm: baseForm,
				PosName:  posName,
				FormName: formName,
				Forms:    filtered,
			}
		}
	}
	return nil
}

// NewAnalyzedTokenForSynth builds AnalyzedToken(word, null, baseForm) for synthesizer.
func NewAnalyzedTokenForSynth(surface, lemma string) *languagetool.AnalyzedToken {
	var lem *string
	if lemma != "" {
		lem = &lemma
	}
	return languagetool.NewAnalyzedToken(surface, nil, lem)
}
