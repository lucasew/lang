package corepack_test

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
	"github.com/stretchr/testify/require"
)

func TestRegister_PortuguesePriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt")
	corepack.Register(lt, "pt")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, -50, lt.PriorityForId("MORFOLOGIK_RULE_PT_PT"))
	require.Equal(t, -52, lt.PriorityForId("COLOCACAO_PRONOMINAL_COM_ATRATOR_X"))
}

func TestRegister_DutchPriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("nl")
	corepack.Register(lt, "nl")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, 1, lt.PriorityForId("NL_SIMPLE_REPLACE_X"))
	require.Equal(t, 3, lt.PriorityForId("SINT_X"))
	require.Equal(t, -51, lt.PriorityForId("AI_NL_HYDRA_LEO_MISSING_COMMA_X"))
}

func TestRegister_PortugalPortuguesePriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("pt-PT")
	corepack.Register(lt, "pt")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, 1, lt.PriorityForId("PT_COMPOUNDS_POST_REFORM"))
	require.Equal(t, -9, lt.PriorityForId("PORTUGUESE_OLD_SPELLING_INTERNAL"))
	// Brazilian: base Portuguese map (−45), not PortugalPortuguese override (+1)
	ltBR := languagetool.NewJLanguageTool("pt-BR")
	corepack.Register(ltBR, "pt")
	require.Equal(t, -45, ltBR.PriorityForId("PT_COMPOUNDS_POST_REFORM"))
}
