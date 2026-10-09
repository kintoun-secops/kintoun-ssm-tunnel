// Package ui is the Bubble Tea front end. It never runs the interactive aws commands itself;
// it returns a Result so the caller can run them with the terminal and signals to itself.
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
	Local    int
	URL      string
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
	stPort
	stError
)

type checkMsg struct{ err error }
type instancesMsg struct {
	list []aws.Instance
	err  error
}

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
		list, err := aws.PortForwardInstances(profile)
		return instancesMsg{list, err}
	}
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
	}
	return 0
}

func (m model) back() (tea.Model, tea.Cmd) {
	m.err, m.input, m.cursor = nil, "", 0
	switch m.stage {
	case stInstances, stPort, stError:
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
		inst := m.instances[m.cursor]
		if p, ok := findPreset(inst.Name); ok {
			m.result = Result{Action: Connect, Profile: m.state.Profile, Instance: inst, Remote: p.Remote, Local: p.Local, URL: p.URL()}
			return m, tea.Quit
		}
		m.picked, m.stage, m.input = inst, stPort, ""

	case stPort:
		port, ok := aws.ParsePort(strings.TrimSpace(m.input))
		if !ok {
			return m, nil
		}
		m.result = Result{Action: Connect, Profile: m.state.Profile, Instance: m.picked, Remote: port, Local: port, URL: fmt.Sprintf("localhost:%d", port)}
		return m, tea.Quit

	case stError:
		return m.back()
	}
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
		b.WriteString(fmt.Sprintf("%s 자격증명을 확인하고 인스턴스를 조회하는 중...\n", m.state.Profile))

	case stInstances:
		b.WriteString(fmt.Sprintf("프로필 %s · 접속할 인스턴스를 선택하세요.\n\n", m.state.Profile))
		if len(m.instances) == 0 {
			b.WriteString(dimStyle.Render("SSMPortForward=true 태그가 붙은 실행 중 인스턴스가 없습니다.") + "\n")
		}
		for i, in := range m.instances {
			note := in.ID
			if p, ok := findPreset(in.Name); ok {
				note += "  →  " + p.URL()
			}
			b.WriteString(m.row(i, in.Name, note))
		}
		b.WriteString(help("↑↓ 이동  Enter 터널 열기  Esc 프로필 변경  q 종료"))

	case stPort:
		b.WriteString(fmt.Sprintf("%s 는 기본 포트가 없습니다. 포트를 입력하세요.\n\n", m.picked.Name))
		b.WriteString("> " + m.input + "█\n")
		b.WriteString(help("Enter 확인  Esc 뒤로  Ctrl+C 종료"))

	case stError:
		b.WriteString(errStyle.Render(m.err.Error()) + "\n")
		b.WriteString(help("Enter 또는 Esc 로 프로필 선택으로 돌아가기  q 종료"))
	}
	return b.String()
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
