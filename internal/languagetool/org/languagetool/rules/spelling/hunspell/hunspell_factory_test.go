package hunspell

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDictionary_CachesSamePath(t *testing.T) {
	t.Cleanup(ClearHunspellCaches)
	p := DiscoverHunspellDic("/da/hunspell/da_DK.dic")
	if p == "" {
		t.Skip("da_DK.dic not in tree")
	}
	aff := companionAff(p)
	d1 := GetDictionary(p, aff)
	d2 := GetDictionary(p, aff)
	require.Same(t, d1, d2, "path cache must return same instance")
	require.False(t, d1.IsClosed())
	require.False(t, d1.Spell("xyzzyqqqnotaword"))
}

func TestForDictionaryInResources_Danish(t *testing.T) {
	t.Cleanup(ClearHunspellCaches)
	if DiscoverHunspellDic("/da/hunspell/da_DK.dic") == "" {
		t.Skip("da_DK.dic not in tree")
	}
	d := ForDictionaryInResources("da_DK", "/da/hunspell/da_DK.dic", "/da/hunspell/da_DK.aff")
	require.NotNil(t, d)
	require.False(t, d.Spell("xyzzyqqqnotaword"))

	// second call hits resource cache
	d2 := ForDictionaryInResources("da_DK", "/da/hunspell/da_DK.dic", "/da/hunspell/da_DK.aff")
	require.Same(t, d, d2)
}

func TestForDictionaryInResources_MissingPanics(t *testing.T) {
	t.Cleanup(ClearHunspellCaches)
	require.Panics(t, func() {
		ForDictionaryInResources("xx", "/xx/hunspell/xx_XX.dic", "/xx/hunspell/xx_XX.aff")
	})
}

func TestCreateFromStreams_TempFiles(t *testing.T) {
	t.Cleanup(ClearHunspellCaches)
	dic := strings.NewReader("2\nhello\nworld\n")
	aff := strings.NewReader("SET UTF-8\n")
	d, err := defaultFactory{}.CreateFromStreams("en", dic, aff)
	require.NoError(t, err)
	require.True(t, d.Spell("hello"))
	require.True(t, d.Spell("world"))
	require.False(t, d.Spell("xyzzy"))
}

func TestSetHunspellStreamFactory(t *testing.T) {
	t.Cleanup(func() {
		SetHunspellStreamFactory(nil)
		ClearHunspellCaches()
	})
	custom := &mapFactory{words: []string{"customword"}}
	SetHunspellStreamFactory(custom)
	dir := t.TempDir()
	dic := filepath.Join(dir, "x.dic")
	aff := filepath.Join(dir, "x.aff")
	require.NoError(t, os.WriteFile(dic, []byte("1\na\n"), 0o644))
	require.NoError(t, os.WriteFile(aff, []byte("SET UTF-8\n"), 0o644))
	ClearHunspellCaches()
	d := GetDictionary(dic, aff)
	require.True(t, d.Spell("customword"))
	require.False(t, d.Spell("a"), "custom factory ignores file contents")
}

// mapFactory returns a fixed MapHunspellDictionary (tests SetHunspellStreamFactory).
type mapFactory struct{ words []string }

func (m *mapFactory) CreateFromLocalFiles(languageCode, dictionary, affix string) (HunspellDictionary, error) {
	return NewMapHunspellDictionary(m.words), nil
}
func (m *mapFactory) CreateFromStreams(languageCode string, dictionary, affix io.Reader) (HunspellDictionary, error) {
	return NewMapHunspellDictionary(m.words), nil
}

// TryOpenFromClasspath still works and uses path cache.
func TestTryOpenFromClasspath_UsesCache(t *testing.T) {
	t.Cleanup(ClearHunspellCaches)
	d1 := TryOpenFromClasspath("/da/hunspell/da_DK.dic")
	if d1 == nil {
		t.Skip("da_DK.dic not openable")
	}
	d2 := TryOpenFromClasspath("/da/hunspell/da_DK.dic")
	require.Same(t, d1, d2)
}
