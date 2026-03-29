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
)

var (
	activeTabStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true).BorderBottom(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("6")).Padding(0, 2)
	inactiveTabStyle = lipgloss.NewStyle().Faint(true).Padding(0, 2)
	helpStyle        = lipgloss.NewStyle().Faint(true)
)

type model struct {
	activeTab   tabIndex
	viewport    viewport.Model
	records     *dns.RecordSet
	propagation *dns.PropagationReport
	server      string
	ready       bool
	width       int
	height      int
}

func newModel(rs *dns.RecordSet, report *dns.PropagationReport, server string) model {
	return model{
		activeTab:   recordsTab,
		records:     rs,
		propagation: report,
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
			m.activeTab = (m.activeTab + 1) % 2
			m.updateViewportContent()
			return m, nil
		case "shift+tab", "left", "h":
			m.activeTab = (m.activeTab + 2 - 1) % 2
			m.updateViewportContent()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		vpHeight := m.height - 3 // tab bar + separator + help line

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
	}
	m.viewport.GotoTop()
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var b strings.Builder

	// Tab bar
	tabs := []string{"DNS Records", "Propagation"}
	var renderedTabs []string
	for i, t := range tabs {
		if tabIndex(i) == m.activeTab {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(t))
		} else {
			renderedTabs = append(renderedTabs, inactiveTabStyle.Render(t))
		}
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Bottom, renderedTabs...))
	b.WriteString("\n")

	// Viewport content
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Help line
	b.WriteString(helpStyle.Render("  tab: switch tabs \u2022 \u2191/\u2193: scroll \u2022 q: quit"))

	return b.String()
}
