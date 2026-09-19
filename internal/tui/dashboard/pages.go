package dashboard

import (
	"fmt"
)

func (m *model) dashboardView() string {

	width := m.width - 4 // Two columns of margin on either side.
	count := fmt.Sprintf("%d services", len(m.table.Rows()))
	if len(m.table.Rows()) == 1 {
		count = "1 service"
	}
	header := headerLine(titleStyle.Render("HOMELABCTL"),
		mutedStyle.Render("LOCAL CONSOLE"), width) + "\n" +
		headerLine(mutedStyle.Render("Service overview"),
			mutedStyle.Render(count), width)

	footer := "↵ details · esc · q quit"
	if width >= 60 {
		footer = "↑↓ move · enter details · esc focus · q / ctrl+c quit"
	}
	if !m.table.Focused() {
		footer = "esc focus table · q quit"
	}
	content := header + "\n\n" + baseStyle.Render(m.table.View()) + "\n\n" + mutedStyle.Render(footer)
	view := screenStyle.Padding(1, 2).Width(m.width).Height(m.height).Render(content)

	return view
}

func (m *model) serviceView() string {
	width := m.width - 4
	header := titleStyle.Render("HOMELABCTL / " + m.selectedService)
	if host := m.checks[m.selectedService].FQDN; host != "" {
		header += "\n" + mutedStyle.Render(host)
	}

	content := header + "\n\n" + m.serviceDetails() + "\n\n" + mutedStyle.Render("esc back · q quit")
	return screenStyle.Padding(1, 2).Width(m.width).Height(m.height).
		Render(screenStyle.Width(width).Render(content))
}
