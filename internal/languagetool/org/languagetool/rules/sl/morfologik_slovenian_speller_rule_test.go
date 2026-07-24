package sl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikSlovenianSpellerRule(t *testing.T) {
	r := NewMorfologikSlovenianSpellerRule()
	// Java MorfologikSlovenianSpellerRule.getId / getFileName
	require.Equal(t, "MORFOLOGIK_RULE_SL_SI", MorfologikSlovenianSpellerRuleID)
	require.Equal(t, "/sl/hunspell/sl_SI.dict", SlovenianSpellerDict)
	require.Equal(t, MorfologikSlovenianSpellerRuleID, r.GetID())
	require.Equal(t, SlovenianSpellerDict, r.GetFileName())
}
