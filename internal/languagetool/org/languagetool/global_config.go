package languagetool

// GlobalConfig ports org.languagetool.GlobalConfig.
type GlobalConfig struct {
	GrammalecteServer   string
	GrammalecteUser     string
	GrammalectePassword string
	BeolingusFile       string // path
	NerURL              string
}

var globalVerbose bool

func IsVerbose() bool         { return globalVerbose }
func SetVerbose(verbose bool) { globalVerbose = verbose }

func (c *GlobalConfig) SetGrammalecteServer(u string)   { c.GrammalecteServer = u }
func (c *GlobalConfig) SetGrammalecteUser(u string)     { c.GrammalecteUser = u }
func (c *GlobalConfig) SetGrammalectePassword(p string) { c.GrammalectePassword = p }
func (c *GlobalConfig) SetBeolingusFile(path string)    { c.BeolingusFile = path }
func (c *GlobalConfig) SetNERUrl(u string)              { c.NerURL = u }

func (c *GlobalConfig) GetGrammalecteServer() string   { return c.GrammalecteServer }
func (c *GlobalConfig) GetGrammalecteUser() string     { return c.GrammalecteUser }
func (c *GlobalConfig) GetGrammalectePassword() string { return c.GrammalectePassword }
func (c *GlobalConfig) GetBeolingusFile() string       { return c.BeolingusFile }
func (c *GlobalConfig) GetNerUrl() string              { return c.NerURL }

func (c *GlobalConfig) Equal(o *GlobalConfig) bool {
	if c == o {
		return true
	}
	if c == nil || o == nil {
		return false
	}
	return c.GrammalecteServer == o.GrammalecteServer &&
		c.GrammalecteUser == o.GrammalecteUser &&
		c.GrammalectePassword == o.GrammalectePassword &&
		c.BeolingusFile == o.BeolingusFile &&
		c.NerURL == o.NerURL
}
