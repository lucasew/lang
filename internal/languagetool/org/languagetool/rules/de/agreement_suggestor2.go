package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// AgreementSuggestor2 ports rules.de.AgreementSuggestor2 (simplified).
// Generates alternative determiner/adjective/noun forms via synthesizer for
// case/number/gender combinations.
type AgreementSuggestor2 struct {
	Synth      synthesis.Synthesizer
	Determiner *languagetool.AnalyzedTokenReadings
	Adj1       *languagetool.AnalyzedTokenReadings
	Adj2       *languagetool.AnalyzedTokenReadings
	Noun       *languagetool.AnalyzedTokenReadings
	// SkipSuggestions filters generated forms.
	SkipSuggestions map[string]struct{}
}

var defaultSkipSuggestions = map[string]struct{}{
	"unsren": {}, "unsrem": {}, "unsres": {}, "unsre": {}, "unsern": {}, "unserm": {}, "unsrer": {},
}

func NewAgreementSuggestor2(synth synthesis.Synthesizer, det, noun *languagetool.AnalyzedTokenReadings) *AgreementSuggestor2 {
	return &AgreementSuggestor2{
		Synth:           synth,
		Determiner:      det,
		Noun:            noun,
		SkipSuggestions: defaultSkipSuggestions,
	}
}

// WithAdjectives sets optional adjective tokens.
func (s *AgreementSuggestor2) WithAdjectives(adj1, adj2 *languagetool.AnalyzedTokenReadings) *AgreementSuggestor2 {
	s.Adj1, s.Adj2 = adj1, adj2
	return s
}

// GetSuggestions returns suggested full phrases (det + adj* + noun) with unified agreement.
func (s *AgreementSuggestor2) GetSuggestions() []string {
	if s == nil || s.Noun == nil {
		return nil
	}
	cases := []string{"NOM", "AKK", "DAT", "GEN"}
	nums := []string{"SIN", "PLU"}
	gens := []string{"MAS", "FEM", "NEU"}
	var out []string
	seen := map[string]struct{}{}
	nounLemma := lemmaOf(s.Noun)
	for _, c := range cases {
		for _, n := range nums {
			for _, g := range gens {
				nounTag := "SUB:" + c + ":" + n + ":" + g
				nounForms := s.synth(nounLemma, nounTag)
				if len(nounForms) == 0 {
					continue
				}
				detForms := []string{""}
				if s.Determiner != nil {
					detForms = s.synth(lemmaOf(s.Determiner), "ART:DEF:"+c+":"+n+":"+g)
					if len(detForms) == 0 {
						detForms = s.synth(lemmaOf(s.Determiner), "ART:IND:"+c+":"+n+":"+g)
					}
					if len(detForms) == 0 {
						detForms = []string{s.Determiner.GetToken()}
					}
				}
				adjForms := []string{""}
				if s.Adj1 != nil {
					adjForms = s.synth(lemmaOf(s.Adj1), "ADJ:"+c+":"+n+":"+g+":GRU:DEF")
					if len(adjForms) == 0 {
						adjForms = []string{s.Adj1.GetToken()}
					}
				}
				for _, det := range detForms {
					for _, adj := range adjForms {
						for _, noun := range nounForms {
							if s.skip(det) || s.skip(adj) || s.skip(noun) {
								continue
							}
							phrase := joinNonEmpty(det, adj, noun)
							if phrase == "" {
								continue
							}
							if _, ok := seen[phrase]; ok {
								continue
							}
							seen[phrase] = struct{}{}
							out = append(out, phrase)
						}
					}
				}
			}
		}
	}
	return out
}

func (s *AgreementSuggestor2) synth(lemma, tag string) []string {
	if s.Synth == nil || lemma == "" {
		return nil
	}
	tok := languagetool.NewAnalyzedToken(lemma, nil, &lemma)
	forms, err := s.Synth.Synthesize(tok, tag)
	if err != nil {
		return nil
	}
	return forms
}

func (s *AgreementSuggestor2) skip(form string) bool {
	if form == "" {
		return false
	}
	_, ok := s.SkipSuggestions[strings.ToLower(form)]
	return ok
}

func lemmaOf(r *languagetool.AnalyzedTokenReadings) string {
	if r == nil {
		return ""
	}
	for _, t := range r.GetReadings() {
		if t != nil && t.GetLemma() != nil && *t.GetLemma() != "" {
			return *t.GetLemma()
		}
	}
	return r.GetToken()
}

func joinNonEmpty(parts ...string) string {
	var b []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			b = append(b, p)
		}
	}
	return strings.Join(b, " ")
}
