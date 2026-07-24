package server

// Twin of RemoteSynthesizerTest — full multi-lang synth dict deferred; pluggable backend.
import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteSynthesizer_Synthesis(t *testing.T) {
	// Map of language+lemma+tag → forms (green stand-in for RemoteSynthesizer Java with real dicts)
	type key struct{ lang, lemma, tag string }
	exact := map[key][]string{
		{"de", "Äußerung", "SUB:NOM:PLU:FEM"}: {"Äußerungen"},
		{"pt", "resolver", "VMIS3S0"}:         {"resolveu"},
		{"es", "cantar", "VMIP1S0"}:           {"canto"},
		{"fr", "monde", "N m p"}:              {"mondes"},
		{"en", "be", "VBZ"}:                   {"is"},
		{"en-US", "be", "VBZ"}:                {"is"},
		{"en-GB", "be", "VBZ"}:                {"is"},
	}
	reForms := map[key][]string{
		{"de", "Äußerung", "SUB:.*:PLU:FEM"}: {"Äußerungen", "Äußerungen"}, // dups removed
		{"es", "señor", "NC.P.*"}:            {"señoras", "señores"},
		{"fr", "chanter", "V ppa.*"}:         {"chantées", "chantée", "chantés", "chanté"},
		{"en", "be", "V.*"}:                  {"be", "was", "were", "being", "been", "are", "is"},
		{"en", "be", "N.*"}:                  {},
	}

	rs := NewRemoteSynthesizer(func(lang, lemma, postag string, postagRegexp bool) ([]string, error) {
		k := key{lang, lemma, postag}
		if postagRegexp {
			if f, ok := reForms[k]; ok {
				return f, nil
			}
			// try match patterns against exact map for same lang/lemma
			for ek, forms := range exact {
				if ek.lang == lang && ek.lemma == lemma {
					if ok, _ := regexp.MatchString("^"+postag+"$", ek.tag); ok {
						return forms, nil
					}
				}
			}
			return nil, nil
		}
		return exact[k], nil
	})

	got, err := rs.SynthesizeForms("de", "Äußerung", "SUB:NOM:PLU:FEM", false)
	require.NoError(t, err)
	require.Equal(t, []string{"Äußerungen"}, got)

	got, err = rs.SynthesizeForms("de", "Äußerung", "SUB:.*:PLU:FEM", true)
	require.NoError(t, err)
	require.Equal(t, []string{"Äußerungen"}, got) // dups removed

	got, err = rs.SynthesizeForms("en", "be", "VBZ", false)
	require.NoError(t, err)
	require.Equal(t, []string{"is"}, got)

	got, err = rs.SynthesizeForms("en-US", "be", "VBZ", false)
	require.NoError(t, err)
	require.Equal(t, []string{"is"}, got)

	got, err = rs.SynthesizeForms("en", "be", "V.*", true)
	require.NoError(t, err)
	require.Equal(t, []string{"be", "was", "were", "being", "been", "are", "is"}, got)

	got, err = rs.SynthesizeForms("en", "be", "N.*", true)
	require.NoError(t, err)
	require.Empty(t, got)

	// nil synthesizer
	nilRS := NewRemoteSynthesizer(nil)
	got, err = nilRS.SynthesizeForms("en", "be", "VBZ", false)
	require.NoError(t, err)
	require.Nil(t, got)
}
