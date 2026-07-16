package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// EnNoInfinitiuSuggestionFilter ports EN_NO_INFINITIU suggestion prefixes.
type EnNoInfinitiuSuggestionFilter struct {
	// Synth synthesizes the infinitive verb with a full postag.
	Synth func(lemma, postag string) string
}

func NewEnNoInfinitiuSuggestionFilter() *EnNoInfinitiuSuggestionFilter {
	return &EnNoInfinitiuSuggestionFilter{}
}

// EnNoInfinitiuInput describes tense/person context around "en no + infinitive".
type EnNoInfinitiuInput struct {
	// TempsVerbal is the neighbouring finite verb postag (e.g. VMIP3S00).
	TempsVerbal string
	// PassatPerifrastic forces imperfective past (VMII) even for present tags.
	PassatPerifrastic bool
	// VerbBefore means the finite verb is before the infinitive → "perquè no"; else "com que no".
	VerbBefore bool
	Lemma      string
	// PronounsAfter weak pronouns after the infinitive.
	PronounsAfter string
	CasingModel   string
}

// Suggest builds "com que no …" / "perquè no …" finite rewrites.
func (f *EnNoInfinitiuSuggestionFilter) Suggest(in EnNoInfinitiuInput) []string {
	if f.Synth == nil || len(in.TempsVerbal) < 6 {
		return nil
	}
	prefix := "VMII"
	// present IP/IF and not periphrastic past → VMIP
	moodTense := in.TempsVerbal[2:4]
	if (moodTense == "IP" || moodTense == "IF") && !in.PassatPerifrastic {
		prefix = "VMIP"
	}
	var synthVerbs []string
	// always offer 3S if original person is not 3S
	personNumber := in.TempsVerbal[4:6]
	if personNumber != "3S" {
		if s := f.Synth(in.Lemma, prefix+"3S"+in.TempsVerbal[6:]); s != "" {
			synthVerbs = append(synthVerbs, s)
		}
	}
	if s := f.Synth(in.Lemma, prefix+in.TempsVerbal[4:]); s != "" {
		synthVerbs = append(synthVerbs, s)
	}
	intro := "com que no "
	if in.VerbBefore {
		intro = "perquè no "
	}
	var out []string
	seen := map[string]struct{}{}
	for _, v := range synthVerbs {
		var b strings.Builder
		b.WriteString(intro)
		if in.PronounsAfter != "" {
			b.WriteString(TransformDavant(in.PronounsAfter, v))
		}
		b.WriteString(v)
		s := b.String()
		if in.CasingModel != "" {
			s = tools.PreserveCase(s, in.CasingModel)
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
