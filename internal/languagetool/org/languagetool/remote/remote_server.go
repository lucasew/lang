package remote

// RemoteServer ports org.languagetool.remote.RemoteServer.
type RemoteServer struct {
	Software  string
	Version   string
	BuildDate string // optional; empty when unknown
}

func NewRemoteServer(software, version string) RemoteServer {
	return RemoteServer{Software: software, Version: version}
}

func NewRemoteServerFull(software, version, buildDate string) RemoteServer {
	return RemoteServer{Software: software, Version: version, BuildDate: buildDate}
}

func (s RemoteServer) GetSoftware() string  { return s.Software }
func (s RemoteServer) GetVersion() string   { return s.Version }
func (s RemoteServer) GetBuildDate() string { return s.BuildDate }

func (s RemoteServer) String() string {
	return s.Software + "/" + s.Version + "/" + s.BuildDate
}
