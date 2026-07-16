package languagetool

// LtBuildInfo ports org.languagetool.LtBuildInfo as loadable git property snapshots.
// OS / PREMIUM variants load from property maps (or stay empty when unavailable).
type LtBuildInfo struct {
	Name       string // "OS" or "PREMIUM"
	BuildDate  *string
	ShortGitID *string
	Version    *string
}

// LoadLtBuildInfo parses git.properties-style keys from props.
func LoadLtBuildInfo(name string, props map[string]string) LtBuildInfo {
	info := LtBuildInfo{Name: name}
	if props == nil {
		return info
	}
	if v, ok := props["git.build.time"]; ok && v != "" {
		// Java reformats ISO-ish timestamps; keep raw if parse fails.
		s := v
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

func (i LtBuildInfo) GetBuildDate() *string  { return i.BuildDate }
func (i LtBuildInfo) GetShortGitId() *string { return i.ShortGitID }
func (i LtBuildInfo) GetVersion() *string    { return i.Version }
