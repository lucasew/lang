package commandline

import "fmt"

// WrongParameterNumberException ports commandline.WrongParameterNumberException.
type WrongParameterNumberException struct{}

func (WrongParameterNumberException) Error() string {
	return "wrong number of parameters"
}

// UnknownParameterException ports commandline.UnknownParameterException.
type UnknownParameterException struct {
	Param string
}

func (e UnknownParameterException) Error() string {
	return fmt.Sprintf("unknown parameter: %s", e.Param)
}
