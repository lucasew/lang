package synthesis

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// SpellNumber tags used by BaseSynthesizer.
const (
	SpellNumberTag         = "_spell_number_"
	SpellNumberFeminineTag = "_spell_number_:feminine"
	SpellNumberRomanTag    = "_spell_number_:Roman"
)

// BaseSynthesizer ports the non-Morfologik surface of
// org.languagetool.synthesis.BaseSynthesizer — ManualSynthesizer-backed forms.
type BaseSynthesizer struct {
	LangShortCode    string
	ResourceFileName string
	TagFileName      string
	Manual           *ManualSynthesizer
	Removal          *ManualSynthesizer
	// Lookup is optional binary-dict synthesis (lemma+pos → forms).
	Lookup func(lemma, posTag string) []string
	// PossibleTags lists known POS tags when loaded.
	PossibleTags []string
}

func NewBaseSynthesizer(langShortCode string, manual *ManualSynthesizer) *BaseSynthesizer {
	return &BaseSynthesizer{LangShortCode: langShortCode, Manual: manual}
}

// Synthesize ports BaseSynthesizer.synthesize for exact POS tags.
func (s *BaseSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.SynthesizeRE(token, posTag, false)
}

// SynthesizeRE ports synthesize with optional POS regexp.
func (s *BaseSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error) {
	if token == nil {
		return nil, nil
	}
	lemma := ""
	if token.GetLemma() != nil {
		lemma = *token.GetLemma()
	}
	if lemma == "" {
		lemma = token.GetToken()
	}
	if posTagRegExp {
		re, err := regexp.Compile("^(?:" + posTag + ")$")
		if err != nil {
			return nil, err
		}
		return s.SynthesizeForPosTags(lemma, re.MatchString), nil
	}
	// Exact POS: look up that tag directly (not filtered through possibleTags).
	return collectForms(s, lemma, []string{posTag}), nil
}

// SynthesizeForPosTags ports BaseSynthesizer.synthesizeForPosTags (Java ≥5.3):
// all forms for lemma where acceptTag returns true for the POS tag.
// Used by SpellingData ß→ss expansion (accept all tags) and LineExpander (VER:*).
func (s *BaseSynthesizer) SynthesizeForPosTags(lemma string, acceptTag func(string) bool) []string {
	if s == nil || lemma == "" || acceptTag == nil {
		return nil
	}
	var tags []string
	for _, tag := range s.allTags() {
		if acceptTag(tag) {
			tags = append(tags, tag)
		}
	}
	return collectForms(s, lemma, tags)
}

func collectForms(s *BaseSynthesizer, lemma string, tags []string) []string {
	if s == nil || lemma == "" {
		return nil
	}
	var forms []string
	seen := map[string]struct{}{}
	for _, tag := range tags {
		for _, f := range s.lookupForms(lemma, tag) {
			if s.isRemoved(lemma, tag, f) {
				continue
			}
			if _, ok := seen[f]; ok {
				continue
			}
			seen[f] = struct{}{}
			forms = append(forms, f)
		}
	}
	return forms
}

func (s *BaseSynthesizer) lookupForms(lemma, posTag string) []string {
	var out []string
	if s.Lookup != nil {
		out = append(out, s.Lookup(lemma, posTag)...)
	}
	if s.Manual != nil {
		if v := s.Manual.Lookup(lemma, posTag); len(v) > 0 {
			out = append(out, v...)
		}
	}
	return out
}

func (s *BaseSynthesizer) isRemoved(lemma, posTag, form string) bool {
	if s.Removal == nil {
		return false
	}
	for _, f := range s.Removal.Lookup(lemma, posTag) {
		if f == form {
			return true
		}
	}
	return false
}

func (s *BaseSynthesizer) allTags() []string {
	if len(s.PossibleTags) > 0 {
		return s.PossibleTags
	}
	if s.Manual != nil {
		var tags []string
		for t := range s.Manual.GetPossibleTags() {
			tags = append(tags, t)
		}
		return tags
	}
	return nil
}

// GetTargetPosTag is a stub for language-specific POS selection.
func (s *BaseSynthesizer) GetTargetPosTag(posTags []string, posTag string) string {
	if len(posTags) == 0 {
		return posTag
	}
	return posTags[0]
}

// GetPosTagCorrection ports BaseSynthesizer.getPosTagCorrection (identity).
// Polish/Arabic override when setpos synthesizes regexp-rewritten tags.
func (s *BaseSynthesizer) GetPosTagCorrection(posTag string) string {
	return posTag
}

var _ Synthesizer = (*BaseSynthesizer)(nil)
