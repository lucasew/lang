package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.tools.ContextToolsTest.

func TestContextTools_GetContext(t *testing.T) {
	ct := NewContextTools()
	context := ct.GetContext(4, 8, "Hi, this is some nice text waiting for its error markers.")
	require.Equal(t, `Hi, <b><font bgcolor="#ff8b8b">this</font></b> is some nice text waiting for its error...`, context)
	context2 := ct.GetContext(3, 5, "xxx\n \nyyy")
	require.Equal(t, `xxx<b><font bgcolor="#ff8b8b">&nbsp;&nbsp;</font></b> yyy`, context2)
}

func TestContextTools_PlainTextContext(t *testing.T) {
	ct := NewContextTools()
	ct.SetContextSize(5)
	input := "This is a test sentence. Here's another sentence with more text."
	result := ct.GetPlainTextContext(8, 14, input)
	require.Equal(t, "...s is a test sent...\n        ^^^^^^     ", result)
}

func TestContextTools_PlainTextContextWithLineBreaks(t *testing.T) {
	ct := NewContextTools()
	ct.SetContextSize(5)
	input := "One.\nThis is a test sentence.\nHere's another sentence."
	result := ct.GetPlainTextContext(15, 19, input)
	require.Equal(t, "...is a test sent...\n        ^^^^     ", result)
}

func TestContextTools_PlainTextContextWithDosLineBreaks(t *testing.T) {
	ct := NewContextTools()
	ct.SetContextSize(5)
	input := "One.\r\nThis is a test sentence.\r\nHere's another sentence."
	result := ct.GetPlainTextContext(16, 20, input)
	require.Equal(t, "...is a test sent...\n        ^^^^     ", result)
}

func TestContextTools_LargerContext(t *testing.T) {
	ct := NewContextTools()
	ct.SetContextSize(100)
	context := ct.GetContext(4, 8, "Hi, this is some nice text waiting for its error markers.")
	require.Equal(t, `Hi, <b><font bgcolor="#ff8b8b">this</font></b> is some nice text waiting for its error markers.`, context)
}

func TestContextTools_HtmlEscape(t *testing.T) {
	ct := NewContextTools()
	context1 := ct.GetContext(0, 2, "Hi, this is <html>.")
	require.Equal(t, `<b><font bgcolor="#ff8b8b">Hi</font></b>, this is &lt;html&gt;.`, context1)

	ct.SetEscapeHtml(false)
	context2 := ct.GetContext(0, 2, "Hi, this is <html>.")
	require.Equal(t, `<b><font bgcolor="#ff8b8b">Hi</font></b>, this is <html>.`, context2)
}

func TestContextTools_Markers(t *testing.T) {
	ct := NewContextTools()
	ct.SetErrorMarker("<X>", "</X>")
	context := ct.GetContext(0, 2, "Hi, this is it.")
	require.Equal(t, "<X>Hi</X>, this is it.", context)
}
