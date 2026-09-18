package dashboard

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/marcusw0/homelabctl/internal/check"
	"github.com/marcusw0/homelabctl/internal/runner"
)

func Run(ctx context.Context, names []string, checks map[string]check.Service) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	t := newTable(names)

	m := model{
		page:       "dashboard",
		table:      t,
		checks:     checks,
		ctx:        ctx,
		interval:   20 * time.Second,
		refreshing: false,
		allResults: make(map[string]runner.Result),
	}
	if _, err := tea.NewProgram(m, tea.WithContext(ctx)).Run(); err != nil {
		return err
	}
	return nil
}
