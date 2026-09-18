package dashboard

import (
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"github.com/marcusw0/homelabctl/internal/check"
)

func (m *model) serviceDetails() string {
	result, ok := m.allResults[m.selectedService]
	if !ok {
		return "Waiting for first check…"
	}

	expiry := "—"
	if !result.Checks.TLS.Result.After.IsZero() {
		expiry = "Expires " + result.Checks.TLS.Result.After.Format("2006-01-02 15:04:05")
	}
	tcp := result.Checks.TCP.Result.Message
	if tcp == "" {
		tcp = result.Checks.TCP.Result.Target
	}

	checks := []struct {
		name   string
		status check.Status
		detail string
		err    error
	}{
		{"HTTP", result.Checks.HTTP.Status, result.Checks.HTTP.Result.Status, result.Checks.HTTP.Err},
		{"DNS", result.Checks.DNS.Status, strings.Join(result.Checks.DNS.Result.Response, ", "), result.Checks.DNS.Err},
		{"TLS", result.Checks.TLS.Status, expiry, result.Checks.TLS.Err},
		{"TCP", result.Checks.TCP.Status, tcp, result.Checks.TCP.Err},
	}
	sections := make([]string, 0, len(checks))
	for _, c := range checks {
		detail := c.detail
		if c.err != nil {
			detail = c.err.Error()
		}
		if detail == "" || c.status == check.StatusSkipped {
			detail = "—"
		}
		sections = append(sections, titleStyle.Render(c.name+"  "+strings.ToUpper(string(c.status)))+"\n"+detail)
	}
	return strings.Join(sections, "\n\n")
}

func newTable(names []string) table.Model {
	columns := []table.Column{
		{Title: "name", Width: 20},
		{Title: "HTTP", Width: 9},
		{Title: "DNS", Width: 9},
		{Title: "TLS", Width: 9},
		{Title: "TCP", Width: 9},
		{Title: "last checked", Width: 19},
	}

	rows := make([]table.Row, 0, len(names))
	for _, name := range names {
		rows = append(rows, table.Row{
			name, "—", "—", "—", "—", "never",
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		Foreground(phosphorMuted).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(phosphorBorder).
		BorderBackground(phosphorBackground).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(phosphorBright).
		Background(phosphorSelection).
		Bold(true)
	t.SetStyles(s)

	keyStyle := lipgloss.NewStyle().Foreground(phosphorText)
	mutedStyle := lipgloss.NewStyle().Foreground(phosphorMuted)
	t.Help.Styles.ShortKey = keyStyle
	t.Help.Styles.FullKey = keyStyle
	t.Help.Styles.ShortDesc = mutedStyle
	t.Help.Styles.FullDesc = mutedStyle
	t.Help.Styles.ShortSeparator = mutedStyle
	t.Help.Styles.FullSeparator = mutedStyle
	t.Help.Styles.Ellipsis = mutedStyle

	// Include cell/header padding so the viewport fits every column.
	columnFrame := max(s.Header.GetHorizontalFrameSize(), s.Cell.GetHorizontalFrameSize())
	width := 0
	for _, column := range columns {
		width += column.Width + columnFrame
	}
	t.SetWidth(width)

	return t
}
