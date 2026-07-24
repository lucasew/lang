package remote

import "fmt"

// RemoteResult ports org.languagetool.remote.RemoteResult.
// RemoteSentenceRange is one sentence span from /v2/check sentenceRanges.
type RemoteSentenceRange struct {
	Offset int
	Length int
}

type RemoteResult struct {
	Language             string
	LanguageCode         string
	LanguageDetectedCode string
	LanguageDetectedName string
	Matches              []*RemoteRuleMatch
	IgnoreRanges         []RemoteIgnoreRange
	SentenceRanges       []RemoteSentenceRange
	Server               RemoteServer
}

func NewRemoteResult(language, languageCode string, matches []*RemoteRuleMatch, server RemoteServer) *RemoteResult {
	return NewRemoteResultDetected(language, languageCode, "", "", matches, nil, server)
}

// NewRemoteResultDetected ports the full RemoteResult constructor with detected language fields.
func NewRemoteResultDetected(
	language, languageCode, detectedCode, detectedName string,
	matches []*RemoteRuleMatch,
	ignore []RemoteIgnoreRange,
	server RemoteServer,
) *RemoteResult {
	if language == "" || languageCode == "" {
		panic("language and languageCode required")
	}
	return &RemoteResult{
		Language:             language,
		LanguageCode:         languageCode,
		LanguageDetectedCode: detectedCode,
		LanguageDetectedName: detectedName,
		Matches:              append([]*RemoteRuleMatch(nil), matches...),
		IgnoreRanges:         append([]RemoteIgnoreRange(nil), ignore...),
		Server:               server,
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
func (r *RemoteResult) GetSentenceRanges() []RemoteSentenceRange {
	if r == nil {
		return nil
	}
	return append([]RemoteSentenceRange(nil), r.SentenceRanges...)
}
func (r *RemoteResult) String() string {
	if r == nil {
		return ""
	}
	return fmt.Sprint(r.Matches)
}
