package synthesis

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// Synthesizer ports org.languagetool.synthesis.Synthesizer.
type Synthesizer interface {
	// Synthesize generates forms for token with the given POS tag.
	Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error)
	// SynthesizeRE generates forms; posTag is a regex when posTagRegExp is true.
	SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error)
}

// FuncSynthesizer adapts functions to Synthesizer.
type FuncSynthesizer struct {
	Synth   func(token *languagetool.AnalyzedToken, posTag string) ([]string, error)
	SynthRE func(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error)
}

func (f FuncSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if f.Synth == nil {
		return nil, nil
	}
	return f.Synth(token, posTag)
}

func (f FuncSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error) {
	if f.SynthRE != nil {
		return f.SynthRE(token, posTag, posTagRegExp)
	}
	if !posTagRegExp {
		return f.Synthesize(token, posTag)
	}
	return f.Synthesize(token, posTag)
}
