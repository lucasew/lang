package language

import (
	plsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pl"
	pldis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/pl"
	pltag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	pltok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pl"
)

// NewPolish ports new Polish().
func NewPolish() PolishLang {
	return Polish
}

// PolishShortCode ports Polish.getShortCode ("pl").
func PolishShortCode() string { return Polish.ShortCode }

// CreateDefaultTagger ports Polish.createDefaultTagger → new PolishTagger().
// Loads official /pl/polish.dict when discoverable (Java BaseTagger path).
// Nil if dict is missing (fail-closed).
func (v PolishLang) CreateDefaultTagger() *pltag.PolishTagger {
	pltag.EnsureDefaultPolishTagger()
	return pltag.DefaultPolishTagger
}

// CreateDefaultSentenceTokenizer ports Polish.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "pl".
func (v PolishLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Polish.createDefaultWordTokenizer:
//
//	PolishWordTokenizer wordTokenizer = new PolishWordTokenizer();
//	wordTokenizer.setTagger(getTagger());
//	return wordTokenizer;
func (v PolishLang) CreateDefaultWordTokenizer() *pltok.PolishWordTokenizer {
	wt := pltok.NewPolishWordTokenizer()
	if tagger := v.CreateDefaultTagger(); tagger != nil {
		wt.SetTagger(polishTaggerAsHyphen(tagger))
	}
	return wt
}

// polishTaggerAsHyphen adapts *pltag.PolishTagger to PolishHyphenTagger
// (Java setTagger(getTagger())).
func polishTaggerAsHyphen(t *pltag.PolishTagger) pltok.PolishHyphenTagger {
	if t == nil {
		return nil
	}
	return pltok.ATRTagFunc(func(tokens []string) []pltok.PolishHyphenReadings {
		atrs := t.Tag(tokens)
		out := make([]pltok.PolishHyphenReadings, len(atrs))
		for i, a := range atrs {
			out[i] = a
		}
		return out
	})
}

// CreateDefaultDisambiguator ports Polish.createDefaultDisambiguator
// → new PolishHybridDisambiguator().
func (v PolishLang) CreateDefaultDisambiguator() *pldis.PolishHybridDisambiguator {
	return pldis.NewPolishHybridDisambiguator()
}

// CreateDefaultSynthesizer ports Polish.createDefaultSynthesizer
// → PolishSynthesizer.INSTANCE (lazy official load).
// Nil if polish_synth.dict cannot be opened (fail-closed).
func (v PolishLang) CreateDefaultSynthesizer() *plsynth.PolishSynthesizer {
	return plsynth.Instance()
}

// GetPriorityForId ports Polish.getPriorityForId.
func (v PolishLang) GetPriorityForId(id string) int {
	return PolishPriorityForId(id)
}
