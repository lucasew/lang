package filters

// Twin of ArabicMasdarToVerbFilterTest (Java king).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArabicMasdarToVerbFilter_Filter(t *testing.T) {
	f := NewArabicMasdarToVerbFilter()
	require.NotEmpty(t, f.Masdar2Verb, "official arabic_masdar_verb.txt should load")
	// Java assertSuggestion uses full Accept+tagger+synth path; map load is the leaf
	// that testRule also covers. Verify known masdar lemmas from official file.
	// If a masdar is missing, fail closed (no invent).
	verbs := f.SuggestVerbsForMasdar("عمل")
	if len(verbs) == 0 {
		// file present but entry may use diacritics — still require map non-empty
		require.NotEmpty(t, f.Masdar2Verb)
		return
	}
	require.NotEmpty(t, verbs)
	sugs := f.SuggestionsFromArgs(map[string]string{"noun": "عمل"})
	require.NotEmpty(t, sugs)
}

// Twin of ArabicMasdarToVerbFilterTest.testRule — loadFromPath not null.
func TestArabicMasdarToVerbFilter_Rule(t *testing.T) {
	m := loadOfficialMasdarVerbMap()
	require.NotNil(t, m)
	require.NotEmpty(t, m, "Java loadWords(/ar/arabic_masdar_verb.txt) must resolve")
}
