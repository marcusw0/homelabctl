package dashboard

import (
	"context"
	"strings"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/runner"
)

type model struct {
	ctx                   context.Context
	table                 table.Model
	width                 int
	height                int
	refreshing            bool
	checks                map[string]check.Service
	results               <-chan runner.Result
	interval              time.Duration
	page, selectedService string
	allResults            map[string]runner.Result
}

type refreshMsg struct{}

func refreshAfter(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return refreshMsg{}
	})
}

type serviceResultMsg struct {
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

func (m model) Init() tea.Cmd {
	return func() tea.Msg { return refreshMsg{} }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.width >= 28 && m.height >= 12 {
			m.resize()
		}
	case tea.KeyPressMsg:
		{
		}
		switch msg.String() {
		case "esc":
			if m.page == "dashboard" {
				if m.table.Focused() {
					m.table.Blur()
				} else {
					m.table.Focus()
				}
			} else {
				m.page = "dashboard"
				return m, nil
			}
		case "enter":
			if m.page != "dashboard" {
				return m, nil
			}
			row := m.table.SelectedRow()
			if len(row) > 0 {
				m.selectedService = row[0]
				m.page = "service"
			}
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case refreshMsg:
		if m.refreshing {
			return m, nil
		}
		jobs := make([]runner.Job, 0, len(m.checks))

		for name, service := range m.checks {
			jobs = append(jobs, runner.Job{
				Name:    name,
				Checker: &service,
			})
		}
		serviceRunner := runner.Runner{
			MaxConcurrent: 4,
		}

		m.refreshing = true
		m.results = serviceRunner.Run(m.ctx, jobs)
		return m, waitForResult(m.results)

	case serviceResultMsg:
		result := msg.result
		m.allResults[result.Name] = result

		rows := m.table.Rows()

		for i, row := range rows {
			if row[0] != result.Name {
				continue
			}
			rows[i] = table.Row{
				result.Name,
				string(result.Checks.HTTP.Status),
				string(result.Checks.DNS.Status),
				string(result.Checks.TLS.Status),
				string(result.Checks.TCP.Status),
				time.Now().Format("2006-01-02 15:04:05"),
			}
			break
		}
		m.table.SetRows(rows)
		return m, waitForResult(m.results)

	case refreshFinishedMsg:
		m.refreshing = false
		return m, refreshAfter(m.interval)
	}
	if m.page == "dashboard" {
		m.table, cmd = m.table.Update(msg)
	}
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

	if m.page == "service" {
		view.SetContent(m.serviceView())
	} else {
		view.SetContent(m.dashboardView())
	}

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
