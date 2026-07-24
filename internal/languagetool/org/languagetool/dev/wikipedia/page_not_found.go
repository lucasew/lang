package wikipedia

// PageNotFoundError ports PageNotFoundException.
type PageNotFoundError struct {
	Msg string
}

func (e PageNotFoundError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return "page not found"
}
