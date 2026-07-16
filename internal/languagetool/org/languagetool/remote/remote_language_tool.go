package remote

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	v2Check         = "/v2/check"
	v2MaxTextLength = "/v2/maxtextlength"
	v2ConfigInfo    = "/v2/configinfo"
)

// HTTPDoer abstracts http.Client for tests.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// RemoteLanguageTool ports org.languagetool.remote.RemoteLanguageTool.
// HTTP POST to /v2/check; ParseCheckJSON is also usable offline.
type RemoteLanguageTool struct {
	ServerBaseURL string // must not end with /
	Client        HTTPDoer
}

// NewRemoteLanguageTool builds a client for serverBaseURL.
// Panics if the URL ends with '/', has an empty/unsupported scheme, or is not a valid URL.
// Supported schemes: http, https (Java MalformedURLException / RuntimeException parity).
func NewRemoteLanguageTool(serverBaseURL string) *RemoteLanguageTool {
	if strings.HasSuffix(serverBaseURL, "/") {
		panic("Server base URL must not end with '/': " + serverBaseURL)
	}
	u, err := url.Parse(serverBaseURL)
	if err != nil {
		panic("invalid server base URL: " + err.Error())
	}
	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "http", "https":
		// ok
	case "":
		panic("server base URL missing scheme: " + serverBaseURL)
	default:
		// ftp, htp typo, etc. — Java throws MalformedURLException / fails at connect
		panic("unsupported server URL scheme '" + scheme + "': " + serverBaseURL)
	}
	if u.Host == "" {
		panic("server base URL missing host: " + serverBaseURL)
	}
	return &RemoteLanguageTool{
		ServerBaseURL: serverBaseURL,
		Client:        http.DefaultClient,
	}
}

// Check runs a text check with a language code.
func (r *RemoteLanguageTool) Check(text, langCode string) (*RemoteResult, error) {
	cfg := NewCheckConfigurationBuilder(langCode).Build()
	return r.CheckWithConfig(text, cfg, nil)
}

// CheckWithConfig runs a check with full configuration.
func (r *RemoteLanguageTool) CheckWithConfig(text string, cfg *CheckConfiguration, custom map[string]string) (*RemoteResult, error) {
	params := GetURLParams(text, cfg, custom)
	return r.checkParams(params)
}

// GetURLParams builds application/x-www-form-urlencoded body fields.
func GetURLParams(text string, cfg *CheckConfiguration, custom map[string]string) url.Values {
	params := url.Values{}
	params.Set("text", text)
	if cfg == nil {
		params.Set("language", "auto")
		return params
	}
	if cfg.MotherTongueLangCode != "" {
		params.Set("motherTongue", cfg.MotherTongueLangCode)
	}
	if cfg.GuessLanguage {
		params.Set("language", "auto")
	} else if cfg.LangCode != "" {
		params.Set("language", cfg.LangCode)
	} else {
		params.Set("language", "auto")
	}
	if len(cfg.EnabledRuleIDs) > 0 {
		params.Set("enabledRules", strings.Join(cfg.EnabledRuleIDs, ","))
	}
	if cfg.EnabledOnly {
		params.Set("enabledOnly", "yes")
	}
	if len(cfg.DisabledRuleIDs) > 0 {
		params.Set("disabledRules", strings.Join(cfg.DisabledRuleIDs, ","))
	}
	if cfg.Mode != "" {
		params.Set("mode", cfg.Mode)
	}
	if cfg.Level != "" {
		params.Set("level", cfg.Level)
	}
	if len(cfg.RuleValues) > 0 {
		params.Set("ruleValues", strings.Join(cfg.RuleValues, ","))
	}
	if cfg.TextSessionID != "" {
		params.Set("textSessionId", cfg.TextSessionID)
	}
	if cfg.Username != "" {
		params.Set("username", cfg.Username)
	}
	if cfg.APIKey != "" {
		params.Set("apiKey", cfg.APIKey)
	}
	for k, v := range custom {
		params.Set(k, v)
	}
	return params
}

