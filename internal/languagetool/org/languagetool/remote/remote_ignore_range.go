package remote

// RemoteIgnoreRange ports org.languagetool.remote.RemoteIgnoreRange.
type RemoteIgnoreRange struct {
	From int
	To   int
	Lang string // languageCode
}

func NewRemoteIgnoreRange(from, to int, lang string) RemoteIgnoreRange {
	return RemoteIgnoreRange{From: from, To: to, Lang: lang}
}

func (r RemoteIgnoreRange) GetFrom() int            { return r.From }
func (r *RemoteIgnoreRange) SetFrom(from int)       { r.From = from }
func (r RemoteIgnoreRange) GetTo() int              { return r.To }
func (r *RemoteIgnoreRange) SetTo(to int)           { r.To = to }
func (r RemoteIgnoreRange) GetLanguageCode() string { return r.Lang }
func (r *RemoteIgnoreRange) SetLanguageCode(code string) {
	r.Lang = code
}
