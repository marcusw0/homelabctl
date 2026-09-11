package dashboard

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/marcusw0/homelabctl/internal/check"
)

func Run(ctx context.Context, names []string, checks map[string][]check.Kind) error {

	t := newTable(names)

	m := model{table: t, checks: checks}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		return err
	}
	return nil
}
