package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUkIsGoodPosTag(t *testing.T) {
	require.True(t, ukIsGoodPosTag("verb:inf"))
	require.False(t, ukIsGoodPosTag(""))
	require.False(t, ukIsGoodPosTag(languagetool.SentenceEndTagName))
	require.False(t, ukIsGoodPosTag("adj:bad"))
	require.False(t, ukIsGoodPosTag("noun:subst"))
	require.False(t, ukIsGoodPosTag("<ignore>"))
}

func TestLoadDerivats_Перелиставши(t *testing.T) {
	// Official resource /uk/derivats.txt (resource dir) embedded as data/derivats.txt
	m := loadDerivats()
	require.NotEmpty(t, m)
	_, ok := m["перелиставши"]
	require.True(t, ok, "перелиставши should be in derivats")
}
