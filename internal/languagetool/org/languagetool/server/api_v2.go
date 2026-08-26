package server

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const (
	APIV2DocURL     = "https://languagetool.org/http-api/swagger-ui/#!/default"
	JSONContentType = "application/json"
	TextContentType = "text/plain"
)

// ApiV2 ports org.languagetool.server.ApiV2 request routing (without net/http wire-up).
type ApiV2 struct {
	Config         *HTTPServerConfig
	AllowOriginURL string
	TextChecker    *V2TextChecker
	// Languages is a pluggable list of short codes for /languages.
	Languages []LanguageInfo
	// UserDict in-memory dictionary for /v2/words*.
	UserDict *UserDictionary
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
		UserDict:       NewUserDictionary(),
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
	case "metrics":
		body, err := a.GetMetricsJSON()
		if err != nil {
			return HandleResult{}, err
		}
		Metrics().LogResponse(200)
		return HandleResult{Status: 200, ContentType: JSONContentType, Body: body}, nil
	case "configinfo":
		body, err := a.GetConfigurationInfoJSON(parameters["language"])
		if err != nil {
			return HandleResult{}, err
		}
		Metrics().LogResponse(200)
		return HandleResult{Status: 200, ContentType: JSONContentType, Body: body}, nil
	case "words":
		body, err := a.handleWordsList(parameters)
		if err != nil {
			return HandleResult{}, err
		}
		Metrics().LogResponse(200)
		return HandleResult{Status: 200, ContentType: JSONContentType, Body: body}, nil
	case "words/add":
		body, err := a.handleWordsAdd(parameters)
		if err != nil {
			return HandleResult{}, err
		}
		Metrics().LogResponse(200)
		return HandleResult{Status: 200, ContentType: JSONContentType, Body: body}, nil
	case "words/delete":
		body, err := a.handleWordsDelete(parameters)
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
		// Parse query knobs (JSONP callback, category filters, …)
		qp, err := ParseCheckQueryParams(parameters)
		if err != nil {
			return HandleResult{}, err
		}
		callback := parameters["callback"]
		text := parameters["text"]
		dataJSON := parameters["data"]
		var annotated *markup.AnnotatedText
		if text == "" && dataJSON != "" {
			at, err := ParseDataAnnotation(dataJSON)
			if err != nil {
				return HandleResult{}, err
			}
			annotated = at
			text = at.GetPlainText()
		}
		limits := DefaultUserLimits(a.Config)
		if err := a.TextChecker.ValidateTextLength(text, limits); err != nil {
			return HandleResult{}, err
		}
		langParam := parameters["language"]
		if langParam == "" {
			langParam = "auto"
		}
		// Java V2TextChecker: preferredVariants / noopLanguages only with language=auto;
		// detectLanguageOfString always runs for nested language.detectedLanguage.
		if err := ValidateNoopLanguages(parameters); err != nil {
			return HandleResult{}, err
		}
		preferred, err := ParsePreferredVariants(parameters)
		if err != nil {
			return HandleResult{}, err
		}
		if err := ValidatePreferredVariants(preferred, nil); err != nil {
			return HandleResult{}, err
		}
		// Java TextChecker: noopLanguages/preferredLanguages split(",") into detectLanguageOfString
		noopLangs := ParseNoopLanguages(parameters)
		preferredLangs := ParsePreferredLanguages(parameters)
		// Java V2: forcePreferredLanguages = "true".equals(...)
		forcePreferred := ParseForcePreferredLanguages(parameters)
		// Java V2 getLanguage: always detectLanguageOfString via languageIdentifier;
		// given = detected when auto else parseLanguage(langParam)
		det, err := a.TextChecker.DetectLanguageOfStringForce(text, preferred, noopLangs, preferredLangs, forcePreferred)
		if err != nil {
			return HandleResult{}, err
		}
		if det.Code == "" {
			det.Code = "en"
		}
		givenCode := langParam
		if langParam == "auto" {
			givenCode = det.Code
		}
		lang := givenCode
		ignoreWords := commaSeparated(parameters["ignoreWords"])
		// Merge in-memory user dictionary for username (or anon) — Java premium DB path.
		if a.UserDict != nil {
			ignoreWords = append(ignoreWords, a.UserDict.All(parameters["username"])...)
		}
		opts := CheckOptions{
			Disabled:       a.TextChecker.GetDisabledRuleIDs(parameters),
			Enabled:        a.TextChecker.GetEnabledRuleIDs(parameters),
			// Java: "yes".equals || "true".equals (via ParseCheckQueryParams)
			UseEnabledOnly: qp.UseEnabledOnly,
			MotherTongue:   parameters["motherTongue"],
			// ignoreWords CSV + user dictionary (Java user-config ignore list).
			IgnoreWords: ignoreWords,
			// category filters from disabledCategories / enabledCategories
			DisabledCategories: qp.DisabledCategories,
			EnabledCategories:  qp.EnabledCategories,
			RuleValues:         commaSeparated(parameters["ruleValues"]),
			// Java: QueryParams.altLanguages → Pipeline(lang, altLanguages, …)
			// Already validated in ParseCheckQueryParams when present.
			AltLanguages: append([]string(nil), qp.AltLanguages...),
			// Mode/Level/ToneTags from ParseCheckQueryParams (ServerTools + TextChecker)
			Mode:     qp.Mode,
			Level:    qp.Level,
			ToneTags: append([]string(nil), qp.ToneTags...),
		}
		langName := LanguageNameForCode(lang)
		// AllowIncompleteResults: Java returns partial matches + incompleteResultsReason
		// on ErrorRateTooHigh / TimeoutException — not invent size-threshold soft warnings.
		opts.AllowIncompleteResults = qp.AllowIncompleteResults
		// Java: future.get(limits.getMaxCheckTimeMillis(), …); -1 = unlimited
		opts.MaxCheckTimeMillis = limits.MaxCheckTimeMillis
		// Java PipelinePool: lt.setMaxErrorsPerWordRate(config.getMaxErrorsPerWordRate())
		if a.Config != nil && a.Config.MaxErrorsPerWordRate > 0 {
			opts.MaxErrorsPerWordRate = a.Config.MaxErrorsPerWordRate
		}
		// Multi-language: ignoreRanges from CheckResults (NewLanguageMatches).
		var ignoreRanges []IgnoreRangeInfo
		var incompleteReason string
		var body string
		checkStart := time.Now()
		if annotated != nil {
			var matches []RemoteRuleMatch
			matches, ignoreRanges, incompleteReason = a.TextChecker.CheckAnnotatedWithOptionsAndIgnore(annotated, lang, opts)
			matches = filterRemoteByIgnoreRanges(matches, ignoreRanges)
			body, err = a.TextChecker.BuildResponseExDetected(annotated.GetTextWithMarkup(), lang, langName, matches, det, incompleteReason, ignoreRanges, time.Since(checkStart).Milliseconds())
		} else {
			var matches []RemoteRuleMatch
			matches, ignoreRanges, incompleteReason = a.TextChecker.CheckWithOptionsAndIgnore(text, lang, opts)
			matches = filterRemoteByIgnoreRanges(matches, ignoreRanges)
			body, err = a.TextChecker.BuildResponseExDetected(text, lang, langName, matches, det, incompleteReason, ignoreRanges, time.Since(checkStart).Milliseconds())
		}
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

// GetMetricsJSON returns process-local ServerMetricsCollector snapshot.
func (a *ApiV2) GetMetricsJSON() (string, error) {
	snap := Metrics().Snapshot()
	b, err := json.Marshal(snap)
	return string(b), err
}

// GetConfigurationInfoJSON soft-ports /v2/configinfo for a language.
func (a *ApiV2) GetConfigurationInfoJSON(lang string) (string, error) {
	if tools.JavaStringTrimIsEmpty(lang) {
		return "", NewBadRequestError("'language' parameter missing")
	}
	lt := languagetool.NewJLanguageTool(lang)
	corepack.Register(lt, lang)
	ids := lt.GetAllRegisteredRuleIDs()
	type ruleInfo struct {
		RuleID              string `json:"ruleId"`
		Description         string `json:"description"`
		CategoryID          string `json:"categoryId"`
		CategoryName        string `json:"categoryName"`
		LocQualityIssueType string `json:"locQualityIssueType"`
	}
	rules := make([]ruleInfo, 0, len(ids))
	for _, id := range ids {
		catID, catName, issue, _ := RuleMeta(id)
		desc := RuleDescription(id)
		if desc == "" {
			desc = id
		}
		rules = append(rules, ruleInfo{
			RuleID:              id,
			Description:         desc,
			CategoryID:          catID,
			CategoryName:        catName,
			LocQualityIssueType: issue,
		})
	}
	maxLen := 0
	if a != nil && a.Config != nil {
		maxLen = a.Config.MaxTextLengthAnonymous
		if a.Config.MaxTextHardLength > 0 {
			maxLen = a.Config.MaxTextHardLength
		}
	}
	payload := map[string]any{
		"software": NewSoftwareInfo("dev"),
		"parameter": map[string]any{
			"maxTextLength": maxLen,
		},
		"rules": rules,
	}
	b, err := json.Marshal(payload)
	return string(b), err
}

func (a *ApiV2) handleWordsList(parameters map[string]string) (string, error) {
	if a.UserDict == nil {
		a.UserDict = NewUserDictionary()
	}
	offset, _ := strconv.Atoi(parameters["offset"])
	limit, _ := strconv.Atoi(parameters["limit"])
	if limit <= 0 {
		limit = 10
	}
	words := a.UserDict.List(parameters["username"], offset, limit)
	if words == nil {
		words = []string{}
	}
	b, err := json.Marshal(map[string]any{"words": words})
	return string(b), err
}

func (a *ApiV2) handleWordsAdd(parameters map[string]string) (string, error) {
	if a.UserDict == nil {
		a.UserDict = NewUserDictionary()
	}
	user := parameters["username"]
	// single word or batch: mode=batch&words="a b c"
	// Java: parameters.get("words").split("\\s+") — ASCII whitespace runs (not Fields).
	// Java ServerTools.getMode: "batch".equals(modeParam) — case-sensitive
	if parameters["mode"] == "batch" {
		added := 0
		for _, w := range javaSplitASCIIWhitespacePlus(parameters["words"]) {
			if w == "" {
				continue
			}
			if a.UserDict.Add(user, w) {
				added++
			}
		}
		b, err := json.Marshal(map[string]any{"added": true, "count": added})
		return string(b), err
	}
	word := parameters["word"]
	if word == "" {
		return "", NewBadRequestError("Missing 'word' parameter")
	}
	ok := a.UserDict.Add(user, word)
	b, err := json.Marshal(map[string]any{"added": ok})
	return string(b), err
}

func (a *ApiV2) handleWordsDelete(parameters map[string]string) (string, error) {
	if a.UserDict == nil {
		a.UserDict = NewUserDictionary()
	}
	user := parameters["username"]
	if parameters["mode"] == "batch" {
		deleted := 0
		for _, w := range javaSplitASCIIWhitespacePlus(parameters["words"]) {
			if w == "" {
				continue
			}
			if a.UserDict.Delete(user, w) {
				deleted++
			}
		}
		b, err := json.Marshal(map[string]any{"deleted": true, "count": deleted})
		return string(b), err
	}
	word := parameters["word"]
	if word == "" {
		return "", NewBadRequestError("Missing 'word' parameter")
	}
	ok := a.UserDict.Delete(user, word)
	b, err := json.Marshal(map[string]any{"deleted": ok})
	return string(b), err
}

// javaSplitASCIIWhitespacePlus ports Java String.split("\\s+") without UNICODE_CHARACTER_CLASS
// (ApiV2 batch words). Trailing empties dropped; mid empties not produced by \s+.
func javaSplitASCIIWhitespacePlus(s string) []string {
	if s == "" {
		return []string{""}
	}
	var parts []string
	start := 0
	i := 0
	for i < len(s) {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\f' || c == '\r' {
			if start < i {
				parts = append(parts, s[start:i])
			}
			j := i + 1
			for j < len(s) {
				c2 := s[j]
				if c2 != ' ' && c2 != '\t' && c2 != '\n' && c2 != '\v' && c2 != '\f' && c2 != '\r' {
					break
				}
				j++
			}
			start = j
			i = j
			continue
		}
		i++
	}
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	if len(parts) == 0 {
		return []string{""}
	}
	return parts
}
