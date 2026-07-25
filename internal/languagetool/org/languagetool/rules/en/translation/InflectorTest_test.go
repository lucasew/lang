package translation

// Twin of org.languagetool.rules.en.translation.InflectorTest
// Uses real English synthesizer + official english_synth.dict (no invent FuncSynthesizer).

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func requireEnglishSynth(t *testing.T) {
	t.Helper()
	if discoverEnglishSynthDict() == "" || defaultEnglishSynthesizer() == nil {
		t.Skip("english_synth.dict not discoverable — Inflector needs Languages en synthesizer like Java")
	}
}

// Port of InflectorTest.inflect
func TestInflector_Inflect(t *testing.T) {
	requireEnglishSynth(t)
	inflector := NewInflector()
	// Java: assertThat(...toString(), is("[pumps]")) etc.
	require.Equal(t, []string{"pumps"}, inflector.Inflect("pump", "SUB:AKK:PLU:FEM"))     // "Pumpen"
	require.Equal(t, []string{"children"}, inflector.Inflect("child", "SUB:NOM:PLU:NEU")) // "Kinder"
	require.Equal(t, []string{"walks"}, inflector.Inflect("walk", "VER:3:SIN:PRÄ:NON"))   // "er geht"
	require.Equal(t, []string{"walked"}, inflector.Inflect("walk", "VER:3:SIN:PRT:NON"))  // "ging" (-> gehen)
	require.Equal(t, []string{"walking"}, inflector.Inflect("walk", "PA1:PRD:GRU:VER"))   // "gehend"
	require.Equal(t, []string{"walked"}, inflector.Inflect("walk", "PA2:PRD:GRU:VER"))    // "gegangen"
	// forms that are not inflected in English do not need to be supported here

	require.Equal(t, []string{"larger"}, inflector.Inflect("large", "ADJ:PRD:KOM"))              // "größer"
	require.Equal(t, []string{"largest"}, inflector.Inflect("large", "ADJ:AKK:SIN:FEM:SUP:DEF")) // "größte"

	require.Equal(t, []string{"walk"}, inflector.Inflect("walk", "VER:1:SIN:PRÄ:NON")) // "ich gehe"
	require.Equal(t, []string{"walk"}, inflector.Inflect("walk", "VER:2:SIN:PRÄ:NON")) // "du gehst"
	require.Equal(t, []string{"walk"}, inflector.Inflect("walk", "VER:1:PLU:PRÄ:NON")) // "wir gehen"
	require.Equal(t, []string{"walk"}, inflector.Inflect("walk", "VER:2:PLU:PRÄ:NON")) // "ihr geht"
	require.Equal(t, []string{"walk"}, inflector.Inflect("walk", "VER:3:PLU:PRÄ:NON")) // "sie gehen"

	// Java null → Go empty string stand-in
	require.Equal(t, []string{"walk"}, inflector.Inflect("walk", ""))
	require.Equal(t, []string{"walk"}, inflector.Inflect("walk", "FAKE-TAG"))
}

// Port of InflectorTest.inflectMultiWord
func TestInflector_InflectMultiWord(t *testing.T) {
	requireEnglishSynth(t)
	inflector := NewInflector()
	require.Equal(t, []string{"tire pumps"}, inflector.Inflect("tire pump", "SUB:AKK:PLU:FEM")) // "Pumpen"
	require.Equal(t, []string{"fake tire pumps"}, inflector.Inflect("fake tire pump", "SUB:AKK:PLU:FEM"))
}
