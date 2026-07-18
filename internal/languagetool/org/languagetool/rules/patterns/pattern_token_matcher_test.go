package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPatternTokenMatcher(t *testing.T) {
	m := NewPatternTokenMatcher(Token("Hello"))
	tok := languagetool.NewAnalyzedToken("hello", nil, nil)
	require.True(t, m.IsMatched(tok))
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("world", nil, nil)))

	cs := NewPatternTokenMatcher(CsToken("Hello"))
	require.False(t, cs.IsMatched(tok))
	require.True(t, cs.IsMatched(languagetool.NewAnalyzedToken("Hello", nil, nil)))

	re := NewPatternTokenMatcher(TokenRegex("c.t"))
	require.True(t, re.IsMatched(languagetool.NewAnalyzedToken("cat", nil, nil)))

	pos := "NN"
	pm := NewPatternTokenMatcher(Pos("NN"))
	require.True(t, pm.IsMatched(languagetool.NewAnalyzedToken("dog", &pos, nil)))
}

// Soft path: upstream goldens only (RU kurica, RU software adj, TA polum).
func TestSoftInflectedAndSurfacePOS(t *testing.T) {
	// яйцо / яйца — short shared stem (min 3)
	require.True(t, softSharedStemMatch("яйца", "яйцо"))
	require.True(t, softInflectedSurfaceMatch("высиживает", "высиживать", false))

	// RE alternatives with inflected="yes" (Adj_NN_number_Software)
	pt := NewPatternToken("программный|аппаратный", false, true, true)
	m := NewPatternTokenMatcher(pt)
	require.True(t, m.IsMatched(languagetool.NewAnalyzedToken("программных", nil, nil)))

	// Surface "." with postag SENT_END (TA polum second token)
	sentEnd := NewPatternToken(".", false, false, false)
	sentEnd.Pos = &PosToken{PosTag: "SENT_END"}
	sm := NewPatternTokenMatcher(sentEnd)
	require.True(t, sm.IsMatched(languagetool.NewAnalyzedToken(".", nil, nil)))

	// Surface RE + word POS without tagger (TL ADJECTIVE-V_COMMON_NOUN)
	adj := NewPatternToken(".*[aeiou]", false, true, false)
	adj.Pos = &PosToken{PosTag: "(ADMO|ADCO).*", Regexp: true}
	am := NewPatternTokenMatcher(adj)
	require.True(t, am.IsMatched(languagetool.NewAnalyzedToken("mababa", nil, nil)))
	require.False(t, am.IsMatched(languagetool.NewAnalyzedToken("madasalin", nil, nil)))
}

func TestSoftRegexpAlternatives(t *testing.T) {
	require.Equal(t, []string{"a", "b", "c"}, softRegexpAlternatives("a|b|c"))
	require.Equal(t, []string{"foo", "bar"}, softRegexpAlternatives("(?:foo|bar)"))
}

func TestSoftIrregularLemma(t *testing.T) {
	require.True(t, softInflectedSurfaceMatch("was", "be", false))
	require.True(t, softInflectedSurfaceMatch("est", "être", false))
	require.True(t, softInflectedSurfaceMatch("va", "dir", false)) // AST
	require.True(t, softInflectedSurfaceMatch("va", "ir", false))  // ES
	require.True(t, softInflectedSurfaceMatch("va", "aller", false))
	require.True(t, softInflectedSurfaceMatch("ist", "sein", false))
	require.False(t, softInflectedSurfaceMatch("va", "be", false))
}

func TestSoftGermanGeParticiple(t *testing.T) {
	require.True(t, softGermanGeParticiple("gemacht", "machen"))
	require.True(t, softGermanGeParticiple("gelernt", "lernen"))
	require.True(t, softInflectedSurfaceMatch("gemacht", "machen", false))
	require.True(t, softInflectedSurfaceMatch("genommen", "nehmen", false)) // irregular map
	require.True(t, softInflectedSurfaceMatch("ging", "gehen", false) || softGermanGeParticiple("gegangen", "gehen"))
	require.True(t, softInflectedSurfaceMatch("gegangen", "gehen", false) || softGermanGeParticiple("gegangen", "gehen"))
	// separable: ausgelost (typo form) ← auslosen (ALARM_AUSLOSEN soft golden)
	require.True(t, softGermanGeParticiple("ausgelost", "auslosen"))
	require.True(t, softInflectedSurfaceMatch("ausgelost", "auslosen", false))
}

func TestSoftFrenchErInflected(t *testing.T) {
	require.True(t, softFrenchErInflected("placé", "placer"))
	require.True(t, softFrenchErInflected("places", "placer"))
	require.True(t, softFrenchErInflected("rencontré", "rencontrer"))
	require.True(t, softInflectedSurfaceMatch("mis", "mettre", false))
	require.False(t, softFrenchErInflected("chat", "placer"))
}

func TestSoftClosedClassPOS(t *testing.T) {
	// DT_PRP: empty PRP$ must not soft-match nouns.
	prp := Pos("PRP$?")
	prp.Pos.Regexp = true
	m := NewPatternTokenMatcher(prp)
	require.True(t, m.IsMatched(languagetool.NewAnalyzedToken("my", nil, nil)))
	require.True(t, m.IsMatched(languagetool.NewAnalyzedToken("you", nil, nil)))
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("man", nil, nil)))
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("search", nil, nil)))

	// Open-class empty POS still soft-matches words (NN-like).
	nn := Pos("NN.*")
	nn.Pos.Regexp = true
	nm := NewPatternTokenMatcher(nn)
	require.True(t, nm.IsMatched(languagetool.NewAnalyzedToken("man", nil, nil)))
}

