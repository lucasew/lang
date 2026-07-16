package remote

import "fmt"

// RemoteResult ports org.languagetool.remote.RemoteResult.
type RemoteResult struct {
	Language             string
	LanguageCode         string
	LanguageDetectedCode string
	LanguageDetectedName string
	Matches              []*RemoteRuleMatch
	IgnoreRanges         []RemoteIgnoreRange
	Server               RemoteServer
}

func NewRemoteResult(language, languageCode string, matches []*RemoteRuleMatch, server RemoteServer) *RemoteResult {
	if language == "" || languageCode == "" {
		panic("language and languageCode required")
	}
	return &RemoteResult{
		Language:     language,
		LanguageCode: languageCode,
		Matches:      append([]*RemoteRuleMatch(nil), matches...),
		Server:       server,
	}
}

func (r *RemoteResult) GetMatches() []*RemoteRuleMatch {
	return append([]*RemoteRuleMatch(nil), r.Matches...)
}
func (r *RemoteResult) GetLanguage() string     { return r.Language }
func (r *RemoteResult) GetLanguageCode() string { return r.LanguageCode }
func (r *RemoteResult) GetRemoteServer() RemoteServer { return r.Server }
func (r *RemoteResult) GetLanguageDetectedCode() string { return r.LanguageDetectedCode }
func (r *RemoteResult) GetLanguageDetectedName() string { return r.LanguageDetectedName }
func (r *RemoteResult) GetIgnoreRanges() []RemoteIgnoreRange {
	return append([]RemoteIgnoreRange(nil), r.IgnoreRanges...)
}
func (r *RemoteResult) String() string {
	if r == nil {
		return ""
	}
	return fmt.Sprint(r.Matches)
}
