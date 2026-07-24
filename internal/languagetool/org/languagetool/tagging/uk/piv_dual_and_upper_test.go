package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/stretchr/testify/require"
)

func TestDynamicPivNvDual(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		switch w {
		case "півгодини":
			return []tagging.TaggedWord{{Lemma: "півгодини", PosTag: "noun:inanim:p:v_naz:nv"}}
		case "годину":
			return []tagging.TaggedWord{{Lemma: "година", PosTag: "noun:inanim:f:v_zna"}}
		}
		return nil
	}
	rs := DynamicPivNvDualReadings("півгодини-годину", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), ":p:")
	require.Contains(t, *rs[0].GetPOSTag(), "noun:inanim")
	require.Equal(t, "півгодини-годину", *rs[0].GetLemma())
	// case-sensitive startsWith пів
	require.Nil(t, DynamicPivNvDualReadings("Півгодини-годину", tag))
	require.Nil(t, DynamicPivNvDualReadings("півгодини-годину", nil))
}

func TestDynamicUpperRight_adjBad(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "кримсько-татарський" {
			return []tagging.TaggedWord{{Lemma: "кримсько-татарський", PosTag: "adj:m:v_naz:bad"}}
		}
		return nil
	}
	rs := DynamicUpperRightCompoundReadings("Кримсько-Татарський", tag)
	require.NotEmpty(t, rs)
	require.Contains(t, *rs[0].GetPOSTag(), "bad")
}

func TestDynamicUpperRightBlocks(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		switch w {
		case "Нью", "нью":
			return []tagging.TaggedWord{{Lemma: "Нью", PosTag: "noninfl:prop"}}
		case "Париж", "париж":
			return []tagging.TaggedWord{{Lemma: "Париж", PosTag: "noun:inanim:m:v_naz:prop:geo"}}
		}
		return nil
	}
	// upper right, not dual non-prop noun → block
	require.True(t, DynamicUpperRightBlocks("Нью-Париж", tag))
}

func TestDynamicNoDashSolidHasTags(t *testing.T) {
	tag := func(w string) []tagging.TaggedWord {
		if w == "зовнішньоекономічний" {
			return []tagging.TaggedWord{{Lemma: "зовнішньоекономічний", PosTag: "adj:m:v_naz"}}
		}
		if w == "зовнішньо" {
			return []tagging.TaggedWord{{Lemma: "зовнішньо", PosTag: "adv"}}
		}
		return nil
	}
	require.True(t, DynamicNoDashSolidHasTags("зовнішньо-економічний", tag))
	require.False(t, DynamicNoDashSolidHasTags("вгору-вниз", tag))
}
