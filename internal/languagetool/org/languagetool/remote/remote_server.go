package remote

// RemoteServer ports org.languagetool.remote.RemoteServer.
type RemoteServer struct {
	Software string
	Version  string
}

func NewRemoteServer(software, version string) RemoteServer {
	return RemoteServer{Software: software, Version: version}
}

func (s RemoteServer) GetSoftware() string { return s.Software }
func (s RemoteServer) GetVersion() string  { return s.Version }
