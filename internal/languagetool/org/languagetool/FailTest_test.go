package languagetool

import "testing"

// Port of org.languagetool.FailTest — Java marks this @Ignore (circleci only).
func TestFail_Fail(t *testing.T) {
	t.Skip("Java FailTest is @Ignore; not run in normal suites")
}
