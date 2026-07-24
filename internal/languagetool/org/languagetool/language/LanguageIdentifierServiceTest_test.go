package language

// Twin of LanguageIdentifierServiceTest — factory surface.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier"
	"github.com/stretchr/testify/require"
)

func TestLanguageIdentifierService_Factory(t *testing.T) {
	svc := &identifier.LanguageIdentifierService{}
	d1 := svc.ClearLanguageIdentifier("default").GetDefaultLanguageIdentifier(0)
	d2 := svc.ClearLanguageIdentifier("default").GetDefaultLanguageIdentifier(1000)
	s1 := svc.ClearLanguageIdentifier("simple").GetSimpleLanguageIdentifier(nil)
	s2 := svc.ClearLanguageIdentifier("simple").GetSimpleLanguageIdentifier(nil)

	_, ok := d1.(*identifier.DefaultLanguageIdentifier)
	require.True(t, ok)
	_, ok = d2.(*identifier.DefaultLanguageIdentifier)
	require.True(t, ok)
	require.NotSame(t, d1, d2)

	_, ok = s1.(*identifier.SimpleLanguageIdentifier)
	require.True(t, ok)
	_, ok = s2.(*identifier.SimpleLanguageIdentifier)
	require.True(t, ok)
	require.NotSame(t, s1, s2)
}

func TestLanguageIdentifierService_FactoryWithoutReset(t *testing.T) {
	svc := &identifier.LanguageIdentifierService{}
	svc.Clear()
	d1 := svc.GetDefaultLanguageIdentifier(0)
	d2 := svc.GetDefaultLanguageIdentifier(1000)
	s1 := svc.GetSimpleLanguageIdentifier(nil)
	s2 := svc.GetSimpleLanguageIdentifier(nil)
	require.Same(t, d1, d2)
	require.Same(t, s1, s2)
	require.NotSame(t, s1, d1)
}

// also re-export Instance usage
func TestLanguageIdentifierService_Instance(t *testing.T) {
	identifier.Instance.Clear()
	require.Nil(t, identifier.Instance.GetInitialized())
	_ = identifier.Instance.GetSimpleLanguageIdentifier(nil)
	require.NotNil(t, identifier.Instance.GetInitialized())
	identifier.Instance.Clear()
}
