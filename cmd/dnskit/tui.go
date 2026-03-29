package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/serkanaltuntas/dnskit/pkg/dns"
)

type tabIndex int

const (
	recordsTab     tabIndex = 0
	propagationTab tabIndex = 1
	mapTab         tabIndex = 2
	tabCount                = 3
)

var (
	activeTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("6")).
			Bold(true).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("7")).
				Background(lipgloss.Color("8")).
				Padding(0, 2)

	tabSeparator = lipgloss.NewStyle().
			Faint(true).
			Padding(0, 1).
			Render("│")

	helpStyle = lipgloss.NewStyle().Faint(true)
)

type model struct {
	activeTab   tabIndex
	viewport    viewport.Model
	records     *dns.RecordSet
	propagation *dns.PropagationReport
	geoData     []dns.GeoLocation
	server      string
	ready       bool
	width       int
	height      int
}

func newModel(rs *dns.RecordSet, report *dns.PropagationReport, geoData []dns.GeoLocation, server string) model {
	return model{
		activeTab:   recordsTab,
		records:     rs,
		propagation: report,
		geoData:     geoData,
		server:      server,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab", "right", "l":
			m.activeTab = (m.activeTab + 1) % tabCount
			m.updateViewportContent()
			return m, nil
		case "shift+tab", "left", "h":
			m.activeTab = (m.activeTab + tabCount - 1) % tabCount
			m.updateViewportContent()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		vpHeight := m.height - 3

		if !m.ready {
			m.viewport = viewport.New(m.width, vpHeight)
			m.updateViewportContent()
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = vpHeight
		}

		return m, nil
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *model) updateViewportContent() {
	switch m.activeTab {
	case recordsTab:
		m.viewport.SetContent(renderRecordsTab(m.records, m.server))
	case propagationTab:
		if m.propagation != nil {
			m.viewport.SetContent(renderPropagationTab(m.propagation))
		} else {
			m.viewport.SetContent("\n  No propagation data available.\n")
		}
	case mapTab:
		m.viewport.SetContent(renderMapTab(m.geoData, m.records.Domain))
	}
	m.viewport.GotoTop()
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var b strings.Builder

	// Tab bar.
	tabs := []string{"DNS Records", "Propagation", "Map"}
	for i, t := range tabs {
		if i > 0 {
			b.WriteString(tabSeparator)
		}
		if tabIndex(i) == m.activeTab {
			b.WriteString(activeTabStyle.Render(t))
		} else {
			b.WriteString(inactiveTabStyle.Render(t))
		}
	}
	b.WriteString("\n")

	// Viewport content.
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Help line.
	b.WriteString(helpStyle.Render("  tab: switch tabs \u2022 \u2191/\u2193: scroll \u2022 q: quit"))

	return b.String()
}
