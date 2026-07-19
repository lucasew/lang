package patterns

import (
	"fmt"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// CompileFromReference ports PatternToken.compile / doCompile for setpos and surface refs.
// Returns a new PatternToken configured from the referenced analysis token.
func (p *PatternToken) CompileFromReference(ref *languagetool.AnalyzedTokenReadings, synth synthesis.Synthesizer) *PatternToken {
	if p == nil {
		return nil
	}
	cp := clonePatternToken(p)
	if p.TokenMatch == nil || ref == nil {
		return cp
	}
	tm := p.TokenMatch
	ms := NewMatchStateWithSynth(tm, synth)
	ms.SetToken(ref)
	refMarker := fmt.Sprintf("\\%d", tm.GetTokenRef())
	if tm.SetsPos() {
		posReference := ms.GetTargetPosTag()
		if posReference != "" {
			neg := false
			if cp.Pos != nil {
				neg = cp.Pos.Negate
			}
			// Java setPosToken(new PosToken(posReference, tokenReference.posRegExp(), getNegation()))
			// Uses Match.posRegExp (postag_regexp on match), not always true for .* patterns.
			cp.SetPosToken(PosToken{
				PosTag: posReference,
				Regexp: tm.IsPostagRegexp() || looksLikePosRegexp(posReference),
				Negate: neg,
			})
		}
		if cp.Token != "" {
			cp.Token = strings.ReplaceAll(cp.Token, refMarker, "")
		}
	} else {
		forms := ms.ToFinalString("")
		repl := ""
		if len(forms) > 0 {
			repl = forms[0]
		}
		if cp.Token != "" {
			cp.Token = strings.ReplaceAll(cp.Token, refMarker, repl)
		} else {
			cp.Token = repl
		}
	}
	// Compiled token is no longer a live reference for further resolveReference.
	cp.TokenMatch = nil
	return cp
}

func looksLikePosRegexp(s string) bool {
	return strings.ContainsAny(s, ".*+?|[](){}")
}