// Upstream EN NON_ENGLISH_CHARACTER_IN_A_WORD uses Java \uXXXX escapes.
func TestSoftNormalizeJavaRegexpUnicode(t *testing.T) {
	pat := `[a-z]*(\u043E|\u0455|\u0435|\u0440|\u03BF)[a-z]*`
	got := softNormalizeJavaRegexp(pat)
	require.Contains(t, got, `\x{043e}`)
	require.Contains(t, got, `\x{0455}`)
	m := NewPatternTokenMatcher(NewPatternToken(pat, false, true, false))
	// U+0455 CYRILLIC SMALL LETTER DZE looks like Latin s
	require.True(t, m.IsMatched(languagetool.NewAnalyzedToken("ѕee", nil, nil)))
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("see", nil, nil)))
}

func TestIsMatchedReadings_ChunkTag(t *testing.T) {
	// Java chunk="B-NP" must match token chunk tags (AND with surface).
	pt := NewPatternToken("house", false, false, false)
	pt.SetChunkTag("B-NP", false)
	m := NewPatternTokenMatcher(pt)
	nn := "NN"
	atr := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("house", &nn, nil))
	require.False(t, m.IsMatchedReadings(atr), "no chunk tag yet")
	atr.SetChunkTags([]string{"B-NP"})
	require.True(t, m.IsMatchedReadings(atr))
	atr.SetChunkTags([]string{"I-NP"})
	require.False(t, m.IsMatchedReadings(atr))

	// chunk_re
	pt2 := NewPatternToken("", false, false, false)
	pt2.SetChunkTag(".-NP", true)
	m2 := NewPatternTokenMatcher(pt2)
	atr2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("x", &nn, nil))
	atr2.SetChunkTags([]string{"B-NP"})
	require.True(t, m2.IsMatchedReadings(atr2))
}

func TestIsMatchedReadings_AndGroupAcrossReadings(t *testing.T) {
	// Java <and><token postag="VBP"/><token postag="NN:UN"/></and>
	// matches a token that has both readings (not one reading with both tags).
	base := NewPatternToken("", false, false, false)
	base.SetPosToken(PosToken{PosTag: "VBP", Regexp: false})
	andNN := NewPatternToken("", false, false, false)
	andNN.SetPosToken(PosToken{PosTag: "NN:UN", Regexp: false})
	base.AddAndGroupElement(andNN)
	m := NewPatternTokenMatcher(base)

	vbp, nnun, nn := "VBP", "NN:UN", "NN"
	// only VBP — fail
	onlyVB := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("fall", &vbp, nil))
	require.False(t, m.IsMatchedReadings(onlyVB))
	// VBP + NN:UN — pass
	both := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("fall", &vbp, nil))
	both.AddReading(languagetool.NewAnalyzedToken("fall", &nnun, nil), "dict")
	require.True(t, m.IsMatchedReadings(both))
	// VBP + NN — fail and-group
	vbnn := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("fall", &vbp, nil))
	vbnn.AddReading(languagetool.NewAnalyzedToken("fall", &nn, nil), "dict")
	require.False(t, m.IsMatchedReadings(vbnn))
}

func TestPreviousException_BlocksMatch(t *testing.T) {
	// Java: token "mine" with exception scope=previous "not"
	mine := NewPatternToken("mine", false, false, false)
	mine.SetPreviousException("not", false, false)
	rule := NewAbstractTokenBasedRule("T", "t", "en", []*PatternToken{mine})
	m := NewPatternRuleMatcher(rule)

	nn := "NN"
	// "is mine" — should match
	toksOK := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("is", strPtr("VBZ"), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("mine", &nn, nil)),
	}
	pos := 0
	for _, atr := range toksOK {
		atr.SetStartPos(pos)
		pos += len(atr.GetToken()) + 1
	}
	ms, err := m.Match(languagetool.NewAnalyzedSentence(toksOK))
	require.NoError(t, err)
	require.NotEmpty(t, ms)

	// "not mine" — previous exception blocks
	toksBlock := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("not", strPtr("RB"), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("mine", &nn, nil)),
	}
	pos = 0
	for _, atr := range toksBlock {
		atr.SetStartPos(pos)
		pos += len(atr.GetToken()) + 1
	}
	ms2, err := m.Match(languagetool.NewAnalyzedSentence(toksBlock))
	require.NoError(t, err)
	require.Empty(t, ms2)
}

func TestNextException_BlocksMatch(t *testing.T) {
	// Java: can with exception scope=next be|do|not
	can := NewPatternToken("can", false, false, false)
	can.SetNextException("be|do|not", true, false)
	rule := NewAbstractTokenBasedRule("T", "t", "en", []*PatternToken{can})
	m := NewPatternRuleMatcher(rule)
	md := "MD"
	// "can run" — match
	toksOK := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("can", &md, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("run", strPtr("VB"), nil)),
	}
	pos := 0
	for _, atr := range toksOK {
		atr.SetStartPos(pos)
		pos += len(atr.GetToken()) + 1
	}
	ms, err := m.Match(languagetool.NewAnalyzedSentence(toksOK))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	// "can be" — next exception blocks
	toksBlock := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strPtr(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("can", &md, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("be", strPtr("VB"), nil)),
	}
	pos = 0
	for _, atr := range toksBlock {
		atr.SetStartPos(pos)
		pos += len(atr.GetToken()) + 1
	}
	ms2, err := m.Match(languagetool.NewAnalyzedSentence(toksBlock))
	require.NoError(t, err)
	require.Empty(t, ms2)
}
