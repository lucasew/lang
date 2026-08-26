package dumpcheck

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommonCrawlSentenceSource_Filters(t *testing.T) {
	in := strings.Join([]string{
		"this starts lower and should skip.",
		"Too short.",
		"This is a valid common crawl sentence.",
		"No terminal punctuation at all here today",
		"Another valid common crawl line here!",
	}, "\n")
	src := NewCommonCrawlSentenceSource(strings.NewReader(in))
	var got []string
	for src.HasNext() {
		s, err := src.Next()
		require.NoError(t, err)
		got = append(got, s.GetText())
	}
	require.Equal(t, []string{
		"This is a valid common crawl sentence.",
		"Another valid common crawl line here!",
	}, got)
	require.Equal(t, "commoncrawl", src.GetSource())
	require.Greater(t, src.WrongStartChar, 0)
}
