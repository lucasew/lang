package spelling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommonFileTypes_Suffix(t *testing.T) {
	re := GetSuffixPattern()
	require.True(t, re.MatchString("photo.JPG"))
	require.True(t, re.MatchString("report.pdf"))
	require.False(t, re.MatchString("hello"))
	require.False(t, re.MatchString(".pdf")) // needs something before extension per pattern
}
