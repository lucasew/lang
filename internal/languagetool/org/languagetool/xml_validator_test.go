package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestXMLValidator(t *testing.T) {
	v := NewXMLValidator()
	require.NoError(t, v.ValidateWellFormed(`<root><a/></root>`))
	require.Error(t, v.ValidateWellFormed(`<root><a></root>`))
}
