package ui

import (
	"fmt"
	"strings"
)

type Preset struct {
	Match  string
	Remote int
	Local  int
	Scheme string
	Path   string
}

// presets follow the ports in the kintoun-infra runbooks.
var presets = []Preset{
	{Match: "kali", Remote: 6080, Local: 6080, Scheme: "http", Path: "/vnc.html"},
	{Match: "wazuh", Remote: 443, Local: 56789, Scheme: "https"},
	{Match: "velociraptor", Remote: 8889, Local: 8889, Scheme: "https"},
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

func (p Preset) URL() string {
	return fmt.Sprintf("%s://localhost:%d%s", p.Scheme, p.Local, p.Path)
}
