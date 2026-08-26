package languagetool

// Twin of languagetool-core/src/test/java/org/languagetool/InputSentenceTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of languagetool-core/src/test/java/org/languagetool/InputSentenceTest.java :: InputSentenceTest.test
func TestInputSentence_Test(t *testing.T) {
	uc1 := NewUserConfig()
	uc1.UserSpecificSpellerWords = []string{"foo1", "foo2"}
	disabled := map[string]struct{}{"ID1": {}}
	disabledCat := map[string]struct{}{"C1": {}}
	enabled := map[string]struct{}{"ID2": {}}
	enabledCat := map[string]struct{}{"C2": {}}

	a1 := NewInputSentence(AnalyzePlain("foo"), "xx-XX", "xx-XX",
		disabled, disabledCat, enabled, enabledCat, uc1, nil, "ALL", LevelDefault, nil, nil)
	a2 := NewInputSentence(AnalyzePlain("foo"), "xx-XX", "xx-XX",
		disabled, disabledCat, enabled, enabledCat, uc1, nil, "ALL", LevelDefault, nil, nil)
	require.True(t, a1.Equal(a2))
	require.Equal(t, "foo", a1.String())

	// mother tongue differs
	noMT := NewInputSentence(AnalyzePlain("foo"), "xx-XX", "",
		disabled, disabledCat, enabled, enabledCat, uc1, nil, "ALL", LevelDefault, nil, nil)
	require.False(t, a1.Equal(noMT))

	// same word lists on a different UserConfig instance still equal (Equal ignores pointer identity)
	uc2 := NewUserConfig()
	uc2.UserSpecificSpellerWords = []string{"foo1", "foo2"}
	aUc2 := NewInputSentence(AnalyzePlain("foo"), "xx-XX", "xx-XX",
		disabled, disabledCat, enabled, enabledCat, uc2, nil, "ALL", LevelDefault, nil, nil)
	require.True(t, a1.Equal(aUc2))

	// mode differs
	otherMode := NewInputSentence(AnalyzePlain("foo"), "xx-XX", "xx-XX",
		disabled, disabledCat, enabled, enabledCat, uc1, nil, "TEXTLEVEL_ONLY", LevelDefault, nil, nil)
	require.False(t, a1.Equal(otherMode))

	// empty alt languages same as nil
	sameAlt := NewInputSentence(AnalyzePlain("foo"), "xx-XX", "xx-XX",
		disabled, disabledCat, enabled, enabledCat, uc1, []string{}, "ALL", LevelDefault, nil, nil)
	require.True(t, a1.Equal(sameAlt))

	// non-empty alt languages differ
	otherAlt := NewInputSentence(AnalyzePlain("foo"), "xx-XX", "xx-XX",
		disabled, disabledCat, enabled, enabledCat, uc1, []string{"xx"}, "ALL", LevelDefault, nil, nil)
	require.False(t, a1.Equal(otherAlt))
}
