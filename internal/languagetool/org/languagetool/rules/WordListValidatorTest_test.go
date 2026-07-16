package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWordListValidatorTest_Stub(t *testing.T) {
	v := NewWordListValidator()
	require.Empty(t, v.ValidateLines(strings.NewReader("# ok\nhello\nworld\n")))
	errs := v.ValidateLines(strings.NewReader(" trailing\nbad\tword\n"))
	require.NotEmpty(t, errs)
}
