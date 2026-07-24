package translation

// Twin of InflectorTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func injectSynth() synthesis.Synthesizer {
	return synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			return []string{token.GetToken() + "s"}, nil
		},
		SynthRE: func(token *languagetool.AnalyzedToken, posTag string, _ bool) ([]string, error) {
			// crude: map POS regex to English-ish forms used in Java test
			w := token.GetToken()
			switch {
			case posTag == "NNP?S" || matchRE(posTag, `NNP?S`):
				if w == "child" {
					return []string{"children"}, nil
				}
				return []string{w + "s"}, nil
			case posTag == "VBZ":
				return []string{w + "s"}, nil
			case posTag == "VBD" || posTag == "VBN":
				return []string{w + "ed"}, nil
			case posTag == "VBG":
				return []string{w + "ing"}, nil
			case posTag == "JJR":
				return []string{w + "r"}, nil // large → larger would need stem; soft
			case posTag == "JJS":
				return []string{w + "st"}, nil
			default:
				if matchRE(posTag, `NNP?S`) {
					if w == "child" {
						return []string{"children"}, nil
					}
					return []string{w + "s"}, nil
				}
				if matchRE(posTag, `VBZ`) {
					return []string{w + "s"}, nil
				}
				if matchRE(posTag, `VBD|VBN`) {
					return []string{w + "ed"}, nil
				}
				if matchRE(posTag, `VBG`) {
					return []string{w + "ing"}, nil
				}
				if matchRE(posTag, `JJR`) {
					if w == "large" {
						return []string{"larger"}, nil
					}
					return []string{w + "er"}, nil
				}
				if matchRE(posTag, `JJS`) {
					if w == "large" {
						return []string{"largest"}, nil
					}
					return []string{w + "est"}, nil
				}
				return []string{w}, nil
			}
		},
	}
}

// Port of InflectorTest.inflect
func TestInflector_Inflect(t *testing.T) {
	inf := NewInflector(injectSynth())
	require.Equal(t, []string{"pumps"}, inf.Inflect("pump", "SUB:AKK:PLU:FEM"))
	require.Equal(t, []string{"children"}, inf.Inflect("child", "SUB:NOM:PLU:NEU"))
	require.Equal(t, []string{"walks"}, inf.Inflect("walk", "VER:3:SIN:PRÄ:NON"))
	require.Equal(t, []string{"walked"}, inf.Inflect("walk", "VER:3:SIN:PRT:NON"))
	require.Equal(t, []string{"walking"}, inf.Inflect("walk", "PA1:PRD:GRU:VER"))
	require.Equal(t, []string{"walked"}, inf.Inflect("walk", "PA2:PRD:GRU:VER"))
	require.Equal(t, []string{"larger"}, inf.Inflect("large", "ADJ:PRD:KOM"))
	require.Equal(t, []string{"largest"}, inf.Inflect("large", "ADJ:AKK:SIN:FEM:SUP:DEF"))
	// uninflected persons
	require.Equal(t, []string{"walk"}, inf.Inflect("walk", "VER:1:SIN:PRÄ:NON"))
	require.Equal(t, []string{"walk"}, inf.Inflect("walk", ""))
	require.Equal(t, []string{"walk"}, inf.Inflect("walk", "FAKE-TAG"))
}

// Port of InflectorTest.inflectMultiWord
func TestInflector_InflectMultiWord(t *testing.T) {
	inf := NewInflector(injectSynth())
	require.Equal(t, []string{"tire pumps"}, inf.Inflect("tire pump", "SUB:AKK:PLU:FEM"))
	require.Equal(t, []string{"fake tire pumps"}, inf.Inflect("fake tire pump", "SUB:AKK:PLU:FEM"))
}
