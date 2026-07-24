package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
)

// AnalyzeEnglishSentence ports EN getAnalyzedSentence for faithful speller/tagger twins:
// EnglishWordTokenizer + EnglishTagger (english.dict) + EnglishHybridDisambiguator
// multiword IGNORE_SPELLING. XML disambiguator is optional (not required for speller).
func AnalyzeEnglishSentence(text string) *languagetool.AnalyzedSentence {
	EnsureDefaultEnglishTagger() // wires IsTaggedEN for tokenizer apostrophe keep
	tagWord := EnglishTagWord()
	var sent *languagetool.AnalyzedSentence
	if tagWord != nil {
		sent = languagetool.AnalyzeWithTaggerAndTokenizer(text, tagWord, entok.NewEnglishWordTokenizer())
	} else {
		sent = languagetool.AnalyzeWithTokenizer(text, entok.NewEnglishWordTokenizer())
	}
	if d := DefaultEnglishHybridDisambiguator(); d != nil {
		sent = d.Disambiguate(sent)
	}
	return sent
}
