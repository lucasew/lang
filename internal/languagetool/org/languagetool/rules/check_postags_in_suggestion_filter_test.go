package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckPostagsInSuggestionFilter(t *testing.T) {
	f := NewCheckPostagsInSuggestionFilter(func(tok string) []string {
		switch tok {
		case "the":
			return []string{"DT"}
		case "cat":
			return []string{"NN"}
		case "runs":
			return []string{"VBZ"}
		default:
			return []string{"XX"}
		}
	})
	got := f.Filter([]string{"the cat", "runs cat", "the runs"}, "DT,NN")
	require.Equal(t, []string{"the cat"}, got)
}
