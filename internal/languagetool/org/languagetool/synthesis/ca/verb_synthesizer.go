package ca

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// VerbSynthesizer ports org.languagetool.synthesis.ca.VerbSynthesizer (subset).
// Locates a verb group in tokens and records target lemma/POS for synthesis.
type VerbSynthesizer struct {
	Tokens           []*languagetool.AnalyzedTokenReadings
	IFirstVerb       int
	ILastVerb        int
	NewLemma         string
	NewPostag        string
	NumPronounsBefore int
}

var (
	PVerb           = regexp.MustCompile(`^V.*`)
	PInflectedVerb  = regexp.MustCompile(`^V.[SIM].*`)
	PImperativeVerb = regexp.MustCompile(`^V.M.*`)
	PVerbIS         = regexp.MustCompile(`^V.[IS].*`)
	PNonParticiple  = regexp.MustCompile(`^V.[^P].*`)
	PParticiple     = regexp.MustCompile(`^V.P.*`)
)

func NewVerbSynthesizer(tokens []*languagetool.AnalyzedTokenReadings) *VerbSynthesizer {
	return &VerbSynthesizer{
		Tokens:     tokens,
		IFirstVerb: -1,
		ILastVerb:  -1,
		NumPronounsBefore: -1,
	}
}

// FindVerbGroup scans tokens for a contiguous V.* group (simplified).
func (v *VerbSynthesizer) FindVerbGroup() bool {
	if v == nil {
		return false
	}
	v.IFirstVerb, v.ILastVerb = -1, -1
	for i, tok := range v.Tokens {
		if tok == nil {
			continue
		}
		if hasVerbTag(tok) {
			if v.IFirstVerb < 0 {
				v.IFirstVerb = i
			}
			v.ILastVerb = i
		} else if v.IFirstVerb >= 0 {
			// stop at first non-verb after group start (unless whitespace only handled outside)
			break
		}
	}
	return v.IFirstVerb >= 0
}

func hasVerbTag(tok *languagetool.AnalyzedTokenReadings) bool {
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		if PVerb.MatchString(*r.GetPOSTag()) {
			return true
		}
	}
	return false
}

// SetTarget sets the synthesised lemma and POS for the verb group.
func (v *VerbSynthesizer) SetTarget(lemma, postag string) {
	if v == nil {
		return
	}
	v.NewLemma = lemma
	v.NewPostag = postag
}

// HasTarget reports whether lemma and postag were set.
func (v *VerbSynthesizer) HasTarget() bool {
	return v != nil && v.NewLemma != "" && v.NewPostag != ""
}
