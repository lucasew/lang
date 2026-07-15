package exitcode

import "errors"

const (
	OK          = 0
	HasErrors   = 1
	ToolFailure = 2
)

// ExitError carries a process exit code from a successful lint run that found issues.
type ExitError struct {
	Code int
}

func (e *ExitError) Error() string {
	return "lint finished with findings"
}

func HasErrorFindings() error {
	return &ExitError{Code: HasErrors}
}

func FromError(err error) int {
	if err == nil {
		return OK
	}
	var ee *ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return ToolFailure
}

// IsErrorSeverity reports whether a finding severity should fail the process (exit 1).
// LanguageTool does not use eslint-style levels; we treat spelling/unknown-word class
// as error and everything else as non-failing by default (SPEC exit policy B).
func IsErrorSeverity(sev string) bool {
	switch sev {
	case "misspelling", "UnknownWord", "unknownword", "error":
		return true
	default:
		return false
	}
}
