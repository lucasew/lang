package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestStrictPOS_OF_VBN_JJ(t *testing.T) {
	// of + untagged + JJ should NOT match with StrictPOS (VBN requires real POS)
	of := patterns.NewPatternToken("of", false, false, false)
	vbn := patterns.NewPatternToken("", false, false, false)
	vbn.SetPosToken(patterns.PosToken{PosTag: "VBN", Regexp: false})
	jj := patterns.NewPatternToken("", false, false, false)
	jj.SetPosToken(patterns.PosToken{PosTag: "JJ", Regexp: false})
	rule := NewDisambiguationPatternRule("OF_VBN_JJ", "t", "en",
		[]*patterns.PatternToken{of, vbn, jj}, "JJ", nil, ActionReplace)

	// Build sentence: of / America1s(untagged) / real(JJ)
	jjTag := "JJ"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &[]string{languagetool.SentenceStartTagName}[0], nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("of", strp("IN"), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("America1s", nil, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("real", &jjTag, nil)),
	}
	// set start positions
	pos := 0
	for _, t := range toks {
		t.SetStartPos(pos)
		pos += len(t.GetToken()) + 1
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	out := rule.Replace(sent)
	// America1s should remain untagged
	var tags []string
	for _, r := range out.GetTokensWithoutWhitespace()[1].GetReadings() { // America1s is index 1 without WS? 0=of,1=America1s,2=real
		_ = r
	}
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() == "America1s" {
			for _, r := range tok.GetReadings() {
				if r != nil && r.GetPOSTag() != nil {
					tags = append(tags, *r.GetPOSTag())
				}
			}
		}
	}
	t.Logf("America1s tags after replace: %v", tags)
	require.NotContains(t, tags, "JJ")
}

func strp(s string) *string { return &s }
