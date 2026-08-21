package markdown

import (
	"fmt"

	"charm.land/glamour/v2"
)

func Render(source []byte, width int, style string) ([]byte, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, fmt.Errorf("create Markdown renderer: %w", err)
	}

	rendered, err := renderer.RenderBytes(source)
	if err != nil {
		return nil, fmt.Errorf("render Markdown: %w", err)
	}

	return rendered, nil
}
