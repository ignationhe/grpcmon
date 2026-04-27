package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"

	"github.com/yourorg/grpcmon/internal/probe"
)

// MetadataPanel renders a table of discovered service metadata.
type MetadataPanel struct {
	view *tview.TextView
}

// NewMetadataPanel creates a MetadataPanel ready to embed in a layout.
func NewMetadataPanel() *MetadataPanel {
	v := tview.NewTextView()
	v.SetDynamicColors(true)
	v.SetBorder(true)
	v.SetTitle(" Service Metadata ")
	return &MetadataPanel{view: v}
}

// Update refreshes the panel content from the provided metadata slice.
func (p *MetadataPanel) Update(metas []*probe.ServiceMeta) {
	if len(metas) == 0 {
		p.view.SetText("[grey]No metadata collected yet.[-]")
		return
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "[yellow]%-30s %-20s %-12s %-20s[-]\n",
		"Address", "Server", "gRPC Ver", "Discovered")
	fmt.Fprintf(&sb, "%s\n", strings.Repeat("─", 86))

	for _, m := range metas {
		serverName := m.ServerName
		if serverName == "" {
			serverName = "[grey]unknown[-]"
		}
		grpcVer := m.GRPCVersion
		if grpcVer == "" {
			grpcVer = "[grey]-[-]"
		}
		discovered := m.DiscoveredAt.Format(time.RFC3339)
		fmt.Fprintf(&sb, "%-30s %-20s %-12s %-20s\n",
			m.Address, serverName, grpcVer, discovered)
	}

	p.view.SetText(tview.TranslateANSI(sb.String()))
}

// Primitive returns the underlying tview primitive for layout embedding.
func (p *MetadataPanel) Primitive() tview.Primitive {
	return p.view
}
