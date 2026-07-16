package dumpcheck

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAtDResultXML(t *testing.T) {
	xml := `<?xml version="1.0"?><results>
  <error>
    <string>teh</string>
    <description>Spelling</description>
  </error>
  <error>
    <string>a hour</string>
    <description>Grammar</description>
  </error>
</results>`
	matches, err := ParseAtDResultXML(xml)
	require.NoError(t, err)
	require.Len(t, matches, 2)
	require.Equal(t, "Spelling: teh", matches[0].Format())
	require.Equal(t, "Grammar: a hour", matches[1].Format())
}

func TestAfterTheDeadlineChecker_Run(t *testing.T) {
	src := NewPlainTextSentenceSource(strings.NewReader(
		"First long enough sentence here.\nSecond long enough sentence here.\nThird long enough sentence here.\n"))
	atd := NewAfterTheDeadlineChecker("http://localhost/check?data=", 2)
	atd.Query = func(text string) (string, error) {
		return `<results><error><string>x</string><description>E</description></error></results>`, nil
	}
	res, err := atd.Run(src)
	require.NoError(t, err)
	require.Len(t, res, 2)
	require.Len(t, res[0].Matches, 1)
	require.Equal(t, "E: x", res[0].Matches[0].Format())
}
