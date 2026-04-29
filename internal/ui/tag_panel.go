package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/rivo/tview"
)

// TagPanel renders the user-defined tags associated with a probe target.
type TagPanel struct {
	view *tview.TextView
}

// NewTagPanel creates a TagPanel ready to embed in a layout.
func NewTagPanel() *TagPanel {
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetBorder(true)
	tv.SetTitle(" Tags ")
	return &TagPanel{view: tv}
}

// Update redraws the panel with the provided tag map.
// Passing nil or an empty map renders a placeholder message.
func (p *TagPanel) Update(tags map[string]string) {
	p.view.Clear()

	if len(tags) == 0 {
		fmt.Fprintf(p.view, "[grey]no tags defined[-]")
		return
	}

	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&sb, "[yellow]%-14s[-] %s\n", k, tags[k])
	}
	fmt.Fprint(p.view, sb.String())
}

// View returns the underlying tview primitive for layout embedding.
func (p *TagPanel) View() *tview.TextView {
	return p.view
}
