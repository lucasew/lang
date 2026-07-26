package pt

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Meta + message twins for PortugueseUnitConversionRule (match matrix lives in
// PortugueseUnitConversionRuleTest_test.go).

func TestPortugueseUnitConversionRule_ID(t *testing.T) {
	r := NewPortugueseUnitConversionRule(nil)
	// Java PortugueseUnitConversionRule.getId / getDescription
	require.Equal(t, "UNIDADES_METRICAS", r.GetID())
	require.Equal(t, "Sugere ou verifica informações equivalentes à métrica de unidades de medida específicas.", r.GetDescription())
}

func TestPortugueseUnitConversionRule_Messages(t *testing.T) {
	r := NewPortugueseUnitConversionRule(nil)
	// Java getMessage / getShortMessage
	require.Equal(t, "Esta conversão não parece estar precisa. Gostaria de corrigi-la?", r.GetMessage(rules.UnitMsgCheck))
	require.Equal(t, "Deseja adicionar automaticamente uma conversão ao sistema métrico?", r.GetMessage(rules.UnitMsgSuggestion))
	require.Equal(t, "A unidade usada nesta conversão não foi reconhecida.", r.GetMessage(rules.UnitMsgCheckUnknownUnit))
	require.Equal(t, "Estas unidades de medida não são compatíveis.", r.GetMessage(rules.UnitMsgUnitMismatch))
	require.Equal(t, "Conversão incorreta. Corrigir?", r.GetShortMessage(rules.UnitMsgCheck))
	require.Equal(t, "Adicionar conversão ao sistema métrico?", r.GetShortMessage(rules.UnitMsgSuggestion))
	require.Equal(t, "Unidade desconhecida.", r.GetShortMessage(rules.UnitMsgCheckUnknownUnit))
	require.Equal(t, "Unidade incompatível.", r.GetShortMessage(rules.UnitMsgUnitMismatch))
	// Java formatRounded
	require.Equal(t, "aprox. 1,83 m", r.FormatRounded("1,83 m"))
}

func TestBrazilianToponymMapLoader_Smoke(t *testing.T) {
	// Full twins live in brazilian_toponym_map_loader_test.go.
	m := NewBrazilianToponymMapLoader().BuildMap()
	require.Equal(t, 27, len(m))
	require.Contains(t, m["SP"], "são paulo")
}

func TestConfusionPairsDataLoader(t *testing.T) {
	var l ConfusionPairsDataLoader
	in := strings.NewReader("secreto;secréto;AQ0MS0\n")
	m, err := l.LoadWords(in, "test")
	require.NoError(t, err)
	require.Contains(t, m, "secreto")
}
