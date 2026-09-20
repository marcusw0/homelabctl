package serviceform

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/config"
)

func Run(
	ctx context.Context,
	in io.Reader,
	out io.Writer,
	name string,
) (config.Service, error) {
	if err := ctx.Err(); err != nil {
		return config.Service{}, err
	}

	final, err := tea.NewProgram(
		newModel(name),
		tea.WithContext(ctx),
		tea.WithInput(in),
		tea.WithOutput(out)).Run()
	if err != nil {
		if ctx.Err() != nil {
			return config.Service{}, ctx.Err()
		}
		return config.Service{}, fmt.Errorf("service form: %w", err)
	}

	m, ok := final.(model)
	if !ok {
		return config.Service{}, fmt.Errorf("unexpected form model %T", final)
	}
	if !m.submitted {
		return config.Service{}, errors.New("service creation canceled")
	}
	return m.service, nil
}

type fieldID int

const (
	fqdnField fieldID = iota
	ipField
	portField
	checksField
	advancedField
	runbookField
	expectedField
	timeoutField
	saveField
)

type field struct {
	id       fieldID
	label    string
	advanced bool
	input    textinput.Model
}

type checkOption struct {
	kind     check.Kind
	selected bool
}

type fieldError struct {
	field fieldID
	err   error
}

type model struct {
	name         string
	fields       []field
	focus        fieldID
	checks       []checkOption
	checkCursor  int
	advancedOpen bool
	submitted    bool
	service      config.Service
	failure      *fieldError
}

func newModel(name string) model {
	m := model{name: name}
	defaults := config.Service{}
	for _, spec := range []struct {
		id                 fieldID
		label, placeholder string
		advanced           bool
	}{
		{fqdnField, "FQDN", "example.com", false},
		{ipField, "IP", "127.0.0.1", false},
		{portField, "Port", "443", false},
		{runbookField, "Runbook", "runbooks/myservice.md", true},
		{expectedField, "Expected status", strconv.Itoa(defaults.EffectiveStatusCode()) + " (default)", true},
		{timeoutField, "Timeout", defaults.EffectiveTimeout().String() + " (default)", true},
	} {
		input := textinput.New()
		input.Placeholder = spec.placeholder
		input.SetWidth(32)
		input.SetVirtualCursor(true)
		m.fields = append(m.fields, field{spec.id, spec.label, spec.advanced, input})
	}

	for _, kind := range defaults.EffectiveChecks() {
		m.checks = append(m.checks, checkOption{kind, true})
	}
	m.syncFocus()
	return m
}

func (m model) visibleOrder() []fieldID {
	var order []fieldID
	for _, f := range m.fields {
		if !f.advanced {
			order = append(order, f.id)
		}
	}
	order = append(order, checksField, advancedField)
	if m.advancedOpen {
		for _, f := range m.fields {
			if f.advanced {
				order = append(order, f.id)
			}
		}
	}
	return append(order, saveField)
}

