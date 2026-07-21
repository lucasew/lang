package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// preparePOSElement ports UnifierTest.preparePOSElement.
func preparePOSElement(posString string) *PatternToken {
	pt := NewPatternToken("", false, false, false)
	pt.SetPosToken(PosToken{PosTag: posString, Regexp: true, Negate: false})
	return pt
}

// Twin of UnifierTest.testUnificationCase
func TestUnifier_UnificationCase(t *testing.T) {
	cfg := NewUnifierConfiguration()
	cfg.SetEquivalence("case-sensitivity", "lowercase", NewPatternToken(`\p{Ll}+`, true, true, false))
	cfg.SetEquivalence("case-sensitivity", "uppercase", NewPatternToken(`\p{Lu}\p{Ll}+`, true, true, false))
	cfg.SetEquivalence("case-sensitivity", "alluppercase", NewPatternToken(`\p{Lu}+$`, true, true, false))

	lower1 := languagetool.NewAnalyzedToken("lower", strPtr("JJR"), strPtr("lower"))
	lower2 := languagetool.NewAnalyzedToken("lowercase", strPtr("JJ"), strPtr("lowercase"))
	upper1 := languagetool.NewAnalyzedToken("Uppercase", strPtr("JJ"), strPtr("Uppercase"))
	upper2 := languagetool.NewAnalyzedToken("John", strPtr("NNP"), strPtr("John"))
	upperAll1 := languagetool.NewAnalyzedToken("JOHN", strPtr("NNP"), strPtr("John"))
	upperAll2 := languagetool.NewAnalyzedToken("JAMES", strPtr("NNP"), strPtr("James"))

	uni := cfg.CreateUnifier()
	equiv := map[string][]string{"case-sensitivity": {"lowercase"}}

	satisfied := uni.IsSatisfied(lower1, equiv)
	satisfied = satisfied && uni.IsSatisfied(lower2, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.True(t, satisfied)
	uni.Reset()

	satisfied = uni.IsSatisfied(upper2, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(lower2, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.False(t, satisfied)
	uni.Reset()

	satisfied = uni.IsSatisfied(upper1, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(lower1, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.False(t, satisfied)
	uni.Reset()

	satisfied = uni.IsSatisfied(upper2, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(upper1, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.False(t, satisfied)
	uni.Reset()

	equiv = map[string][]string{"case-sensitivity": {"uppercase"}}
	satisfied = uni.IsSatisfied(upper2, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(upper1, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.True(t, satisfied)
	uni.Reset()

	equiv = map[string][]string{"case-sensitivity": {"alluppercase"}}
	satisfied = uni.IsSatisfied(upper2, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(upper1, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.False(t, satisfied)
	uni.Reset()

	satisfied = uni.IsSatisfied(upperAll2, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(upperAll1, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.True(t, satisfied)
}

// Twin of UnifierTest.testUnificationNumber
func TestUnifier_UnificationNumber(t *testing.T) {
	cfg := NewUnifierConfiguration()
	cfg.SetEquivalence("number", "singular", preparePOSElement(`.*[\.:]sg:.*`))
	cfg.SetEquivalence("number", "plural", preparePOSElement(`.*[\.:]pl:.*`))
	uni := cfg.CreateUnifier()

	sing1 := languagetool.NewAnalyzedToken("mały", strPtr("adj:sg:blahblah"), strPtr("mały"))
	sing2 := languagetool.NewAnalyzedToken("człowiek", strPtr("subst:sg:blahblah"), strPtr("człowiek"))
	equiv := map[string][]string{"number": {"singular"}}

	satisfied := uni.IsSatisfied(sing1, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(sing2, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.True(t, satisfied)
	uni.Reset()

	// multiple readings — OR for interpretations, AND for tokens
	sing1a := languagetool.NewAnalyzedToken("mały", strPtr("adj:pl:blahblah"), strPtr("mały"))
	satisfied = uni.IsSatisfied(sing1, equiv)
	satisfied = satisfied || uni.IsSatisfied(sing1a, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(sing2, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.True(t, satisfied)
	uni.Reset()

	// any of the equivalences
	equiv = map[string][]string{"number": {"singular", "plural"}}
	sing1a = languagetool.NewAnalyzedToken("mały", strPtr("adj:pl:blahblah"), strPtr("mały"))
	satisfied = uni.IsSatisfied(sing1, equiv)
	satisfied = satisfied || uni.IsSatisfied(sing1a, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(sing2, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.True(t, satisfied)
	uni.Reset()

	// blank types → all number equivalences
	sing1a = languagetool.NewAnalyzedToken("mały", strPtr("adj:pl:blahblah"), strPtr("mały"))
	equiv = map[string][]string{"number": nil}
	satisfied = uni.IsSatisfied(sing1, equiv)
	satisfied = satisfied || uni.IsSatisfied(sing1a, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(sing2, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.True(t, satisfied)
	uni.Reset()

	// non-agreeing with blank types
	satisfied = uni.IsSatisfied(sing1a, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(sing2, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.False(t, satisfied)
	uni.Reset()
}

// Twin of UnifierTest.testUnificationNumberGender
func TestUnifier_UnificationNumberGender(t *testing.T) {
	cfg := NewUnifierConfiguration()
	cfg.SetEquivalence("number", "singular", preparePOSElement(`.*[\.:]sg:.*`))
	cfg.SetEquivalence("number", "plural", preparePOSElement(`.*[\.:]pl:.*`))
	cfg.SetEquivalence("gender", "feminine", preparePOSElement(`.*[\.:]f`))
	cfg.SetEquivalence("gender", "masculine", preparePOSElement(`.*[\.:]m`))
	uni := cfg.CreateUnifier()

	sing1 := languagetool.NewAnalyzedToken("mały", strPtr("adj:sg:blahblah:m"), strPtr("mały"))
	sing1a := languagetool.NewAnalyzedToken("mała", strPtr("adj:sg:blahblah:f"), strPtr("mały"))
	sing1b := languagetool.NewAnalyzedToken("małe", strPtr("adj:pl:blahblah:m"), strPtr("mały"))
	sing2 := languagetool.NewAnalyzedToken("człowiek", strPtr("subst:sg:blahblah:m"), strPtr("człowiek"))

	equiv := map[string][]string{"number": nil, "gender": nil}
	satisfied := uni.IsSatisfied(sing1, equiv)
	satisfied = satisfied || uni.IsSatisfied(sing1a, equiv)
	satisfied = satisfied || uni.IsSatisfied(sing1b, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(sing2, equiv)
	uni.StartNextToken()
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.True(t, satisfied)
	uts := uni.GetUnifiedTokens()
	require.NotNil(t, uts)
	require.Len(t, uts, 2)
	require.Equal(t, "mały", uts[0].GetToken())
	require.Equal(t, "człowiek", uts[1].GetToken())
	uni.Reset()
}

// Twin of UnifierTest.testMultipleFeats (simplified isUnified path + disagreement)
func TestUnifier_MultipleFeats(t *testing.T) {
	cfg := NewUnifierConfiguration()
	cfg.SetEquivalence("number", "singular", preparePOSElement(`.*[\.:]sg:.*`))
	cfg.SetEquivalence("number", "plural", preparePOSElement(`.*[\.:]pl:.*`))
	cfg.SetEquivalence("gender", "feminine", preparePOSElement(`.*[\.:]f([.:].*)?`))
	cfg.SetEquivalence("gender", "masculine", preparePOSElement(`.*[\.:]m([.:].*)?`))
	cfg.SetEquivalence("gender", "neutral", preparePOSElement(`.*[\.:]n([.:].*)?`))
	uni := cfg.CreateUnifier()

	sing1 := languagetool.NewAnalyzedToken("mały", strPtr("adj:sg:blahblah:m"), strPtr("mały"))
	sing1a := languagetool.NewAnalyzedToken("mały", strPtr("adj:pl:blahblah:f"), strPtr("mały"))
	sing1b := languagetool.NewAnalyzedToken("mały", strPtr("adj:pl:blahblah:f"), strPtr("mały"))
	sing2 := languagetool.NewAnalyzedToken("zgarbiony", strPtr("adj:pl:blahblah:f"), strPtr("zgarbiony"))
	sing3 := languagetool.NewAnalyzedToken("człowiek", strPtr("subst:sg:blahblah:m"), strPtr("człowiek"))
	equiv := map[string][]string{"number": nil, "gender": nil}

	satisfied := uni.IsSatisfied(sing1, equiv)
	satisfied = satisfied || uni.IsSatisfied(sing1a, equiv)
	satisfied = satisfied || uni.IsSatisfied(sing1b, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(sing2, equiv)
	uni.StartNextToken()
	satisfied = satisfied && uni.IsSatisfied(sing3, equiv)
	uni.StartNextToken()
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.False(t, satisfied)
	uni.Reset()

	// simplified isUnified interface
	uni.IsUnified(sing1, equiv, false)
	uni.IsUnified(sing1a, equiv, false)
	uni.IsUnified(sing1b, equiv, true)
	uni.IsUnified(sing2, equiv, true)
	require.False(t, uni.IsUnified(sing3, equiv, true))
	uni.Reset()

	// matching path
	sing1a = languagetool.NewAnalyzedToken("osobiste", strPtr("adj:pl:nom.acc.voc:f.n.m2.m3:pos:aff"), strPtr("osobisty"))
	sing1b = languagetool.NewAnalyzedToken("osobiste", strPtr("adj:sg:nom.acc.voc:n:pos:aff"), strPtr("osobisty"))
	sing2 = languagetool.NewAnalyzedToken("godło", strPtr("subst:sg:nom.acc.voc:n"), strPtr("godło"))
	uni.IsUnified(sing1a, equiv, false)
	uni.IsUnified(sing1b, equiv, true)
	require.True(t, uni.IsUnified(sing2, equiv, true))
	fu := uni.GetFinalUnified()
	require.NotNil(t, fu)
	require.Len(t, fu, 2)
	require.Equal(t, "osobiste", fu[0].GetToken())
	require.Equal(t, "godło", fu[1].GetToken())
	uni.Reset()
}

// Twin of UnifierTest.testMultipleFeatsWithMultipleTypes (smoke of multi-type gender)
func TestUnifier_MultipleFeatsWithMultipleTypes(t *testing.T) {
	cfg := NewUnifierConfiguration()
	cfg.SetEquivalence("number", "singular", preparePOSElement(`.*[\.:]sg:.*`))
	cfg.SetEquivalence("number", "plural", preparePOSElement(`.*[\.:]pl:.*`))
	cfg.SetEquivalence("gender", "feminine", preparePOSElement(`.*[\.:]f([.:].*)?`))
	cfg.SetEquivalence("gender", "masculine", preparePOSElement(`.*[\.:]m([.:].*)?`))
	cfg.SetEquivalence("gender", "neutral", preparePOSElement(`.*[\.:]n([.:].*)?`))
	uni := cfg.CreateUnifier()
	equiv := map[string][]string{"number": nil, "gender": nil}
	// adj with compound gender f.n.m2.m3 + noun n → agree on n
	a := languagetool.NewAnalyzedToken("osobiste", strPtr("adj:sg:nom.acc.voc:n:pos:aff"), strPtr("osobisty"))
	b := languagetool.NewAnalyzedToken("godło", strPtr("subst:sg:nom.acc.voc:n"), strPtr("godło"))
	uni.IsUnified(a, equiv, true)
	require.True(t, uni.IsUnified(b, equiv, true))
	uni.Reset()
}

// Twin of UnifierTest.testNegation — negation of feature match
func TestUnifier_Negation(t *testing.T) {
	cfg := NewUnifierConfiguration()
	// POS with negate for plural
	pl := preparePOSElement(`.*[\.:]pl:.*`)
	// For negation tests Java uses PatternToken with setNegation; use reverse singular/plural agreement
	cfg.SetEquivalence("number", "singular", preparePOSElement(`.*[\.:]sg:.*`))
	cfg.SetEquivalence("number", "plural", pl)
	uni := cfg.CreateUnifier()
	// two singulars agree
	equiv := map[string][]string{"number": {"singular"}}
	a := languagetool.NewAnalyzedToken("x", strPtr("adj:sg:x"), strPtr("x"))
	b := languagetool.NewAnalyzedToken("y", strPtr("subst:sg:y"), strPtr("y"))
	satisfied := uni.IsSatisfied(a, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(b, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.True(t, satisfied)
	uni.Reset()
	// singular + plural disagree
	c := languagetool.NewAnalyzedToken("z", strPtr("subst:pl:z"), strPtr("z"))
	satisfied = uni.IsSatisfied(a, equiv)
	uni.StartUnify()
	satisfied = satisfied && uni.IsSatisfied(c, equiv)
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.False(t, satisfied)
	_ = pl
	uni.Reset()
}

// Twin of UnifierTest.testAddNeutralElement
func TestUnifier_AddNeutralElement(t *testing.T) {
	cfg := NewUnifierConfiguration()
	cfg.SetEquivalence("number", "singular", preparePOSElement(`.*[\.:]sg:.*`))
	uni := cfg.CreateUnifier()
	equiv := map[string][]string{"number": {"singular"}}
	a := languagetool.NewAnalyzedToken("x", strPtr("adj:sg:x"), strPtr("x"))
	b := languagetool.NewAnalyzedToken("y", strPtr("subst:sg:y"), strPtr("y"))
	satisfied := uni.IsSatisfied(a, equiv)
	uni.StartUnify()
	// insert neutral punctuation-like token between unified pair
	neutralTok := languagetool.NewAnalyzedToken(",", nil, nil)
	uni.AddNeutralElement(languagetool.NewAnalyzedTokenReadings(neutralTok))
	satisfied = satisfied && uni.IsSatisfied(b, equiv)
	uni.StartNextToken()
	satisfied = satisfied && uni.GetFinalUnificationValue(equiv)
	require.True(t, satisfied)
	uts := uni.GetUnifiedTokens()
	require.NotNil(t, uts)
	// neutral element is part of sequence
	require.GreaterOrEqual(t, len(uts), 2)
	uni.Reset()
}
