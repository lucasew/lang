package sr

// Twin of MorfologikSerbianSpellerRuleTest (Java has no @Test).
// Surface smoke via ekavian/jekavian constructors from sibling packages.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sr/ekavian"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sr/jekavian"
	"github.com/stretchr/testify/require"
)

// Port of MorfologikSerbianSpellerRuleTest (no @Test)
func TestMorfologikSerbianSpellerRule_NoTests(t *testing.T) {
	e := ekavian.NewMorfologikEkavianSpellerRule()
	require.Equal(t, ekavian.MorfologikEkavianSpellerRuleID, e.GetID())
	require.Equal(t, ekavian.EkavianSpellerDict, e.GetFileName())

	j := jekavian.NewMorfologikJekavianSpellerRule()
	require.Equal(t, jekavian.MorfologikJekavianSpellerRuleID, j.GetID())
	require.Equal(t, jekavian.JekavianSpellerDict, j.GetFileName())
}
