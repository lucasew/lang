package rules

// SuggestionType ports org.languagetool.rules.SuggestedReplacement.SuggestionType.
type SuggestionType string

const (
	SuggestionTypeDefault     SuggestionType = "Default"
	SuggestionTypeTranslation SuggestionType = "Translation"
	SuggestionTypeCurated     SuggestionType = "Curated"
)

// SpellingHighConfidence matches SpellingCheckRule.HIGH_CONFIDENCE (0.99).
const SpellingHighConfidence float32 = 0.99

// SuggestedReplacement ports org.languagetool.rules.SuggestedReplacement.
type SuggestedReplacement struct {
	Replacement      string
	ShortDescription *string
	Suffix           *string
	Features         map[string]float32 // sorted order not preserved; OK for Go surface
	Confidence       *float32
	Type             SuggestionType
	Weight           *int
}

func NewSuggestedReplacement(replacement string) *SuggestedReplacement {
	return NewSuggestedReplacementFull(replacement, nil, nil)
}

func NewSuggestedReplacementWithDesc(replacement string, shortDescription *string) *SuggestedReplacement {
	return NewSuggestedReplacementFull(replacement, shortDescription, nil)
}

func NewSuggestedReplacementFull(replacement string, shortDescription, suffix *string) *SuggestedReplacement {
	if replacement == "" {
		// Java requireNonNull only; empty string is allowed.
	}
	return &SuggestedReplacement{
		Replacement:      replacement,
		ShortDescription: shortDescription,
		Suffix:           suffix,
		Features:         map[string]float32{},
		Type:             SuggestionTypeDefault,
	}
}

func CopySuggestedReplacement(src *SuggestedReplacement) *SuggestedReplacement {
	if src == nil {
		return nil
	}
	out := NewSuggestedReplacementFull(src.Replacement, src.ShortDescription, src.Suffix)
	out.Confidence = src.Confidence
	out.Type = src.Type
	out.Weight = src.Weight
	if len(src.Features) > 0 {
		out.Features = make(map[string]float32, len(src.Features))
		for k, v := range src.Features {
			out.Features[k] = v
		}
	}
	return out
}

func (s *SuggestedReplacement) GetReplacement() string { return s.Replacement }
func (s *SuggestedReplacement) SetReplacement(r string) {
	s.Replacement = r
}

func (s *SuggestedReplacement) GetShortDescription() *string { return s.ShortDescription }
func (s *SuggestedReplacement) SetShortDescription(desc *string) {
	s.ShortDescription = desc
}

func (s *SuggestedReplacement) GetType() SuggestionType {
	if s.Type == "" {
		return SuggestionTypeDefault
	}
	return s.Type
}
func (s *SuggestedReplacement) SetType(t SuggestionType) { s.Type = t }

func (s *SuggestedReplacement) GetSuffix() *string  { return s.Suffix }
func (s *SuggestedReplacement) SetSuffix(v *string) { s.Suffix = v }

func (s *SuggestedReplacement) GetConfidence() *float32  { return s.Confidence }
func (s *SuggestedReplacement) SetConfidence(c *float32) { s.Confidence = c }

func (s *SuggestedReplacement) GetFeatures() map[string]float32 { return s.Features }
func (s *SuggestedReplacement) SetFeatures(f map[string]float32) {
	if f == nil {
		s.Features = map[string]float32{}
		return
	}
	s.Features = f
}

func (s *SuggestedReplacement) GetWeight() *int  { return s.Weight }
func (s *SuggestedReplacement) SetWeight(w *int) { s.Weight = w }

func (s *SuggestedReplacement) String() string {
	// Java: replacement + '(' + shortDescription + ')' — null desc prints as "null"
	desc := "null"
	if s != nil && s.ShortDescription != nil {
		desc = *s.ShortDescription
	}
	repl := ""
	if s != nil {
		repl = s.Replacement
	}
	return repl + "(" + desc + ")"
}

// ConvertSuggestions maps bare strings to SuggestedReplacement list.
func ConvertSuggestions(suggestions []string) []*SuggestedReplacement {
	out := make([]*SuggestedReplacement, len(suggestions))
	for i, s := range suggestions {
		out[i] = NewSuggestedReplacement(s)
	}
	return out
}

// TopMatch builds a single high-confidence suggestion.
func TopMatch(word string, shortDesc *string) []*SuggestedReplacement {
	s := NewSuggestedReplacementWithDesc(word, shortDesc)
	c := SpellingHighConfidence
	s.SetConfidence(&c)
	return []*SuggestedReplacement{s}
}
