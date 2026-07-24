package server

// HTTPSServerConfig ports org.languagetool.server.HTTPSServerConfig.
type HTTPSServerConfig struct {
	*HTTPServerConfig
	KeystorePath     string
	KeyStorePassword string
}

func NewHTTPSServerConfig(keystorePath, password string) *HTTPSServerConfig {
	base := NewHTTPServerConfig()
	return &HTTPSServerConfig{
		HTTPServerConfig: base,
		KeystorePath:     keystorePath,
		KeyStorePassword: password,
	}
}

func NewHTTPSServerConfigPort(port int, verbose bool, keystorePath, password string) *HTTPSServerConfig {
	base := NewHTTPServerConfigPortVerbose(port, verbose)
	return &HTTPSServerConfig{
		HTTPServerConfig: base,
		KeystorePath:     keystorePath,
		KeyStorePassword: password,
	}
}

// ApplyHTTPSArgs reads --config is not fully implemented; sets keystore from map.
func (c *HTTPSServerConfig) ApplyKeystoreProps(props map[string]string) error {
	if c == nil {
		return NewIllegalConfigurationError("nil HTTPS config")
	}
	ks, ok := props["keystore"]
	if !ok || ks == "" {
		return NewIllegalConfigurationError("keystore property required")
	}
	pw, ok := props["password"]
	if !ok {
		return NewIllegalConfigurationError("password property required")
	}
	c.KeystorePath = ks
	c.KeyStorePassword = pw
	return nil
}

func (c *HTTPSServerConfig) GetKeystore() string {
	if c == nil {
		return ""
	}
	return c.KeystorePath
}

func (c *HTTPSServerConfig) GetKeyStorePassword() string {
	if c == nil {
		return ""
	}
	return c.KeyStorePassword
}
