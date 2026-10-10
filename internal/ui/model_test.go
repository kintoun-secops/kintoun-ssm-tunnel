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
	cases := map[string]Preset{
		"kintoun-secops-infra-kali-attacker-ec2": {Remote: 6080, Scheme: "http", Path: "/vnc.html"},
		"kintoun-secops-infra-wazuh-ec2":         {Remote: 443, Scheme: "https"},
		"kintoun-secops-infra-velociraptor-ec2":  {Remote: 8889, Scheme: "https"},
	}
	for name, want := range cases {
		p, ok := findPreset(name)
		if !ok || p.Remote != want.Remote || p.Scheme != want.Scheme || p.Path != want.Path {
			t.Errorf("%s: got ok=%v %+v, want %+v", name, ok, p, want)
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
	m = update(m, instancesMsg{list: []aws.Instance{{Name: "x-kali-attacker-ec2", ID: "i-1", State: "running"}}})
	m = update(m, key("enter")) // instance -> action menu
	if m.stage != stAction {
		t.Fatalf("want action menu, got stage %v", m.stage)
	}
	m = update(m, key("enter")) // 터널 열기 (cursor 0)
	if m.result.Action != Connect || m.result.Remote != 6080 || m.result.Instance.ID != "i-1" {
		t.Fatalf("unexpected result %+v", m.result)
	}
}

func TestUnknownInstanceAsksPort(t *testing.T) {
	m := newModel(State{Profile: "me"})
	m = update(m, instancesMsg{list: []aws.Instance{{Name: "x-other-ec2", ID: "i-2", State: "running"}}})
	m = update(m, key("enter")) // instance -> action menu
	m = update(m, key("enter")) // 터널 열기 -> port
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
	if m.result.Action != Connect || m.result.Remote != 8080 {
		t.Fatalf("unexpected result %+v", m.result)
	}
}

func TestStoppedInstanceStarts(t *testing.T) {
	m := newModel(State{Profile: "me"})
	m = update(m, instancesMsg{list: []aws.Instance{{Name: "x-php-ec2", ID: "i-3", State: "stopped"}}})
	m = update(m, key("enter")) // instance -> action menu
	if m.stage != stAction || len(m.actions) != 1 || m.actions[0].kind != actStart {
		t.Fatalf("stopped instance should offer only start, got stage %v actions %+v", m.stage, m.actions)
	}
	m = update(m, key("enter")) // 시작
	if m.stage != stPowering {
		t.Fatalf("want powering stage, got %v", m.stage)
	}
	m = update(m, powerMsg{}) // 성공하면 목록을 다시 불러온다
	if m.stage != stChecking {
		t.Fatalf("after power change, want re-list, got %v", m.stage)
	}
}

func TestRunningInstanceStops(t *testing.T) {
	m := newModel(State{Profile: "me"})
	m = update(m, instancesMsg{list: []aws.Instance{{Name: "x-php-ec2", ID: "i-4", State: "running"}}})
	m = update(m, key("enter")) // instance -> action menu
	if len(m.actions) != 2 {
		t.Fatalf("running instance should offer tunnel and stop, got %+v", m.actions)
	}
	m = update(m, key("j"))     // move to 중지
	m = update(m, key("enter")) // 중지
	if m.stage != stPowering {
		t.Fatalf("want powering stage, got %v", m.stage)
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
