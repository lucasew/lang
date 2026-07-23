package sr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

// Java SerbianTagger paths:
// BASE_DICTIONARY_PATH = "/sr/dictionary"
// EKAVIAN_DICTIONARY_PATH = BASE_DICTIONARY_PATH + "/ekavian/"
// default dict = EKAVIAN_DICTIONARY_PATH + "serbian.dict"
const (
	SerbianBaseDictPath    = "/sr/dictionary"
	EkavianDictionaryDir   = SerbianBaseDictPath + "/ekavian/"
	JekavianDictionaryDir  = SerbianBaseDictPath + "/jekavian/"
	EkavianDictionaryPath  = EkavianDictionaryDir + "serbian.dict"
	JekavianDictionaryPath = JekavianDictionaryDir + "serbian.dict"
	// Java SerbianTagger.getManualAdditionsFileName (singular override).
	SerbianManualAdditionsPath = "/sr/dictionary/added.txt"
)

// SerbianTagger ports org.languagetool.tagging.sr.SerbianTagger (Ekavian default).
// Java: super(EKAVIAN_DICTIONARY_PATH + "serbian.dict", new Locale("sr"))
// → tagLowercaseWithUppercase true (BaseTagger 2-arg ctor).
type SerbianTagger struct {
	*tagging.BaseTagger
}

// NewSerbianTagger builds a SerbianTagger over the given WordTagger (Ekavian dict path).
func NewSerbianTagger(wt tagging.WordTagger) *SerbianTagger {
	return NewSerbianTaggerWithPath(wt, EkavianDictionaryPath)
}

// NewSerbianTaggerWithPath ports SerbianTagger(String fileName, Locale conversionLocale).
func NewSerbianTaggerWithPath(wt tagging.WordTagger, path string) *SerbianTagger {
	return &SerbianTagger{BaseTagger: tagging.NewBaseTagger(wt, path, "sr", true)}
}

// Tag ports BaseTagger.tag via getAnalyzedTokens (SerbianTagger has no Java override).
func (t *SerbianTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		var readings []*languagetool.AnalyzedToken
		for _, tw := range t.TagWord(word) {
			readings = append(readings, tagged(word, tw))
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
		}
		out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
		pos += tagging.UTF16Len(word)
	}
	return out
}

// EkavianTagger ports org.languagetool.tagging.sr.EkavianTagger.
// Java: super(EKAVIAN_DICTIONARY_PATH + "serbian.dict", new Locale("sr"));
// manuals: ekavian/added.txt, ekavian/removed.txt.
type EkavianTagger struct{ *SerbianTagger }

// NewEkavianTagger builds an EkavianTagger over the given WordTagger.
func NewEkavianTagger(wt tagging.WordTagger) *EkavianTagger {
	return &EkavianTagger{SerbianTagger: NewSerbianTaggerWithPath(wt, EkavianDictionaryPath)}
}

// JekavianTagger ports org.languagetool.tagging.sr.JekavianTagger.
// Java: super(jekavian/serbian.dict, new Locale("sr"));
// manuals: jekavian/added.txt, jekavian/removed.txt.
type JekavianTagger struct{ *SerbianTagger }

// NewJekavianTagger builds a JekavianTagger over the given WordTagger.
func NewJekavianTagger(wt tagging.WordTagger) *JekavianTagger {
	return &JekavianTagger{SerbianTagger: NewSerbianTaggerWithPath(wt, JekavianDictionaryPath)}
}

func tagged(surface string, tw tagging.TaggedWord) *languagetool.AnalyzedToken {
	var pos, lemma *string
	if tw.PosTag != "" {
		p := tw.PosTag
		pos = &p
	}
	if tw.Lemma != "" {
		l := tw.Lemma
		lemma = &l
	}
	return languagetool.NewAnalyzedToken(surface, pos, lemma)
}
