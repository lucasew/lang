package server

// GRPCServer ports org.languagetool.server.GRPCServer process/analyze surface
// without generated protobuf stubs (wire deferred).
type GRPCServer struct {
	Pool       *PipelinePool
	UserConfig string // opaque user config key
	GlobalKey  string
}

func NewGRPCServer() *GRPCServer {
	return &GRPCServer{}
}

// InitPool attaches a pipeline pool from HTTP config.
func (g *GRPCServer) InitPool(cfg *HTTPServerConfig) {
	if g == nil {
		return
	}
	g.Pool = NewPipelinePool(cfg)
}

// ProcessingOptions is a proto-free stand-in for MLServerProto.ProcessingOptions.
type ProcessingOptions struct {
	Language     string
	EnabledRules []string
	DisabledRules []string
	EnabledOnly  bool
	Premium      bool
	TempOff      bool
	Level        string
}

// BuildSettings maps processing options to a PipelineSettings key.
func (g *GRPCServer) BuildSettings(opt ProcessingOptions) PipelineSettings {
	q := QueryParams{
		EnabledRules:     opt.EnabledRules,
		DisabledRules:    opt.DisabledRules,
		UseEnabledOnly:   opt.EnabledOnly,
		UseQuerySettings: true,
		Premium:          opt.Premium,
		EnableTempOffRules: opt.TempOff,
		LanguageCode:     opt.Language,
	}
	return NewPipelineSettingsFull(opt.Language, "", q, g.GlobalKey, g.UserConfig)
}

// Analyze borrows a pipeline for the language and returns whether it succeeded.
// Full sentence analysis is deferred to CheckEngine wiring.
func (g *GRPCServer) Analyze(opt ProcessingOptions, text string) (lang string, tokenCount int, err error) {
	if g == nil || g.Pool == nil {
		return "", 0, NewUnavailableError("pool not initialized", nil)
	}
	settings := g.BuildSettings(opt)
	pl, err := g.Pool.Borrow(settings)
	if err != nil {
		return "", 0, err
	}
	defer g.Pool.Return(settings, pl)
	// Placeholder: count whitespace-separated tokens
	n := 0
	for _, f := range splitFields(text) {
		if f != "" {
			n++
		}
	}
	return opt.Language, n, nil
}

func splitFields(s string) []string {
	var out []string
	start := -1
	for i, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if start >= 0 {
				out = append(out, s[start:i])
				start = -1
			}
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		out = append(out, s[start:])
	}
	return out
}
