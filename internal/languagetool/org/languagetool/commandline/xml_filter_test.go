package commandline

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMaybeFilterXML(t *testing.T) {
	in := "<p>Hello <b>world</b></p>"
	require.Equal(t, in, MaybeFilterXML(in, false))
	require.Equal(t, "Hello world", MaybeFilterXML(in, true))
}

func TestMain_XmlFiltering(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "--xmlfilter", "-"}, RunHooks{
		ReadStdin: func() (string, error) {
			return "<div>This is <em>an</em> test.</div>", nil
		},
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.True(t, opts.XMLFiltering)
			filtered := MaybeFilterXML(text, opts.XMLFiltering)
			require.Equal(t, "This is an test.", filtered)
			_, _ = io.WriteString(w, filtered)
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "This is an test.")
}

func TestMain_NoXmlFilteringByDefault(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "<b>x</b>", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.False(t, opts.XMLFiltering)
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
}

func TestMain_StdInWithExternalFalseFriends(t *testing.T) {
	var out, errb bytes.Buffer
	code := RunWithIO([]string{"-l", "en", "-m", "de", "-"}, RunHooks{
		ReadStdin: func() (string, error) { return "I became a doctor.", nil },
		Check: func(w io.Writer, text string, opts *CommandLineOptions) (int, error) {
			require.Equal(t, "de", opts.MotherTongue)
			require.Equal(t, "en", opts.Language)
			// false-friend rules would use mother tongue; surface green path
			_, _ = io.WriteString(w, "mother="+opts.MotherTongue+"\n")
			return 0, nil
		},
	}, &out, &errb)
	require.Equal(t, 0, code)
	require.Contains(t, out.String(), "mother=de")
}
