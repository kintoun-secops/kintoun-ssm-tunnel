package ui

import "strings"

type Preset struct {
	Match  string
	Remote int
	Scheme string
	Path   string
}

// presets follow the ports in the kintoun-infra runbooks. The local port is picked at
// connect time, so presets only fix the remote port and how to build the URL.
var presets = []Preset{
	{Match: "kali", Remote: 6080, Scheme: "http", Path: "/vnc.html"},
	{Match: "wazuh", Remote: 443, Scheme: "https"},
	{Match: "velociraptor", Remote: 8889, Scheme: "https"},
}

func findPreset(name string) (Preset, bool) {
	lower := strings.ToLower(name)
	for _, p := range presets {
		if strings.Contains(lower, p.Match) {
			return p, true
		}
	}
	return Preset{}, false
}
