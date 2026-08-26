package language

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
	rusynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ru"
	rudis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/ru"
	rutag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ru"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	rutok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ru"
)

// NewRussian ports new Russian() / Russian.getInstance() language surface.
func NewRussian() RussianLang {
	return Russian
}

// RussianShortCode ports Russian.getShortCode ("ru").
func RussianShortCode() string { return Russian.ShortCode }

// CreateDefaultTagger ports Russian.createDefaultTagger → RussianTagger.INSTANCE.
// Loads official /ru/russian.dict when discoverable (Java BaseTagger path).
func (v RussianLang) CreateDefaultTagger() *rutag.RussianTagger {
	rutag.EnsureDefaultRussianTagger()
	return rutag.DefaultRussianTagger
}

// CreateDefaultDisambiguator ports Russian.createDefaultDisambiguator
// → RussianHybridDisambiguator.getInstance().
func (v RussianLang) CreateDefaultDisambiguator() *rudis.RussianHybridDisambiguator {
	return rudis.GetInstance()
}

// CreateDefaultPostDisambiguationChunker ports
// Russian.createDefaultPostDisambiguationChunker → new RussianChunker().
// Pre-disambiguation Chunker is null for Russian (Java getChunker).
func (v RussianLang) CreateDefaultPostDisambiguationChunker() *chunking.RussianChunker {
	return chunking.NewRussianChunker()
}

// CreateDefaultSynthesizer ports Russian.createDefaultSynthesizer
// → RussianSynthesizer.INSTANCE (lazy official load).
// Nil if russian_synth.dict cannot be opened (fail-closed).
func (v RussianLang) CreateDefaultSynthesizer() *rusynth.RussianSynthesizer {
	return rusynth.Instance()
}

// CreateDefaultSentenceTokenizer ports Russian.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "ru".
func (v RussianLang) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Russian.createDefaultWordTokenizer
// → new RussianWordTokenizer().
func (v RussianLang) CreateDefaultWordTokenizer() *rutok.RussianWordTokenizer {
	return rutok.NewRussianWordTokenizer()
}

// GetPriorityForId ports Russian.getPriorityForId.
func (v RussianLang) GetPriorityForId(id string) int {
	return RussianPriorityForId(id)
}

// GetRelevantLanguageModelRuleIDs ports Russian.getRelevantLanguageModelRules
// → RussianConfusionProbabilityRule only (CONFUSION_RULE).
func (v RussianLang) GetRelevantLanguageModelRuleIDs() []string {
	return RussianLanguageModelRelevantRuleIDs()
}
