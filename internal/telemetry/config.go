package telemetry

type TelemetryConf struct {
	URL    string `toml:"url"`
	DB     string `toml:"db"`
	Token  string `toml:"token"`
	Suffix string
}

type GrafanaConf struct {
	URL    string `toml:"url"`
	User   string `toml:"user"`
	Passwd string `toml:"password"`
}
