// Package ui is the Bubble Tea front end. It never runs the interactive aws commands itself;
// it returns a Result so the caller can run them with the terminal and signals to itself.
// Non-interactive calls (listing, power on/off) run as commands inside the TUI.
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kintoun-secops/kintoun-ssm-tunnel/internal/aws"
)

type Action int

const (
	Quit Action = iota
	Login
	Connect
)

type Result struct {
	Action   Action
	Profile  string
	Instance aws.Instance
	Remote   int
	Scheme   string // non-empty builds a scheme://host:port/path URL; empty prints host:port
	Path     string
}

// State carries what the caller keeps between TUI runs.
type State struct {
	Profile    string
	LoginTried bool
}

type stage int

const (
	stProfile stage = iota
	stProfileInput
	stChecking
	stInstances
	stAction
	stPort
	stPowering
	stError
)

type actionKind int

const (
	actTunnel actionKind = iota
	actStart
	actStop
)

type actionItem struct {
	label string
	kind  actionKind
}

// actionsFor lists what can be done to an instance in the given power state. While an instance
// is transitioning there is nothing to do but wait, so the list is empty.
func actionsFor(state string) []actionItem {
	switch state {
	case "running":
		return []actionItem{{"터널 열기", actTunnel}, {"중지", actStop}}
	case "stopped":
		return []actionItem{{"시작", actStart}}
	default:
		return nil
	}
}

type checkMsg struct{ err error }
type instancesMsg struct {
	list []aws.Instance
	err  error
}
type powerMsg struct{ err error }

var (
	titleStyle  = lipgloss.NewStyle().Bold(true)
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

type model struct {
	state     State
	stage     stage
	profiles  []string
	instances []aws.Instance
	actions   []actionItem
	cursor    int
	input     string
	picked    aws.Instance
	err       error
	result    Result
}

func newModel(state State) model {
	m := model{state: state, profiles: aws.Profiles()}
	switch {
	case state.Profile != "":
		m.stage = stChecking
	case len(m.profiles) == 0:
		m.stage = stProfileInput
	default:
		m.stage = stProfile
	}
	return m
}

// Run shows the TUI and returns what the user chose.
func Run(state State) (Result, State, error) {
	final, err := tea.NewProgram(newModel(state)).Run()
	if err != nil {
		return Result{}, state, err
	}
	m := final.(model)
	return m.result, m.state, nil
}

func (m model) Init() tea.Cmd {
	if m.stage == stChecking {
		return checkCmd(m.state.Profile)
	}
	return nil
}

func checkCmd(profile string) tea.Cmd {
	return func() tea.Msg { return checkMsg{aws.CheckCredentials(profile)} }
}

func listCmd(profile string) tea.Cmd {
	return func() tea.Msg {
		list, err := aws.Instances(profile)
		return instancesMsg{list, err}
	}
}

func startCmd(profile, id string) tea.Cmd {
	return func() tea.Msg { return powerMsg{aws.StartInstance(profile, id)} }
}

func stopCmd(profile, id string) tea.Cmd {
	return func() tea.Msg { return powerMsg{aws.StopInstance(profile, id)} }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case checkMsg:
		if msg.err == nil {
			m.state.LoginTried = false
			return m, listCmd(m.state.Profile)
		}
		if !m.state.LoginTried {
			m.state.LoginTried = true
			m.result = Result{Action: Login, Profile: m.state.Profile}
			return m, tea.Quit
		}
		m.err, m.stage = msg.err, stError
		return m, nil

	case instancesMsg:
		if msg.err != nil {
			m.err, m.stage = msg.err, stError
			return m, nil
		}
		m.instances, m.cursor, m.stage = msg.list, 0, stInstances
		return m, nil

	case powerMsg:
		if msg.err != nil {
			m.err, m.stage = msg.err, stError
			return m, nil
		}
		// 전원 상태가 바뀌었으니 목록을 다시 불러 반영한다.
		m.stage = stChecking
		return m, listCmd(m.state.Profile)

	case tea.KeyMsg:
		return m.onKey(msg)
	}
	return m, nil
}

func (m model) onKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "ctrl+c" {
		return m, tea.Quit
	}
	typing := m.stage == stProfileInput || m.stage == stPort
	if !typing && key == "q" {
		return m, tea.Quit
	}

	switch key {
	case "up", "k":
		if !typing && m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if !typing && m.cursor < m.listLen()-1 {
			m.cursor++
		}
	case "esc":
		return m.back()
	case "enter":
		return m.enter()
	case "backspace":
		if typing && m.input != "" {
			r := []rune(m.input)
			m.input = string(r[:len(r)-1])
		}
	default:
		if typing && msg.Type == tea.KeyRunes {
			m.input += string(msg.Runes)
		}
	}
	return m, nil
}

func (m model) listLen() int {
	switch m.stage {
	case stProfile:
		return len(m.profiles) + 1
	case stInstances:
		return len(m.instances)
	case stAction:
		return len(m.actions)
	}
	return 0
}

func (m model) back() (tea.Model, tea.Cmd) {
	m.err, m.input, m.cursor = nil, "", 0
	switch m.stage {
	case stPort:
		m.stage = stAction
	case stAction:
		m.stage = stInstances
	case stInstances, stError:
		m.state.Profile, m.state.LoginTried = "", false
		if len(m.profiles) == 0 {
			m.stage = stProfileInput
		} else {
			m.stage = stProfile
		}
	case stProfileInput:
		if len(m.profiles) > 0 {
			m.stage = stProfile
		}
	}
	return m, nil
}

