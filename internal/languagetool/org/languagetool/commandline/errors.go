package commandline

// WrongParameterNumberException ports commandline.WrongParameterNumberException
// (Java empty RuntimeException message).
type WrongParameterNumberException struct{}

func (WrongParameterNumberException) Error() string {
	// Java RuntimeException default message is null; Go needs non-empty Error.
	return "WrongParameterNumberException"
}

// UnknownParameterException ports commandline.UnknownParameterException.
// Java: UnknownParameterException(String message) { super(message); }
type UnknownParameterException struct {
	Message string
	// Param kept for callers that set Param; Message is authoritative.
	Param string
}

// NewUnknownParameterException ports the Java constructor with message.
func NewUnknownParameterException(message string) UnknownParameterException {
	return UnknownParameterException{Message: message, Param: message}
}

func (e UnknownParameterException) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Param != "" {
		return e.Param
	}
	return "UnknownParameterException"
}
