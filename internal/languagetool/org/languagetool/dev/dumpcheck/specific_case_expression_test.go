package dumpcheck

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpecificCaseCounter_Observe(t *testing.T) {
	c := NewSpecificCaseCounter()
	c.ObserveSentence("I visited New York yesterday afternoon.")
	c.ObserveSentence("We left New York last week.")
	c.ObserveSentence("Flying over New York takes time.")
	require.Equal(t, 3, c.Count("New York"))
	top := c.Top(3)
	require.Equal(t, "New York", top[0])
}

func TestSpecificCaseCounter_FromSource(t *testing.T) {
	src := NewPlainTextSentenceSource(strings.NewReader(
		"The United Nations held a meeting today.\n" +
			"United Nations members gathered again.\n"))
	c := NewSpecificCaseCounter()
	require.NoError(t, c.ObserveSource(src))
	require.GreaterOrEqual(t, c.Count("United Nations")+c.Count("The United Nations"), 2)
}
