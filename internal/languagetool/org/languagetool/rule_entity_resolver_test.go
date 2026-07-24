package languagetool

import (
	"io"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/broker"
	"github.com/stretchr/testify/require"
)

func TestRuleEntityResolver(t *testing.T) {
	b := broker.NewMapResourceDataBroker()
	b.Resource["en/entities.ent"] = "<!ENTITY foo 'bar'>"
	r := NewRuleEntityResolver(b)
	require.Equal(t, "en/entities.ent", r.GetPathFromLTResourceFolder("file:///x/resource/en/entities.ent"))
	rc, err := r.ResolveEntity("file:///x/resource/en/entities.ent")
	require.NoError(t, err)
	require.NotNil(t, rc)
	data, _ := io.ReadAll(rc)
	_ = rc.Close()
	require.Contains(t, string(data), "foo")
	rc2, err := r.ResolveEntity("something.xml")
	require.NoError(t, err)
	require.Nil(t, rc2)
}
