package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kintoun-secops/kintoun-ssm-tunnel/internal/aws"
)

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func update(m model, msg tea.Msg) model {
	next, _ := m.Update(msg)
	return next.(model)
}

func TestFindPreset(t *testing.T) {
	cases := map[string]string{
		"kintoun-secops-infra-kali-attacker-ec2": "http://localhost:6080/vnc.html",
		"kintoun-secops-infra-wazuh-ec2":         "https://localhost:56789",
		"kintoun-secops-infra-velociraptor-ec2":  "https://localhost:8889",
	}
	for name, url := range cases {
		p, ok := findPreset(name)
		if !ok || p.URL() != url {
			t.Errorf("%s: got %v %q, want %q", name, ok, p.URL(), url)
		}
	}
	if _, ok := findPreset("kintoun-other-ec2"); ok {
		t.Error("unknown instance must not match a preset")
	}
}

func TestLoginOnlyOnceThenError(t *testing.T) {
	m := newModel(State{Profile: "me"})
	m = update(m, checkMsg{err: assertErr("expired")})
	if m.result.Action != Login || !m.state.LoginTried {
		t.Fatalf("first failure must request login, got %+v", m.result)
	}

	m = newModel(m.state)
	m = update(m, checkMsg{err: assertErr("still expired")})
	if m.stage != stError {
		t.Fatalf("second failure must show the error, got stage %v", m.stage)
	}
}

func TestPresetInstanceConnects(t *testing.T) {
	m := newModel(State{Profile: "me"})
	m = update(m, checkMsg{})
	m = update(m, instancesMsg{list: []aws.Instance{{Name: "x-kali-attacker-ec2", ID: "i-1"}}})
	m = update(m, key("enter"))
	if m.result.Action != Connect || m.result.Remote != 6080 || m.result.Instance.ID != "i-1" {
		t.Fatalf("unexpected result %+v", m.result)
	}
}

func TestUnknownInstanceAsksPort(t *testing.T) {
	m := newModel(State{Profile: "me"})
	m = update(m, instancesMsg{list: []aws.Instance{{Name: "x-other-ec2", ID: "i-2"}}})
	m = update(m, key("enter"))
	if m.stage != stPort {
		t.Fatalf("want port stage, got %v", m.stage)
	}
	m = update(m, key("99999"))
	m = update(m, key("enter"))
	if m.result.Action == Connect {
		t.Fatal("out-of-range port must be rejected")
	}
	m.input = ""
	m = update(m, key("8080"))
	m = update(m, key("enter"))
	if m.result.Action != Connect || m.result.Remote != 8080 || m.result.Local != 8080 {
		t.Fatalf("unexpected result %+v", m.result)
	}
}

func TestEscReturnsToProfiles(t *testing.T) {
	m := newModel(State{Profile: "me"})
	m.profiles = []string{"me"}
	m = update(m, instancesMsg{list: nil})
	m = update(m, key("esc"))
	if m.stage != stProfile || m.state.Profile != "" {
		t.Fatalf("got stage %v profile %q", m.stage, m.state.Profile)
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }
