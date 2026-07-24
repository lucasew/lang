package languagetool

import (
	"strings"
	"time"
)

// LtBuildInfo ports org.languagetool.LtBuildInfo as loadable git property snapshots.
// OS / PREMIUM variants load from property maps (or stay empty when unavailable).
// Java is an enum with OS("/git.properties") and PREMIUM("/git-premium.properties").
type LtBuildInfo struct {
	Name       string // "OS" or "PREMIUM"
	BuildDate  *string
	ShortGitID *string
	Version    *string
}

// OSBuildInfo is the process-wide OS LtBuildInfo snapshot (Java LtBuildInfo.OS).
// Empty until git.properties are loaded via LoadLtBuildInfo.
var OSBuildInfo = LoadLtBuildInfo("OS", nil)

// LoadLtBuildInfo parses git.properties-style keys from props.
// When props is nil (resource missing), all fields stay null like Java.
func LoadLtBuildInfo(name string, props map[string]string) LtBuildInfo {
	info := LtBuildInfo{Name: name}
	if props == nil {
		return info
	}
	// Java parses git.build.time with pattern "yyyy-MM-dd'T'HH:mm:ssXX"
	// and reformats to "yyyy-MM-dd HH:mm:ss Z".
	if v, ok := props["git.build.time"]; ok && v != "" {
		s := formatGitBuildTime(v)
		info.BuildDate = &s
	}
	if v, ok := props["git.commit.id.abbrev"]; ok {
		s := v
		info.ShortGitID = &s
	}
	if v, ok := props["git.build.version"]; ok {
		s := v
		info.Version = &s
	}
	return info
}

// formatGitBuildTime mirrors Java OffsetDateTime parse + reformat.
// Input pattern: yyyy-MM-dd'T'HH:mm:ssXX (e.g. 2024-01-01T12:00:00+0000 or Z)
// Output pattern: yyyy-MM-dd HH:mm:ss Z
func formatGitBuildTime(raw string) string {
	candidates := []string{
		"2006-01-02T15:04:05Z0700", // XX-style (+0000)
		time.RFC3339,               // with Z or +00:00
		"2006-01-02T15:04:05Z07:00",
	}
	for _, layout := range candidates {
		if t, err := time.Parse(layout, raw); err == nil {
			// Java DateTimeFormatter "Z" → +0000 style offset
			return t.Format("2006-01-02 15:04:05 -0700")
		}
	}
	if strings.HasSuffix(raw, "Z") {
		if t, err := time.Parse("2006-01-02T15:04:05Z", raw); err == nil {
			return t.UTC().Format("2006-01-02 15:04:05 -0700")
		}
	}
	// Keep raw if parse fails (defensive; Java would throw at enum init).
	return raw
}

func (i LtBuildInfo) GetBuildDate() *string  { return i.BuildDate }
func (i LtBuildInfo) GetShortGitId() *string { return i.ShortGitID }
func (i LtBuildInfo) GetVersion() *string    { return i.Version }
