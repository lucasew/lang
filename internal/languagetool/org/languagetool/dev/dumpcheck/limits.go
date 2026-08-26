package dumpcheck

import "fmt"

// DocumentLimitReachedError ports DocumentLimitReachedException.
type DocumentLimitReachedError struct {
	Limit int
}

func (e DocumentLimitReachedError) Error() string {
	return fmt.Sprintf("Maximum number of documents (%d) reached", e.Limit)
}

func (e DocumentLimitReachedError) GetLimit() int { return e.Limit }

// ErrorLimitReachedError ports ErrorLimitReachedException.
type ErrorLimitReachedError struct {
	Limit int
}

func (e ErrorLimitReachedError) Error() string {
	return fmt.Sprintf("Maximum number of errors (%d) reached", e.Limit)
}

func (e ErrorLimitReachedError) GetLimit() int { return e.Limit }