func (m model) enter() (tea.Model, tea.Cmd) {
	switch m.stage {
	case stProfile:
		if m.cursor == len(m.profiles) {
			m.stage, m.input = stProfileInput, ""
			return m, nil
		}
		return m.selectProfile(m.profiles[m.cursor])

	case stProfileInput:
		if name := strings.TrimSpace(m.input); name != "" {
			return m.selectProfile(name)
		}

	case stInstances:
		if len(m.instances) == 0 {
			return m, nil
		}
		m.picked = m.instances[m.cursor]
		m.actions, m.cursor, m.stage = actionsFor(m.picked.State), 0, stAction
		return m, nil

	case stAction:
		if len(m.actions) == 0 {
			return m, nil
		}
		switch m.actions[m.cursor].kind {
		case actTunnel:
			return m.openTunnel()
		case actStart:
			m.stage = stPowering
			return m, startCmd(m.state.Profile, m.picked.ID)
		case actStop:
			m.stage = stPowering
			return m, stopCmd(m.state.Profile, m.picked.ID)
		}

	case stPort:
		port, ok := aws.ParsePort(strings.TrimSpace(m.input))
		if !ok {
			return m, nil
		}
		m.result = Result{Action: Connect, Profile: m.state.Profile, Instance: m.picked, Remote: port}
		return m, tea.Quit

	case stError:
		return m.back()
	}
	return m, nil
}

// openTunnel finishes the connect flow for the picked instance. A preset fixes the remote
// port and URL shape; otherwise the user is asked for the remote port.
func (m model) openTunnel() (tea.Model, tea.Cmd) {
	if p, ok := findPreset(m.picked.Name); ok {
		m.result = Result{Action: Connect, Profile: m.state.Profile, Instance: m.picked, Remote: p.Remote, Scheme: p.Scheme, Path: p.Path}
		return m, tea.Quit
	}
	m.stage, m.input = stPort, ""
	return m, nil
}

func (m model) selectProfile(name string) (tea.Model, tea.Cmd) {
	m.state.Profile, m.state.LoginTried = name, false
	m.stage, m.input = stChecking, ""
	return m, checkCmd(name)
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("kintoun SSM 터널") + "\n\n")

	switch m.stage {
	case stProfile:
		b.WriteString("AWS 프로필을 선택하세요.\n\n")
		for i, p := range m.profiles {
			b.WriteString(m.row(i, p, ""))
		}
		b.WriteString(m.row(len(m.profiles), "직접 입력...", ""))
		b.WriteString(help("↑↓ 이동  Enter 선택  q 종료"))

	case stProfileInput:
		b.WriteString("프로필 이름을 입력하세요. 없는 프로필이면 aws login 으로 새로 만듭니다.\n\n")
		b.WriteString("> " + m.input + "█\n")
		b.WriteString(help("Enter 확인  Esc 뒤로  Ctrl+C 종료"))

	case stChecking:
		b.WriteString(fmt.Sprintf("%s · 인스턴스를 조회하는 중...\n", m.state.Profile))

	case stInstances:
		b.WriteString(fmt.Sprintf("프로필 %s · 인스턴스를 선택하세요.\n\n", m.state.Profile))
		if len(m.instances) == 0 {
			b.WriteString(dimStyle.Render("SSMPortForward=true 태그가 붙은 인스턴스가 없습니다.") + "\n")
		}
		for i, in := range m.instances {
			b.WriteString(m.row(i, in.Name, in.ID+"  ·  "+stateLabel(in.State)))
		}
		b.WriteString(help("↑↓ 이동  Enter 선택  Esc 프로필 변경  q 종료"))

	case stAction:
		b.WriteString(fmt.Sprintf("%s  (%s)\n\n", m.picked.Name, stateLabel(m.picked.State)))
		if len(m.actions) == 0 {
			b.WriteString(dimStyle.Render("전환 중이라 지금은 조작할 수 없습니다.") + "\n")
		}
		for i, a := range m.actions {
			b.WriteString(m.row(i, a.label, ""))
		}
		b.WriteString(help("↑↓ 이동  Enter 실행  Esc 뒤로  q 종료"))

	case stPort:
		b.WriteString(fmt.Sprintf("%s 의 원격 포트를 입력하세요. 로컬 포트는 빈 포트를 자동으로 잡습니다.\n\n", m.picked.Name))
		b.WriteString("> " + m.input + "█\n")
		b.WriteString(help("Enter 확인  Esc 뒤로  Ctrl+C 종료"))

	case stPowering:
		b.WriteString(fmt.Sprintf("%s 의 전원 상태를 바꾸는 중...\n", m.picked.Name))

	case stError:
		b.WriteString(errStyle.Render(m.err.Error()) + "\n")
		b.WriteString(help("Enter 또는 Esc 로 프로필 선택으로 돌아가기  q 종료"))
	}
	return b.String()
}

func stateLabel(s string) string {
	switch s {
	case "running":
		return "실행 중"
	case "stopped":
		return "중지됨"
	case "stopping":
		return "중지하는 중"
	case "pending":
		return "시작하는 중"
	default:
		return s
	}
}

func (m model) row(i int, label, note string) string {
	prefix, text := "  ", label
	if i == m.cursor {
		prefix, text = cursorStyle.Render("> "), cursorStyle.Render(label)
	}
	if note != "" {
		text += "  " + dimStyle.Render(note)
	}
	return prefix + text + "\n"
}

func help(s string) string { return "\n" + dimStyle.Render(s) + "\n" }
