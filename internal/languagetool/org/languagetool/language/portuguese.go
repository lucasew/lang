package language

import (
	ptsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pt"
	ptdis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/pt"
	pttag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	pttok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pt"
)

// Portuguese defaults to Portugal Portuguese (product surface for short code "pt").
// Java Languages often resolves "pt" to the default variant pt-PT.
var Portuguese = PortugalPortuguese

// NewPortuguese ports Languages.getLanguageForShortCode("pt") default surface → pt-PT.
func NewPortuguese() PortugueseVariant { return Portuguese }

func NewPortugalPortuguese() PortugueseVariant   { return PortugalPortuguese }
func NewBrazilianPortuguese() PortugueseVariant  { return BrazilianPortuguese }
func NewAngolaPortuguese() PortugueseVariant     { return AngolaPortuguese }
func NewMozambiquePortuguese() PortugueseVariant { return MozambiquePortuguese }

// PortugueseShortCode ports Portuguese.getShortCode ("pt").
func PortugueseShortCode() string { return "pt" }

// PortugueseBaseCountries ports Portuguese.getCountries (base language, not locale variants).
// Java: return new String[]{"", "CV", "GW", "MO", "ST", "TL"};
func PortugueseBaseCountries() []string {
	return []string{"", "CV", "GW", "MO", "ST", "TL"}
}

// CreateDefaultTagger ports Portuguese.createDefaultTagger → new PortugueseTagger().
// Loads official /pt/portuguese.dict when discoverable (Java BaseTagger path).
// Nil if dict is missing (fail-closed).
func (v PortugueseVariant) CreateDefaultTagger() *pttag.PortugueseTagger {
	pttag.EnsureDefaultPortugueseTagger()
	return pttag.DefaultPortugueseTagger
}

// CreateDefaultSentenceTokenizer ports Portuguese.createDefaultSentenceTokenizer
// → new SRXSentenceTokenizer(this) with short code "pt".
func (v PortugueseVariant) CreateDefaultSentenceTokenizer() tokenizers.SentenceTokenizer {
	return tokenizers.NewSRXSentenceTokenizer(v.GetShortCode())
}

// CreateDefaultWordTokenizer ports Portuguese.createDefaultWordTokenizer
// → new PortugueseWordTokenizer() (Java field: tagger = new PortugueseTagger()).
// EnsureDefaultPortugueseTagger wires the official-dict loader used by
// NewPortugueseWordTokenizer (same as Java ctor).
func (v PortugueseVariant) CreateDefaultWordTokenizer() *pttok.PortugueseWordTokenizer {
	// Java: return new PortugueseWordTokenizer(); // embeds new PortugueseTagger()
	pttag.EnsureDefaultPortugueseTagger()
	return pttok.NewPortugueseWordTokenizer()
}

// CreateDefaultDisambiguator ports Portuguese.createDefaultDisambiguator
// → new PortugueseHybridDisambiguator(getDefaultLanguageVariant()).
// Go hybrid ctor loads official multiwords / spelling_global / disambiguation.xml
// (language arg for XmlRuleDisambiguator is the PT default variant in Java; Go twin
// loads the same pt resources without a Language object).
func (v PortugueseVariant) CreateDefaultDisambiguator() *ptdis.PortugueseHybridDisambiguator {
	return ptdis.NewPortugueseHybridDisambiguator()
}

// CreateDefaultSynthesizer ports Portuguese.createDefaultSynthesizer
// → PortugueseSynthesizer.INSTANCE (lazy official load).
// Nil if portuguese_synth.dict cannot be opened (fail-closed).
func (v PortugueseVariant) CreateDefaultSynthesizer() *ptsynth.PortugueseSynthesizer {
	return ptsynth.Instance()
}

// GetPriorityForId ports Portuguese / PortugalPortuguese.getPriorityForId.
// pt-PT (and bare "pt") use PortugalPortuguese map then base; other locales use base only.
func (v PortugueseVariant) GetPriorityForId(id string) int {
	return PortuguesePriorityForIdForCode(v.ShortCode)(id)
}

// GetRelevantLanguageModelRuleIDs ports Portuguese.getRelevantLanguageModelRules.
func (v PortugueseVariant) GetRelevantLanguageModelRuleIDs() []string {
	return PortugueseLanguageModelRelevantRuleIDs()
}
