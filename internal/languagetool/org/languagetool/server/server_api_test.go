package server

import (
	"encoding/base64"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServerLifecycleAndLimiters(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.RequestLimit = 10
	cfg.RequestLimitPeriodInSeconds = 60
	cfg.TimeoutRequestLimit = 5
	s := NewServer(cfg)
	require.False(t, s.IsRunning())
	s.Run()
	require.True(t, s.IsRunning())
	require.NotNil(t, s.RequestLimiter)
	require.NotNil(t, s.ErrorRequestLimiter)
	s.Stop()
	require.False(t, s.IsRunning())
	require.True(t, IsAllowedIP("127.0.0.1"))
	require.True(t, UsageRequested([]string{"--help"}))
}

func TestServerMetricsCollector(t *testing.T) {
	m := NewServerMetricsCollector()
	m.LogRequest()
	m.LogResponse(200)
	m.LogCheck("en", 100, 50, 2, "ALL")
	m.LogRequestError(RequestErrorMaxTextSize)
	require.Equal(t, int64(1), m.HTTPRequests())
	require.Equal(t, int64(1), m.Checks())
	require.Equal(t, int64(2), m.Matches())
	require.Equal(t, int64(50), m.Characters())
	require.Equal(t, int64(1), m.ResponseCount(200))
	require.Equal(t, int64(1), m.ErrorCount(RequestErrorMaxTextSize))
}

func TestRemoteRuleMatchAndResultExtender(t *testing.T) {
	m := NewRemoteRuleMatch("RULE1", "msg", "ctx error here", 4, 10, 5)
	m.Replacements = []string{"ok"}
	m.Category = "Grammar"
	m.CategoryID = "GRAMMAR"
	require.Equal(t, "RULE1@10-15", m.String())
	require.True(t, m.IsTouchedByOneOf([]Span{{From: 12, To: 20}}))
	require.False(t, m.IsTouchedByOneOf([]Span{{From: 20, To: 25}}))
	info := m.ToMatchInfo()
	require.Equal(t, "RULE1", info.Rule.ID)
	require.Equal(t, "ok", info.Replacements[0].Value)

	hidden := GetAsHiddenMatches(
		[]ExtensionMatch{{From: 0, To: 5}},
		[]ExtensionMatch{
			{From: 3, To: 8, CategoryID: "C", CategoryName: "Cat"}, // touches
			{From: 10, To: 15, CategoryID: "D", CategoryName: "Other", IssueType: "misspelling"},
		},
	)
	require.Len(t, hidden, 1)
	require.Equal(t, HiddenRuleID, hidden[0].RuleID)
	require.Equal(t, "misspelling", hidden[0].IssueType)
}

func TestTextCheckerAndV2(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.MaxTextLengthAnonymous = 20
	tc := NewV2TextChecker(cfg, false, nil)

	require.Error(t, tc.CheckParams(map[string]string{"language": "en"}))
	require.Error(t, tc.CheckParams(map[string]string{"language": "en", "text": "a", "data": "b"}))
	require.NoError(t, tc.CheckParams(map[string]string{"language": "en", "text": "hello"}))

	require.Error(t, tc.ValidateTextLength("this text is way too long for limit", DefaultUserLimits(cfg)))
	require.NoError(t, tc.ValidateTextLength("short", DefaultUserLimits(cfg)))

	p, err := ParseCheckQueryParams(map[string]string{
		"enabledRules": "A,B",
		"disabledRules": "C",
		"enabledOnly": "true",
		"mode": "all",
		"level": "picky",
		"callback": "cb",
	})
	require.NoError(t, err)
	require.Equal(t, []string{"A", "B"}, p.EnabledRules)
	require.True(t, p.UseEnabledOnly)
	require.Equal(t, CheckLevelPicky, p.Level)
	_, err = ParseCheckQueryParams(map[string]string{"callback": "bad-1"})
	require.Error(t, err)

	require.True(t, tc.GetLanguageAutoDetect(map[string]string{"language": "auto"}))
	body, err := tc.BuildResponse("hi", "en", "English", nil)
	require.NoError(t, err)
	require.Contains(t, body, "LanguageTool-Go")
	require.Contains(t, body, `"matches":[]`)
}

func TestApiV2Routes(t *testing.T) {
	cfg := NewHTTPServerConfig()
	cfg.MaxTextLengthAnonymous = 100
	api := NewApiV2(cfg, []LanguageInfo{{Name: "English", Code: "en"}})

	r, err := api.Handle("languages", nil)
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	require.Contains(t, r.Body, "English")

	r, err = api.Handle("maxtextlength", nil)
	require.NoError(t, err)
	require.Equal(t, "100", r.Body)

	r, err = api.Handle("info", nil)
	require.NoError(t, err)
	require.Contains(t, r.Body, "software")

	r, err = api.Handle("check", map[string]string{"language": "en", "text": "Hello"})
	require.NoError(t, err)
	require.Equal(t, 200, r.Status)
	require.Contains(t, r.Body, "matches")

	_, err = api.Handle("nope", nil)
	require.Error(t, err)
	var pe *PathNotFoundError
	require.ErrorAs(t, err, &pe)
}

func TestBasicAuthGroupRolesUserInfo(t *testing.T) {
	hdr := "Basic " + base64.StdEncoding.EncodeToString([]byte("alice:secret"))
	auth, err := ParseBasicAuthentication(hdr)
	require.NoError(t, err)
	require.Equal(t, "alice", auth.User)
	require.Equal(t, "secret", auth.Password)
	_, err = ParseBasicAuthentication("Bearer x")
	require.Error(t, err)

	enc := EncodeGroupRoles([]GroupRole{GroupRoleOwner, GroupRoleAdmin})
	require.Equal(t, "OWNER,ADMIN", enc)
	require.True(t, HasGroupPermissions(enc, GroupRoleAdmin))
	require.False(t, HasGroupPermissions(enc, GroupRoleEditor))

	u := NewUserInfoEntry(1, "a@b.c")
	require.False(t, u.HasPremium())
	from := time.Now().Add(-time.Hour)
	to := time.Now().Add(time.Hour)
	u.PremiumFrom = &from
	u.PremiumTo = &to
	require.True(t, u.HasPremium())

	gid := int64(9)
	d := NewDictGroupEntry(1, "dict", &gid)
	require.Equal(t, "dict", d.Name)
	require.Equal(t, int64(9), *d.UserGroupID)
}

func TestLocalAbTestService(t *testing.T) {
	s := NewLocalAbTestService([]string{"featA", "featB"}, regexp.MustCompile("web"))
	require.Nil(t, s.GetActiveAbTestForClient(map[string]string{"abtest": "featA", "useragent": "cli"}, nil))
	got := s.GetActiveAbTestForClient(map[string]string{"abtest": "featA,featX", "useragent": "web-1"}, nil)
	require.Equal(t, []string{"featA"}, got)
}
