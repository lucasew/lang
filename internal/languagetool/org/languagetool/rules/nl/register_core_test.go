package nl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreDutchRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("nl")
	RegisterCoreDutchRules(lt)
	m := lt.Check("Hij deed alsof als of hij niet hoorde.")
	// also match "als of" sequence
	m2 := lt.Check("Het lijkt als of het regent.")
	found := false
	for _, x := range append(m, m2...) {
		if x.RuleID == "NL_ALS_OF" {
			found = true
		}
	}
	require.True(t, found, "%+v %+v", m, m2)
}
