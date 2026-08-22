package runbook

import (
	"io"

	tea "charm.land/bubbletea/v2"
)

func Run(
	in io.Reader,
	out io.Writer,
	path string,
	content string,
	style string,
) error {
	model := newRunbookModel(
		path,
		content,
		style,
	)

	program := tea.NewProgram(
		model,
		tea.WithInput(in),
		tea.WithOutput(out),
	)

	_, err := program.Run()
	return err
}
