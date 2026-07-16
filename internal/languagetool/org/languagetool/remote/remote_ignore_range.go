package remote

// RemoteIgnoreRange ports org.languagetool.remote.RemoteIgnoreRange.
type RemoteIgnoreRange struct {
	From  int
	To    int
	Lang  string
}

func NewRemoteIgnoreRange(from, to int, lang string) RemoteIgnoreRange {
	return RemoteIgnoreRange{From: from, To: to, Lang: lang}
}
