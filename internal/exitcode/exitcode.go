package exitcode

import (
	"errors"

	"github.com/lucasew/lang/internal/finding"
)

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

// IsErrorSeverity reports whether a SARIF severity should fail the process (exit 1).
// Default policy: only SARIF "error" fails; warning/note/none do not.
func IsErrorSeverity(sev string) bool {
	return sev == finding.SeverityError
}
