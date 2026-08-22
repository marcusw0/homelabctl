package runbook

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/marcusw0/homelabctl/internal/markdown"
)

type runbookModel struct {
	path     string
	viewport viewport.Model
	style    string
	ready    bool
	err      error
}

type editorFinishedMsg struct{ err error }

type runbookReloadedMsg struct {
	content string
	err     error
}

func openEditor(path string) tea.Cmd {
	editor := cmp.Or(os.Getenv("EDITOR"), "vim")
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{err}
	})
}

func reloadRunbook(path string, style string, width int) tea.Cmd {
	return func() tea.Msg {
		source, err := os.ReadFile(path)
		if err != nil {
			return runbookReloadedMsg{err: err}
		}

		rendered, err := markdown.Render(
			source,
			width,
			style,
		)
		if err != nil {
			return runbookReloadedMsg{err: err}
		}

		return runbookReloadedMsg{content: string(rendered)}
	}
}

func newRunbookModel(path string, content string, style string) runbookModel {
	view := viewport.New()
	view.SetContent(content)
	view.SoftWrap = true

	return runbookModel{
		path:     path,
		viewport: view,
		style:    style,
	}
}

func (m runbookModel) Init() tea.Cmd {
	return nil
}

func (m runbookModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.Keystroke() {
		case "q", "esc":
			return m, tea.Quit
		case "e":
			return m, openEditor(m.path)
		}
	case editorFinishedMsg:
		if msg.err != nil {
			m.err = fmt.Errorf("edit runbook: %w", msg.err)
			return m, nil
		}

		width := max(20, m.viewport.Width())
		return m, reloadRunbook(
			m.path,
			m.style,
			width,
		)
	case runbookReloadedMsg:
		if msg.err != nil {
			m.err = fmt.Errorf("reload runbook: %w", msg.err)
			return m, nil
		}

		m.err = nil
		m.viewport.SetContent(msg.content)

	case tea.WindowSizeMsg:
		const footerHeight = 1

		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(max(1, msg.Height-footerHeight))
		m.ready = true
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

func (m runbookModel) View() tea.View {
	if !m.ready {
		view := tea.NewView("Loading runbook...")
		view.AltScreen = true
		return view
	}

	footer := fmt.Sprintf(
		"scroll: j/k, up/down edit: e quit: q, esc  %.0f%%",
		m.viewport.ScrollPercent()*100,
	)

	view := tea.NewView(
		m.viewport.View() + "\n" + footer,
	)

	view.AltScreen = true
	view.MouseMode = tea.MouseModeCellMotion
	view.WindowTitle = "homelabctl runbook"

	return view
}
