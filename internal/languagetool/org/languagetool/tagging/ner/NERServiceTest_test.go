package ner

// Twin of NERServiceTest.testParseBuffer
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNERService_ParseBuffer(t *testing.T) {
	res := ParseBuffer("This/O/0/4 is/O/5/7 Peter/PERSON/8/13 's/O/13/15 job/O/16/19 ./O/19/20")
	require.Len(t, res, 1)
	require.Equal(t, 8, res[0].GetStart())
	require.Equal(t, 13, res[0].GetEnd())
}
