package dashboard

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func TestResizeFitsTerminal(t *testing.T) {
	for _, names := range [][]string{nil, {"grafana", "immich", "jellyfin", "dns", "proxy", "nas", "backup"}} {
		m := model{table: newTable(names)}
		// Cross each column breakpoint in both directions, including tiny windows.
		for _, width := range []int{120, 85, 84, 64, 63, 38, 28, 27, 10, 1, 120} {
			for _, height := range []int{24, 12, 11, 1} {
				t.Run(fmt.Sprintf("%d-services/%dx%d", len(names), width, height), func(t *testing.T) {
					updated, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
					m = updated.(model)
					view := m.View()
					if got := lipgloss.Width(view.Content); got > width {
						t.Errorf("render width = %d, terminal width = %d", got, width)
					}
					if got := lipgloss.Height(view.Content); got > height {
						t.Errorf("render height = %d, terminal height = %d", got, height)
					}
					if width >= 28 && height >= 12 && m.table.Height() < 1 {
						t.Error("no room for service rows")
					}
				})
			}
		}
	}
}
