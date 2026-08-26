package languagetool

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/broker"
	"github.com/stretchr/testify/require"
)

func TestXMLValidator(t *testing.T) {
	v := NewXMLValidator()
	require.NoError(t, v.ValidateWellFormed(`<root><a/></root>`))
	require.Error(t, v.ValidateWellFormed(`<root><a></root>`))
}

func TestXMLValidator_DTDDecl(t *testing.T) {
	v := NewXMLValidator()
	xml := `<?xml version="1.0" encoding="UTF-8"?><rules></rules>`
	require.NoError(t, v.ValidateXMLString(xml, "/rules.dtd", "rules"))
	require.Error(t, v.ValidateXMLString(`<rules/>`, "/rules.dtd", "rules"))
}

func TestXMLErrorHandler(t *testing.T) {
	h := &XMLErrorHandler{OnMessage: func(string) {}}
	require.Error(t, h.Error("bad", 1, 2))
}

func TestEntityAsInput(t *testing.T) {
	b := broker.NewMapResourceDataBroker()
	b.Resource["en/x.ent"] = "ENTITY"
	r := NewRuleEntityResolver(b)
	e, err := NewEntityAsInput("pub", "file:///resource/en/x.ent", r)
	require.NoError(t, err)
	require.NotNil(t, e.GetByteStream())
	require.Equal(t, "file:///resource/en/x.ent", e.GetSystemId())
}
