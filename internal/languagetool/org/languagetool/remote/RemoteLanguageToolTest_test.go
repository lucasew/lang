package remote

// Twin of RemoteLanguageToolTest.testResultParsing
import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func remoteFixture(t *testing.T, name string) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	dir := filepath.Dir(file)
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "inspiration/languagetool/languagetool-http-client/src/test/resources/org/languagetool/remote", name)
		if b, err := os.ReadFile(cand); err == nil {
			return b
		}
		dir = filepath.Dir(dir)
	}
	t.Fatalf("fixture not found: %s", name)
	return nil
}

func TestRemoteLanguageTool_ResultParsing(t *testing.T) {
	res, err := ParseCheckJSON(remoteFixture(t, "response.json"))
	require.NoError(t, err)
	require.Equal(t, "English (US)", res.GetLanguage())
	require.Equal(t, "en-US", res.GetLanguageCode())
	require.Equal(t, "LanguageTool", res.GetRemoteServer().GetSoftware())
	require.Equal(t, "3.4-SNAPSHOT", res.GetRemoteServer().GetVersion())
	require.Equal(t, "2016-05-27 12:04", res.GetRemoteServer().GetBuildDate())
	require.Len(t, res.GetMatches(), 1)
	m := res.GetMatches()[0]
	require.Equal(t, "EN_A_VS_AN", m.GetRuleID())
	require.Contains(t, m.GetMessage(), "Use \"an\" instead of 'a'")
	_, ok := m.GetRuleSubID()
	require.False(t, ok)
	require.Equal(t, "It happened a hour ago.", m.GetContext())
	require.Equal(t, 12, m.GetContextOffset())
	require.Equal(t, 1, m.GetErrorLength())
	require.Equal(t, 12, m.GetOffset())
	reps, okReps := m.GetReplacements()
	require.True(t, okReps)
	require.Equal(t, []string{"an"}, reps)
	require.Equal(t, "Miscellaneous", m.Category)
	require.Equal(t, "MISC", m.CategoryID)
	require.Equal(t, "misspelling", m.LocQualityIssueType)
	sm, ok := m.GetShortMessage()
	require.True(t, ok)
	require.Equal(t, "Wrong article", sm)
	require.Empty(t, m.URL)

	res2, err := ParseCheckJSON(remoteFixture(t, "response-with-url.json"))
	require.NoError(t, err)
	require.Equal(t, "https://fake.org/foo", res2.GetMatches()[0].URL)
}
