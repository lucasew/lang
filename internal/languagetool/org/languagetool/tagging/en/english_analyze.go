package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
)

// AnalyzeEnglishSentence ports a minimal EN getAnalyzedSentence for speller twins:
// EnglishWordTokenizer + EnglishHybridDisambiguator multiword IGNORE_SPELLING
// (no full POS tagger / XML disambiguator — leaves → root for those sectors).
func AnalyzeEnglishSentence(text string) *languagetool.AnalyzedSentence {
	sent := languagetool.AnalyzeWithTokenizer(text, entok.NewEnglishWordTokenizer())
	if d := DefaultEnglishHybridDisambiguator(); d != nil {
		sent = d.Disambiguate(sent)
	}
	return sent
}
