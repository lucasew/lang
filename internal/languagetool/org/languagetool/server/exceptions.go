package server

// Server-side runtime errors ported from org.languagetool.server.*Exception.

type TooManyRequestsError struct{ Msg string }

func (e *TooManyRequestsError) Error() string {
	if e == nil {
		return "too many requests"
	}
	return e.Msg
}

func NewTooManyRequestsError(msg string) error { return &TooManyRequestsError{Msg: msg} }

type TextTooLongError struct{ Msg string }

func (e *TextTooLongError) Error() string {
	if e == nil {
		return "text too long"
	}
	return e.Msg
}

func NewTextTooLongError(msg string) error { return &TextTooLongError{Msg: msg} }

type BadRequestError struct{ Msg string }

func (e *BadRequestError) Error() string {
	if e == nil {
		return "bad request"
	}
	return e.Msg
}

func NewBadRequestError(msg string) error { return &BadRequestError{Msg: msg} }

type AuthError struct{ Msg string }

func (e *AuthError) Error() string {
	if e == nil {
		return "auth error"
	}
	return e.Msg
}

func NewAuthError(msg string) error { return &AuthError{Msg: msg} }

type PortBindingError struct{ Msg string }

func (e *PortBindingError) Error() string {
	if e == nil {
		return "port binding failed"
	}
	return e.Msg
}

func NewPortBindingError(msg string) error { return &PortBindingError{Msg: msg} }

type UnavailableError struct {
	Msg   string
	Cause error
}

func (e *UnavailableError) Error() string {
	if e == nil {
		return "unavailable"
	}
	if e.Cause != nil {
		return e.Msg + ": " + e.Cause.Error()
	}
	return e.Msg
}

func (e *UnavailableError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func NewUnavailableError(msg string, cause error) error {
	return &UnavailableError{Msg: msg, Cause: cause}
}

type PathNotFoundError struct{ Msg string }

func (e *PathNotFoundError) Error() string {
	if e == nil {
		return "path not found"
	}
	return e.Msg
}

func NewPathNotFoundError(msg string) error { return &PathNotFoundError{Msg: msg} }

type IllegalConfigurationError struct{ Msg string }

func (e *IllegalConfigurationError) Error() string {
	if e == nil {
		return "illegal configuration"
	}
	return e.Msg
}

func NewIllegalConfigurationError(msg string) error { return &IllegalConfigurationError{Msg: msg} }

// IllegalPipelineMutationError is thrown when a frozen pipeline is mutated.
type IllegalPipelineMutationError struct{}

func (e *IllegalPipelineMutationError) Error() string {
	return "Pipeline is frozen; mutating shared instance is forbidden."
}
