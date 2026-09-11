package dashboard

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/runner"
)

type model struct {
	table  table.Model
	width  int
	height int
	refreshing bool
	checks map[string][]check.Kind
	results <-chan runner.Result
	interval time.Duration
}

type refreshMsg struct {}

func refreshAfter(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return refreshMsg{}
	})
}

type serviceResultMsg struct{
	result runner.Result
}

type refreshFinishedMsg struct{}

func waitForResult(results <-chan runner.Result) tea.Cmd {
	return func() tea.Msg {
		result, ok := <-results
		if !ok {
			return refreshFinishedMsg{}
		}
		return serviceResultMsg{result: result}
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.width >= 28 && m.height >= 12 {
			m.resize()
		}
	case tea.KeyPressMsg:{}
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case refreshMsg:
		if m.refreshing {
			return m, nil
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		return m, waitForResult(m.results)
	case serviceResultMsg:
		return m, waitForResult(m.results)
	case refreshFinishedMsg:
		m.refreshing = false
		return m, refreshAfter(m.interval)
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	view := tea.NewView("")
	view.AltScreen = true
	view.BackgroundColor = phosphorBackground
	view.ForegroundColor = phosphorText
	if m.width == 0 || m.height == 0 {
		return view
	}
	if m.width < 28 || m.height < 12 {
		view.SetContent(screenStyle.MaxWidth(m.width).MaxHeight(m.height).
			Render("Resize to 28×12 or larger.\nq: quit"))
		return view
	}

	width := m.width - 4 // Two columns of margin on either side.
	count := fmt.Sprintf("%d services", len(m.table.Rows()))
	if len(m.table.Rows()) == 1 {
		count = "1 service"
	}
	header := headerLine(titleStyle.Render("HOMELABCTL"), mutedStyle.Render("LOCAL CONSOLE"), width) + "\n" +
		headerLine(mutedStyle.Render("Service overview"), mutedStyle.Render(count), width)
	footer := "↑/↓ move · esc focus · q quit"
	if width >= 60 {
		footer = "↑/↓ navigate   esc toggle focus   q / ctrl+c quit"
	} else if width < 32 {
		footer = "↑↓ move esc focus q quit"
	}
	if !m.table.Focused() {
		footer = "esc focus table · q quit"
	}
	content := header + "\n\n" + baseStyle.Render(m.table.View()) + "\n\n" + mutedStyle.Render(footer)
	view.SetContent(screenStyle.Padding(1, 2).Width(m.width).Height(m.height).Render(content))
	return view
}

func headerLine(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 2 {
		return left
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m *model) resize() {
	width := m.width - 4 - baseStyle.GetHorizontalFrameSize()
	columns := []table.Column{
		{Title: "SERVICE"},
		{Title: "HTTP", Width: 9},
		{Title: "DNS", Width: 9},
		{Title: "TLS", Width: 9},
		{Title: "TCP", Width: 9},
		{Title: "LAST CHECKED", Width: 19},
	}
	// Keep readable status columns; hide secondary columns on smaller screens.
	if width < 79 {
		columns[5].Width = 0
	}
	if width < 58 {
		for i := 1; i < len(columns); i++ {
			columns[i].Width = 0
		}
	}
	remaining := width - 2 // Service cell padding.
	for _, column := range columns[1:] {
		if column.Width > 0 {
			remaining -= column.Width + 2
		}
	}
	columns[0].Width = remaining
	m.table.SetColumns(columns)
	m.table.SetWidth(width)
	// Margins (2), header (2), gaps (2), footer (1), and table border (2).
	m.table.SetHeight(m.height - 9)
}
