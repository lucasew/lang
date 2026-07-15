// Package pipeline defines the LanguageTool analysis chain stages.
//
// Faithful order (SPEC §3.2):
//
//	text → sentence split → tokenize → tag → disambiguate → rules → filters → suggestions
//
// Initial implementation wires tokenize + rules for core text-level rules that do not
// yet require the full tagger/disambiguator. Stages remain explicit so they can be
// filled without redesign.
package pipeline

// Stage names for logging and doctor output.
const (
	StageSentenceSplit = "sentence_split"
	StageTokenize      = "tokenize"
	StageTag           = "tag"
	StageDisambiguate  = "disambiguate"
	StageRules         = "rules"
	StageFilters       = "filters"
	StageSuggestions   = "suggestions"
)

// AllStages is the full LT pipeline order.
var AllStages = []string{
	StageSentenceSplit,
	StageTokenize,
	StageTag,
	StageDisambiguate,
	StageRules,
	StageFilters,
	StageSuggestions,
}

// Sentence is a span of tokens (subset of LT AnalyzedSentence).
type Sentence struct {
	Text   string
	Start  int // rune offset of sentence in original text
	Tokens []Token
}
