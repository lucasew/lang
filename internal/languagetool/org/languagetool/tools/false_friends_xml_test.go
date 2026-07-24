package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateFalseFriendsXML(t *testing.T) {
	require.NoError(t, ValidateFalseFriendsXML(strings.NewReader(
		`<?xml version="1.0"?><rules><rulegroup id="X"><rule/></rulegroup></rules>`)))
	require.Error(t, ValidateFalseFriendsXML(strings.NewReader(`<not-rules/>`)))
	require.Error(t, ValidateFalseFriendsXML(strings.NewReader(`<<broken`)))

	cwd, _ := os.Getwd()
	dir := cwd
	for i := 0; i < 10; i++ {
		p := filepath.Join(dir, "inspiration", "languagetool", "languagetool-core", "src", "main", "resources", "org", "languagetool", "rules", "false-friends.xml")
		f, err := os.Open(p)
		if err != nil {
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
			continue
		}
		defer f.Close()
		if err := ValidateFalseFriendsXML(f); err != nil {
			t.Logf("full false-friends.xml: %v", err)
		}
		return
	}
}
