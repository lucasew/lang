package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	tagginges "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/es"
	"github.com/stretchr/testify/require"
)

func TestWireSpanishSpellerTagger(t *testing.T) {
	wt := tagging.MapWordTagger{
		"come": {tagging.NewTaggedWord("comer", "VMIP3S0")},
	}
	tg := tagginges.NewSpanishTagger(wt)
	r := NewMorfologikSpanishSpellerRule()
	WireSpanishSpellerTagger(r, tg)
	require.NotNil(t, r.TagPOS)
	got := r.orderSpanishSuggestions([]string{"casa", "me come"}, "xxx")
	require.Equal(t, "me come", got[0])
}

func TestWireSpanishSpellerTagPOS(t *testing.T) {
	r := NewMorfologikSpanishSpellerRule()
	WireSpanishSpellerTagPOS(r, func(token string) []languagetool.TokenTag {
		if token == "casa" {
			return []languagetool.TokenTag{{POS: "NCMS000"}}
		}
		return nil
	})
	require.Equal(t, []string{"casa 2"}, r.additionalTopSpanishSuggestions("casa2"))
}
