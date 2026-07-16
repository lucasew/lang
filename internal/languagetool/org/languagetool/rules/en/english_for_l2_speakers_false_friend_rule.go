package en

// EnglishForL2SpeakersFalseFriendRule ports metadata for
// org.languagetool.rules.en.EnglishForL2SpeakersFalseFriendRule variants.
// Full ngram ConfusionProbabilityRule matching is deferred.
type EnglishForL2SpeakersFalseFriendRule struct {
	ID           string
	MotherTongue string // short code, e.g. "de"
	Language     string // target language, e.g. "en"
	// Filenames are confusion set resources under the language resource dir.
	Filenames []string
	// ExampleWrong / ExampleFixed surface for documentation / tests.
	ExampleWrong string
	ExampleFixed string
}

func (r *EnglishForL2SpeakersFalseFriendRule) GetID() string { return r.ID }
func (r *EnglishForL2SpeakersFalseFriendRule) GetFilenames() []string {
	return append([]string(nil), r.Filenames...)
}

// NewEnglishForGermansFalseFriendRule ports EnglishForGermansFalseFriendRule metadata.
func NewEnglishForGermansFalseFriendRule() *EnglishForL2SpeakersFalseFriendRule {
	return &EnglishForL2SpeakersFalseFriendRule{
		ID:           "EN_FOR_DE_SPEAKERS_FALSE_FRIENDS",
		MotherTongue: "de",
		Language:     "en",
		Filenames:    []string{"confusion_sets_l2_de.txt"},
		ExampleWrong: "My handy is broken.",
		ExampleFixed: "My phone is broken.",
	}
}

// NewEnglishForFrenchFalseFriendRule ports EnglishForFrenchFalseFriendRule.
func NewEnglishForFrenchFalseFriendRule() *EnglishForL2SpeakersFalseFriendRule {
	return &EnglishForL2SpeakersFalseFriendRule{
		ID:           "EN_FOR_FR_SPEAKERS_FALSE_FRIENDS",
		MotherTongue: "fr",
		Language:     "en",
		Filenames:    []string{"confusion_sets_l2_fr.txt"},
	}
}

// NewEnglishForSpaniardsFalseFriendRule ports EnglishForSpaniardsFalseFriendRule.
func NewEnglishForSpaniardsFalseFriendRule() *EnglishForL2SpeakersFalseFriendRule {
	return &EnglishForL2SpeakersFalseFriendRule{
		ID:           "EN_FOR_ES_SPEAKERS_FALSE_FRIENDS",
		MotherTongue: "es",
		Language:     "en",
		Filenames:    []string{"confusion_sets_l2_es.txt"},
	}
}

// NewEnglishForDutchmenFalseFriendRule ports EnglishForDutchmenFalseFriendRule.
func NewEnglishForDutchmenFalseFriendRule() *EnglishForL2SpeakersFalseFriendRule {
	return &EnglishForL2SpeakersFalseFriendRule{
		ID:           "EN_FOR_NL_SPEAKERS_FALSE_FRIENDS",
		MotherTongue: "nl",
		Language:     "en",
		Filenames:    []string{"confusion_sets_l2_nl.txt"},
	}
}
