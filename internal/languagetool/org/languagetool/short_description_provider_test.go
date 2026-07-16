package languagetool

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/broker"
	"github.com/stretchr/testify/require"
)

func TestShortDescriptionProvider(t *testing.T) {
	b := broker.NewMapResourceDataBroker()
	b.Resource["en/word_definitions.txt"] = "#c\ncat\ta feline\ndog\ta canine\n"
	p := NewShortDescriptionProvider(b)
	require.Equal(t, "a feline", p.GetShortDescription("cat", "en"))
	require.Equal(t, "", p.GetShortDescription("bird", "en"))
	require.Equal(t, "", p.GetShortDescription("cat", "de"))
}
