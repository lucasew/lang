package server

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	APIV2DocURL      = "https://languagetool.org/http-api/swagger-ui/#!/default"
	JSONContentType  = "application/json"
	TextContentType  = "text/plain"
)

// ApiV2 ports org.languagetool.server.ApiV2 request routing (without net/http wire-up).
type ApiV2 struct {
	Config         *HTTPServerConfig
	AllowOriginURL string
	TextChecker    *V2TextChecker
	// Languages is a pluggable list of short codes for /languages.
	Languages []LanguageInfo
}

func NewApiV2(cfg *HTTPServerConfig, languages []LanguageInfo) *ApiV2 {
	if cfg == nil {
		cfg = NewHTTPServerConfig()
	}
	if languages == nil {
		languages = DefaultCoreLanguages()
	}
	return &ApiV2{
		Config:         cfg,
		AllowOriginURL: cfg.AllowOriginURL,
		TextChecker:    NewV2TextChecker(cfg, false, NewRequestCounter()),
		Languages:      languages,
	}
}

// HandleResult is a wire-free response from ApiV2.Handle.
type HandleResult struct {
	Status      int
	ContentType string
	Body        string
}

// Handle routes a v2 path (e.g. "check", "languages", "info") with query parameters.
func (a *ApiV2) Handle(path string, parameters map[string]string) (HandleResult, error) {
	if a == nil {
		return HandleResult{}, NewUnavailableError("api not initialized", nil)
	}
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimPrefix(path, "v2/")
	Metrics().LogRequest()

	switch path {
	case "languages":
		body, err := a.GetLanguagesJSON()
		if err != nil {
			return HandleResult{}, err
		}
		Metrics().LogResponse(200)
		return HandleResult{Status: 200, ContentType: JSONContentType, Body: body}, nil
	case "maxtextlength":
		body := fmt.Sprintf("%d", a.Config.MaxTextLengthAnonymous)
		Metrics().LogResponse(200)
		return HandleResult{Status: 200, ContentType: TextContentType, Body: body}, nil
	case "info", "software":
		body, err := a.GetSoftwareInfoJSON()
		if err != nil {
			return HandleResult{}, err
		}
		Metrics().LogResponse(200)
		return HandleResult{Status: 200, ContentType: JSONContentType, Body: body}, nil
	case "check":
		if err := a.TextChecker.CheckParams(parameters); err != nil {
			Metrics().LogRequestError(RequestErrorInvalidRequest)
			return HandleResult{}, err
		}
		// validate callback early (JSONP)
		callback := parameters["callback"]
		if callback != "" {
			qp, err := ParseCheckQueryParams(parameters)
			if err != nil {
				return HandleResult{}, err
			}
			_ = qp
		}
		text := parameters["text"]
		if text == "" {
			text = parameters["data"]
		}
		limits := DefaultUserLimits(a.Config)
		if err := a.TextChecker.ValidateTextLength(text, limits); err != nil {
			return HandleResult{}, err
		}
		lang := parameters["language"]
		if lang == "" {
			lang = "auto"
		}
		// Soft language-id: heuristic when auto; otherwise use requested code.
		if strings.EqualFold(lang, "auto") {
			preferred := commaSeparated(parameters["preferredVariants"])
			if err := ValidatePreferredVariants(preferred, nil); err != nil {
				return HandleResult{}, err
			}
			lang = DetectLanguageOfString(text, preferred, nil)
			if lang == "" {
				lang = "en"
			}
		}
		opts := CheckOptions{
			Disabled:       a.TextChecker.GetDisabledRuleIDs(parameters),
			Enabled:        a.TextChecker.GetEnabledRuleIDs(parameters),
			UseEnabledOnly: strings.EqualFold(parameters["enabledOnly"], "true"),
		}
		if v := parameters["mode"]; v != "" {
			opts.Mode = CheckMode(strings.ToUpper(v))
		}
		if v := parameters["level"]; v != "" {
			opts.Level = CheckLevel(strings.ToUpper(v))
		}
		body, err := a.TextChecker.CheckAndBuildJSONWithOptions(text, lang, lang, opts)
		if err != nil {
			return HandleResult{}, err
		}
		ct := JSONContentType
		if callback != "" {
			// JSONP: callbackName({...});
			body = callback + "(" + body + ");"
			ct = "application/javascript"
		}
		Metrics().LogResponse(200)
		return HandleResult{Status: 200, ContentType: ct, Body: body}, nil
	default:
		return HandleResult{}, NewPathNotFoundError("Unsupported action: '" + path + "'. Please see " + APIV2DocURL)
	}
}

func (a *ApiV2) GetLanguagesJSON() (string, error) {
	langs := a.Languages
	if langs == nil {
		langs = []LanguageInfo{}
	}
	b, err := json.Marshal(langs)
	return string(b), err
}

func (a *ApiV2) GetSoftwareInfoJSON() (string, error) {
	info := map[string]any{
		"software": NewSoftwareInfo("dev"),
	}
	b, err := json.Marshal(info)
	return string(b), err
}
