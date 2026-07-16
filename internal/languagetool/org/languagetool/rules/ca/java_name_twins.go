package ca

// CatalanUnpairedExclamationMarksRule is the Java-name twin (¡…! unpaired marks).
// Construct with NewCatalanUnpairedExclamationMarksRule.
type CatalanUnpairedExclamationMarksRule = CatalanUnpairedQuestionMarksRule

// CatalanMorfologikMultitokenSpeller is the Java-name twin for multitoken speller access.
type CatalanMorfologikMultitokenSpeller struct {
	// Factory optional; when nil, GetSuggestions returns nil.
	Factory MultitokenSpellerFactory
}

func NewCatalanMorfologikMultitokenSpeller(factory MultitokenSpellerFactory) *CatalanMorfologikMultitokenSpeller {
	return &CatalanMorfologikMultitokenSpeller{Factory: factory}
}

func (s *CatalanMorfologikMultitokenSpeller) GetSuggestions(word string) []string {
	if s == nil {
		return nil
	}
	return GetCatalanMultitokenSpellerSuggestions(s.Factory, word)
}

// PronomsFeblesHelper is the Java-name twin for weak-pronoun transforms.
type PronomsFeblesHelper struct{}

func (PronomsFeblesHelper) Transform(input string, pos PronounPosition) string {
	return Transform(input, pos)
}
func (PronomsFeblesHelper) TransformDavant(input, next string) string {
	return TransformDavant(input, next)
}
func (PronomsFeblesHelper) TransformDarrere(input, prev string) string {
	return TransformDarrere(input, prev)
}
func (PronomsFeblesHelper) GetReflexivePronoun(key string) string { return GetReflexivePronoun(key) }
func (PronomsFeblesHelper) GetDativePronoun(key string) string    { return GetDativePronoun(key) }

// NounToVerbHelper is the Java-name twin for noun→verb lemma map.
type NounToVerbHelper struct{}

func (NounToVerbHelper) VerbForNoun(noun string) (string, bool) {
	v := NounToVerb(noun)
	return v, v != ""
}

// VerbsHelper is the Java-name twin for dicendi verb checks.
type VerbsHelper struct{}

func (VerbsHelper) IsVerbDicendi(lemma string) bool { return IsVerbDicendi(lemma) }

// CatalanRemoteRewriteHelper is the Java-name twin for remote rewrite config.
type CatalanRemoteRewriteHelper struct {
	Config CatalanRemoteRewriteConfig
}

func NewCatalanRemoteRewriteHelper() *CatalanRemoteRewriteHelper {
	return &CatalanRemoteRewriteHelper{Config: DefaultRemoteRewriteConfig()}
}

func (h *CatalanRemoteRewriteHelper) IsAvailable() bool {
	return h.Config.IsRemoteServiceAvailable()
}

// ApostophationHelper is the Java-name twin for preposition+determiner apostrophation.
type ApostophationHelper struct{}

func (ApostophationHelper) GetPrepositionAndDeterminer(newForm, genderNumber, preposition string) string {
	return GetPrepositionAndDeterminer(newForm, genderNumber, preposition)
}