func (m *model) syncFocus() tea.Cmd {
	var cmds []tea.Cmd
	for i := range m.fields {
		f := &m.fields[i]
		if f.id == m.focus && (!f.advanced || m.advancedOpen) {
			cmds = append(cmds, f.input.Focus())
		} else {
			f.input.Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "esc", "ctrl+c":
			m.submitted = false
			return m, tea.Quit

		case "tab", "shift+tab":
			order := m.visibleOrder()
			position := 0
			for i, id := range order {
				if id == m.focus {
					position = i
					break
				}
			}
			step := 1
			if key.String() == "shift+tab" {
				step = -1
			}
			m.focus = order[(position+step+len(order))%len(order)]
			return m, m.syncFocus()

		case "enter", "space":
			if m.focus == advancedField {
				m.advancedOpen = !m.advancedOpen
				return m, m.syncFocus()
			}
			if m.focus == checksField && key.String() == "space" {
				m.checks[m.checkCursor].selected = !m.checks[m.checkCursor].selected
				return m, nil
			}
			if key.String() == "enter" {
				if m.focus == saveField {
					service, failure := m.submit()
					m.failure = failure
					if failure != nil {
						m.focus = failure.field
						for _, f := range m.fields {
							if f.id == failure.field && f.advanced {
								m.advancedOpen = true
							}
						}
						return m, m.syncFocus()
					}
					m.service, m.submitted = service, true
					return m, tea.Quit
				}
				return m, nil
			}

		case "up", "down":
			if m.focus == checksField {
				step := 1
				if key.String() == "up" {
					step = -1
				}
				m.checkCursor = (m.checkCursor + step + len(m.checks)) % len(m.checks)
				return m, nil
			}
		}
	}
	var cmds []tea.Cmd
	for i := range m.fields {
		var cmd tea.Cmd
		m.fields[i].input, cmd = m.fields[i].input.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m model) value(id fieldID) string {
	for _, f := range m.fields {
		if f.id == id {
			return strings.TrimSpace(f.input.Value())
		}
	}
	return ""
}

func (m model) submit() (config.Service, *fieldError) {
	fail := func(id fieldID, err error) (config.Service, *fieldError) {
		return config.Service{}, &fieldError{id, err}
	}
	s := config.Service{Enabled: true, FQDN: m.value(fqdnField), IP: m.value(ipField), Runbook: m.value(runbookField)}
	if err := check.ValidateHostname(s.FQDN); err != nil {
		return fail(fqdnField, err)
	}
	if err := check.ValidateIP(s.IP); err != nil {
		return fail(ipField, err)
	}
	port, err := strconv.Atoi(m.value(portField))
	if err != nil {
		return fail(portField, errors.New("port must be a number"))
	}
	if err := check.ValidatePort(port); err != nil {
		return fail(portField, err)
	}

	s.Port = port
	for _, option := range m.checks {
		if option.selected {
			s.Checks = append(s.Checks, option.kind)
		}
	}

	if len(s.Checks) == 0 {
		return fail(checksField, errors.New("select at least one check"))
	}
	if value := m.value(expectedField); value != "" {
		s.ExpectedStatus, err = strconv.Atoi(value)
		if err != nil {
			return fail(expectedField, errors.New("expected status must be a number"))
		}
	}
	if err := config.ValidateExpectedStatus(s.ExpectedStatus); err != nil {
		return fail(expectedField, err)
	}
	if value := m.value(timeoutField); value != "" {
		s.Timeout, err = time.ParseDuration(value)
		if err != nil {
			return fail(timeoutField, errors.New("timeout must be a duration, such as 5s"))
		}
	}
	if err := config.ValidateTimeout(s.Timeout); err != nil {
		return fail(timeoutField, err)
	}

	return s, nil
}

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D0F0C0"))
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#77967D"))
)

func (m model) View() tea.View {
	var b strings.Builder

	fmt.Fprintf(&b, "Add service: %s\n\n", m.name)
	for _, id := range m.visibleOrder() {
		switch id {
		case checksField:
			b.WriteString("\nChecks\n")
			for i, option := range m.checks {
				pointer, mark := " ", " "
				if m.focus == checksField && i == m.checkCursor {
					pointer = ">"
				}
				if option.selected {
					mark = "X"
				}
				fmt.Fprintf(&b, "%s [%s] %s\n", pointer, mark, option.kind)
			}
		case advancedField:
			heading := "▸ Advanced options"
			if m.advancedOpen {
				heading = "▾ Advanced options"
			}
			if m.focus == id {
				heading = focusedStyle.Render("> " + heading)
			} else {
				heading = mutedStyle.Render(heading)
			}
			fmt.Fprintf(&b, "\n%s\n", heading)
		case saveField:
			button := "[ Save ]"
			if m.focus == id {
				button = focusedStyle.Render("> " + button)
			} else {
				button = mutedStyle.Render(button)
			}
			fmt.Fprintf(&b, "\n%s\n", button)
		default:
			for _, f := range m.fields {
				if f.id == id {
					label := f.label
					if m.focus == id {
						label = focusedStyle.Render(label)
					} else {
						label = mutedStyle.Render(label)
					}
					fmt.Fprintf(&b, "%s: %s\n", label, f.input.View())
					break
				}
			}
		}
		if m.failure != nil && m.failure.field == id {
			fmt.Fprintf(&b, "  Error: %s\n", m.failure.err)
		}
	}

	b.WriteString("\nTab/Shift+Tab: move · Esc/Ctrl+C: cancel")
	switch m.focus {
	case checksField:
		b.WriteString(" · ↑/↓: choose check · Space: toggle")
	case advancedField:
		b.WriteString(" · Enter/Space: expand or collapse")
	case saveField:
		b.WriteString(" · Enter: save")
	}

	return tea.NewView(b.String())
}