func (r *RemoteLanguageTool) checkParams(params url.Values) (*RemoteResult, error) {
	if r == nil {
		return nil, fmt.Errorf("nil RemoteLanguageTool")
	}
	endpoint := r.ServerBaseURL + v2Check
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := r.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("remote check HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	return ParseCheckJSON(body)
}

// apiJSON mirrors the public /v2/check response subset.
type apiJSON struct {
	Language struct {
		Name string `json:"name"`
		Code string `json:"code"`
		DetectedLanguage *struct {
			Name string `json:"name"`
			Code string `json:"code"`
		} `json:"detectedLanguage"`
	} `json:"language"`
	Matches []apiMatch `json:"matches"`
	Software struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"software"`
	IgnoreRanges []struct {
		From int    `json:"from"`
		To   int    `json:"to"`
		Lang string `json:"language"`
	} `json:"ignoreRanges"`
}

type apiMatch struct {
	Message      string `json:"message"`
	ShortMessage string `json:"shortMessage"`
	Offset       int    `json:"offset"`
	Length       int    `json:"length"`
	Context      struct {
		Text   string `json:"text"`
		Offset int    `json:"offset"`
		Length int    `json:"length"`
	} `json:"context"`
	Replacements []struct {
		Value string `json:"value"`
	} `json:"replacements"`
	Rule struct {
		ID          string `json:"id"`
		SubID       string `json:"subId"`
		Description string `json:"description"`
		URLs        []struct {
			Value string `json:"value"`
		} `json:"urls"`
		Category struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"category"`
		IssueType string `json:"issueType"`
	} `json:"rule"`
}

// ParseCheckJSON parses a /v2/check JSON response body.
func ParseCheckJSON(data []byte) (*RemoteResult, error) {
	var raw apiJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	server := NewRemoteServer(raw.Software.Name, raw.Software.Version)
	matches := make([]*RemoteRuleMatch, 0, len(raw.Matches))
	for _, m := range raw.Matches {
		ctx := m.Context.Text
		if ctx == "" {
			ctx = m.Message
		}
		rm := NewRemoteRuleMatch(m.Rule.ID, m.Rule.Description, m.Message, ctx, m.Context.Offset, m.Offset, m.Length)
		rm.SetShortMessage(m.ShortMessage)
		rm.SetSubID(m.Rule.SubID)
		if len(m.Rule.URLs) > 0 {
			rm.SetURL(m.Rule.URLs[0].Value)
		}
		rm.SetCategory(m.Rule.Category.Name, m.Rule.Category.ID)
		rm.SetLocQualityIssueType(m.Rule.IssueType)
		reps := make([]string, 0, len(m.Replacements))
		for _, r := range m.Replacements {
			reps = append(reps, r.Value)
		}
		rm.SetReplacements(reps)
		matches = append(matches, rm)
	}
	langName := raw.Language.Name
	langCode := raw.Language.Code
	if langName == "" {
		langName = langCode
	}
	if langCode == "" {
		langCode = "unknown"
	}
	if langName == "" {
		langName = langCode
	}
	res := NewRemoteResult(langName, langCode, matches, server)
	if raw.Language.DetectedLanguage != nil {
		res.LanguageDetectedCode = raw.Language.DetectedLanguage.Code
		res.LanguageDetectedName = raw.Language.DetectedLanguage.Name
	}
	for _, ir := range raw.IgnoreRanges {
		res.IgnoreRanges = append(res.IgnoreRanges, NewRemoteIgnoreRange(ir.From, ir.To, ir.Lang))
	}
	return res, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// GetConfigurationInfo fetches /v2/configinfo.
func (r *RemoteLanguageTool) GetConfigurationInfo() (*RemoteConfigurationInfo, error) {
	if r == nil {
		return nil, fmt.Errorf("nil RemoteLanguageTool")
	}
	endpoint := r.ServerBaseURL + v2ConfigInfo
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	client := r.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("configinfo HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	return ParseRemoteConfigurationInfo(resp.Body)
}

// GetMaxTextLength fetches /v2/maxtextlength (plain integer body).
func (r *RemoteLanguageTool) GetMaxTextLength() (int, error) {
	if r == nil {
		return 0, fmt.Errorf("nil RemoteLanguageTool")
	}
	endpoint := r.ServerBaseURL + v2MaxTextLength
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}
	client := r.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("maxtextlength HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}
	var n int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(body)), "%d", &n); err != nil {
		return 0, err
	}
	return n, nil
}
