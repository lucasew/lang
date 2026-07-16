package languagetool

import "fmt"

// ErrorRateTooHighException ports org.languagetool.ErrorRateTooHighException.
type ErrorRateTooHighException struct {
	Message string
}

func NewErrorRateTooHighException(message string) *ErrorRateTooHighException {
	return &ErrorRateTooHighException{Message: message}
}

func (e *ErrorRateTooHighException) Error() string {
	if e == nil {
		return "error rate too high"
	}
	return e.Message
}

func (e *ErrorRateTooHighException) String() string {
	return fmt.Sprintf("ErrorRateTooHighException: %s", e.Error())
}
