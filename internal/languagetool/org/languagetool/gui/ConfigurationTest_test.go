package gui

// Twin of ConfigurationTest — properties save/load with per-language sections.
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of ConfigurationTest.testSaveAndLoadConfiguration
func TestConfiguration_SaveAndLoadConfiguration(t *testing.T) {
	dir := t.TempDir()
	name := "test.cfg"
	conf, err := NewConfiguration(dir, name, "")
	require.NoError(t, err)
	conf.SetDisabledRuleIDs([]string{"FOO1", "Foo2"})
	conf.SetEnabledRuleIDs([]string{"enabledRule"})
	require.NoError(t, conf.SaveConfiguration())

	loaded, err := NewConfiguration(dir, name, "")
	require.NoError(t, err)
	dis := loaded.GetDisabledRuleIDs()
	require.Contains(t, dis, "FOO1")
	require.Contains(t, dis, "Foo2")
	require.Len(t, dis, 2)
	en := loaded.GetEnabledRuleIDs()
	require.Contains(t, en, "enabledRule")
	require.Len(t, en, 1)
}

// Port of ConfigurationTest.testSaveAndLoadConfigurationForManyLanguages
func TestConfiguration_SaveAndLoadConfigurationForManyLanguages(t *testing.T) {
	dir := t.TempDir()
	name := "multi.cfg"

	enConf, err := NewConfiguration(dir, name, "en-US")
	require.NoError(t, err)
	enConf.SetDisabledRuleIDs([]string{"FOO1", "Foo2"})
	enConf.SetEnabledRuleIDs([]string{"enabledRule"})
	require.NoError(t, enConf.SaveConfiguration())

	// switch language — empty for FR
	frConf, err := NewConfiguration(dir, name, "fr")
	require.NoError(t, err)
	require.Empty(t, frConf.GetDisabledRuleIDs())
	require.Empty(t, frConf.GetEnabledRuleIDs())
	frConf.SetEnabledRuleIDs([]string{"enabledFRRule"})
	require.NoError(t, frConf.SaveConfiguration())

	// back to EN — previous settings preserved
	en2, err := NewConfiguration(dir, name, "en-US")
	require.NoError(t, err)
	require.Contains(t, en2.GetDisabledRuleIDs(), "FOO1")
	require.Contains(t, en2.GetDisabledRuleIDs(), "Foo2")
	require.Len(t, en2.GetDisabledRuleIDs(), 2)
	require.Contains(t, en2.GetEnabledRuleIDs(), "enabledRule")
	require.Len(t, en2.GetEnabledRuleIDs(), 1)

	// FR still has its enabled rule
	fr2, err := NewConfiguration(dir, name, "fr")
	require.NoError(t, err)
	require.Contains(t, fr2.GetEnabledRuleIDs(), "enabledFRRule")

	// file exists
	_, err = os.Stat(filepath.Join(dir, name))
	require.NoError(t, err)
}
