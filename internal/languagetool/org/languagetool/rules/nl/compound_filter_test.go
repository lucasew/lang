package nl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of CompoundFilterTest.testFilter
func TestCompoundFilter_Filter(t *testing.T) {
	f := NewCompoundFilter()
	cases := []struct {
		words []string
		want  string
	}{
		{[]string{"tv", "meubel"}, "tv-meubel"},
		{[]string{"test-tv", "meubel"}, "test-tv-meubel"},
		{[]string{"onzin", "tv"}, "onzin-tv"},
		{[]string{"auto", "onderdeel"}, "auto-onderdeel"},
		{[]string{"test", "e-mail"}, "test-e-mail"},
		{[]string{"taxi", "jongen"}, "taxi-jongen"},
		{[]string{"rij", "instructeur"}, "rijinstructeur"},
		{[]string{"ANWB", "wagen"}, "ANWB-wagen"},
		{[]string{"pro-deo", "advocaat"}, "pro-deoadvocaat"},
		{[]string{"ANWB", "tv", "wagen"}, "ANWB-tv-wagen"},
	}
	for _, tc := range cases {
		require.Equal(t, tc.want, f.Suggest(tc.words), "words=%v", tc.words)
	}
}
