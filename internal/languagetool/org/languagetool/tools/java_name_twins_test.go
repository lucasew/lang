package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJavaNameTwins(t *testing.T) {
	require.True(t, (StringTools{}).IsEmpty(""))
	require.Equal(t, "hi", (StringInterner{}).Intern("hi"))
	require.Equal(t, LogMarkerInit, (LoggingTools{}).MarkerInit())
	require.Contains(t, (Tools{}).I18n("x {0}", "a"), "a")
}
