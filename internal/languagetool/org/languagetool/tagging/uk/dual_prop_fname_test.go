package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestDynamicDualProp_fnameTagMatch(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		switch w {
		case "Хуана", "хуана":
			return []tagging.TaggedWord{{Lemma: "Хуан", PosTag: "noun:anim:m:v_rod:prop:fname"}}
		case "Карлоса", "карлоса":
			return []tagging.TaggedWord{{Lemma: "Карлос", PosTag: "noun:anim:m:v_rod:prop:fname"}}
		}
		return nil
	}
	rs := DynamicDualPropReadings("Хуана-Карлоса", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, rs[0].POS, "fname")
	require.Contains(t, rs[0].Lemma, "-")
}
