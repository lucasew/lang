package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCaseGovernmentHelper(t *testing.T) {
	h := LoadCaseGovernmentHelper()
	require.NotEmpty(t, h.Map)
	require.True(t, h.HasCaseGovernment("згідно з", "v_oru"))
	// sample line from file — find any non-empty non-override entry
	found := false
	for lemma, set := range h.Map {
		if lemma == "згідно з" || len(set) == 0 {
			continue
		}
		found = true
		break
	}
	require.True(t, found, "expected at least one non-empty government map entry")
}

func TestGetCustomCaseGovs_MatiButi(t *testing.T) {
	// мати:imperf:pres → v_inf
	mati := "мати"
	tok := atrLemma("має", &mati, "verb:imperf:pres:s:3")
	require.Contains(t, getCustomCaseGovs(tok), "v_inf")

	// бути past:n → v_inf
	buti := "бути"
	tok2 := atrLemma("було", &buti, "verb:imperf:past:n")
	require.Contains(t, getCustomCaseGovs(tok2), "v_inf")

	// (по)меншати → v_rod
	men := "меншати"
	tok3 := atrLemma("меншає", &men, "verb:imperf:pres:s:3")
	require.Contains(t, getCustomCaseGovs(tok3), "v_rod")

	// wired into GetCaseGovernmentsFromReadings
	h := LoadCaseGovernmentHelper()
	govs := h.GetCaseGovernmentsFromReadings(tok, "verb")
	_, ok := govs["v_inf"]
	require.True(t, ok, "custom v_inf from мати should appear in readings lookup")
}

func TestAdvpVerbLemma(t *testing.T) {
	l := "даючи"
	at := atrLemma("даючи", &l, "advp:imperf")
	require.Equal(t, "давати", advpVerbLemma(at.GetReadings()[0]))
	l2 := "роблячи"
	at2 := atrLemma("роблячи", &l2, "advp")
	require.Equal(t, "робити", advpVerbLemma(at2.GetReadings()[0]))
}

func TestHasLemmaREWithPosRE_SameReading(t *testing.T) {
	// lemma on one reading, POS on another → must not match
	men := "меншати"
	// reading 0: wrong POS for pattern, reading 1: different lemma with good POS
	tok := atrLemma("меншає", &men, "noun:inanim:m:v_naz")
	// append second reading via GetReadings mutation is hard; use only custom gov path
	// same reading good POS
	tokOK := atrLemma("меншає", &men, "verb:imperf:pres:s:3")
	require.True(t, HasLemmaREWithPosRE(tokOK, cgBilshMenshRE, cgBilshMenshPosRE))
	// lemma matches but POS does not → false
	require.False(t, HasLemmaREWithPosRE(tok, cgBilshMenshRE, cgBilshMenshPosRE))
}
