package tools

// WrappedVoid ports org.languagetool.tools.WrappedVoid.
type WrappedVoid func() error

// Call runs the void function.
func (w WrappedVoid) Call() error {
	if w == nil {
		return nil
	}
	return w()
}

// WrappedValue ports org.languagetool.tools.WrappedValue.
type WrappedValue[T any] func() (T, error)

// Call runs the value function.
func (w WrappedValue[T]) Call() (T, error) {
	if w == nil {
		var zero T
		return zero, nil
	}
	return w()
}
