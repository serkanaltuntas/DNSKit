package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/serkanaltuntas/dnskit/pkg/dns"
)

var (
	cyanBold  = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	greenBold = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	yellowFg  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	dimStyle  = lipgloss.NewStyle().Faint(true)
	boldStyle = lipgloss.NewStyle().Bold(true)
	redBold   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
)

const viewWidth = 64

func sectionHeader(name string, count int) string {
	var b strings.Builder
	b.WriteString("\n")
	label := fmt.Sprintf("  %-6s", name)
	b.WriteString(greenBold.Render(label))
	remaining := viewWidth - 6
	if count > 0 {
		countStr := fmt.Sprintf("(%d) ", count)
		b.WriteString(dimStyle.Render(countStr))
		remaining -= len(countStr)
	}
	if remaining > 0 {
		b.WriteString(dimStyle.Render(strings.Repeat("\u2500", remaining)))
	}
	b.WriteString("\n\n")
	return b.String()
}

func tableHeader(widths []int, names []string) string {
	var b strings.Builder
	b.WriteString("    ")
	for i, name := range names {
		if i == len(names)-1 {
			b.WriteString(boldStyle.Render(name))
		} else {
			b.WriteString(boldStyle.Render(fmt.Sprintf("%-*s", widths[i], name)))
		}
	}
	b.WriteString("\n    ")
	total := 0
	for i := 0; i < len(names); i++ {
		if i == len(names)-1 {
			total += len(names[i])
		} else {
			total += widths[i]
		}
	}
	b.WriteString(dimStyle.Render(strings.Repeat("\u2500", total)))
	b.WriteString("\n")
	return b.String()
}

func tableDataRow(widths []int, vals []string) string {
	var b strings.Builder
	b.WriteString("    ")
	for i, v := range vals {
		if i == len(vals)-1 {
			b.WriteString(dimStyle.Render(v))
		} else {
			b.WriteString(fmt.Sprintf("%-*s", widths[i], v))
		}
	}
	b.WriteString("\n")
	return b.String()
}

func kvLine(key, value string) string {
	return yellowFg.Render(fmt.Sprintf("    %-16s", key)) + value + "\n"
}

func renderRecordsTab(rs *dns.RecordSet, server string) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  \u256d" + strings.Repeat("\u2500", viewWidth) + "\u256e") + "\n")
	b.WriteString(dimStyle.Render("  \u2502  ") + cyanBold.Render(fmt.Sprintf("%-*s", viewWidth-2, rs.Domain)) + dimStyle.Render("\u2502") + "\n")
	b.WriteString(dimStyle.Render("  \u2502  ") + dimStyle.Render(fmt.Sprintf("server: %-*s", viewWidth-10, server)) + dimStyle.Render("\u2502") + "\n")
	b.WriteString(dimStyle.Render("  \u2570" + strings.Repeat("\u2500", viewWidth) + "\u256f") + "\n")

	w2 := []int{40}
	w3 := []int{32, 12}

	if len(rs.A) > 0 {
		b.WriteString(sectionHeader("A", len(rs.A)))
		b.WriteString(tableHeader(w2, []string{"Address", "TTL"}))
		for _, r := range rs.A {
			b.WriteString(tableDataRow(w2, []string{r.Address, dns.FormatTTL(r.TTL)}))
		}
	}

	if len(rs.AAAA) > 0 {
		b.WriteString(sectionHeader("AAAA", len(rs.AAAA)))
		b.WriteString(tableHeader(w2, []string{"Address", "TTL"}))
		for _, r := range rs.AAAA {
			b.WriteString(tableDataRow(w2, []string{r.Address, dns.FormatTTL(r.TTL)}))
		}
	}

	if len(rs.CNAME) > 0 {
		b.WriteString(sectionHeader("CNAME", len(rs.CNAME)))
		b.WriteString(tableHeader(w2, []string{"Target", "TTL"}))
		for _, r := range rs.CNAME {
			b.WriteString(tableDataRow(w2, []string{r.Target, dns.FormatTTL(r.TTL)}))
		}
	}

	if len(rs.MX) > 0 {
		b.WriteString(sectionHeader("MX", len(rs.MX)))
		b.WriteString(tableHeader(w3, []string{"Host", "Priority", "TTL"}))
		for _, r := range rs.MX {
			b.WriteString(tableDataRow(w3, []string{r.Host, fmt.Sprintf("%d", r.Preference), dns.FormatTTL(r.TTL)}))
		}
	}

	if len(rs.NS) > 0 {
		b.WriteString(sectionHeader("NS", len(rs.NS)))
		b.WriteString(tableHeader(w2, []string{"Nameserver", "TTL"}))
		for _, r := range rs.NS {
			b.WriteString(tableDataRow(w2, []string{r.Nameserver, dns.FormatTTL(r.TTL)}))
		}
	}

	if len(rs.TXT) > 0 {
		b.WriteString(sectionHeader("TXT", len(rs.TXT)))
		for _, r := range rs.TXT {
			b.WriteString(fmt.Sprintf("    %q\n", r.Text))
			b.WriteString(dimStyle.Render(fmt.Sprintf("    TTL %s", dns.FormatTTL(r.TTL))) + "\n\n")
		}
	}

	if len(rs.SOA) > 0 {
		b.WriteString(sectionHeader("SOA", 0))
		for _, r := range rs.SOA {
			b.WriteString(kvLine("Primary NS", r.PrimaryNS))
			b.WriteString(kvLine("Mailbox", r.Mailbox))
			b.WriteString(kvLine("Serial", fmt.Sprintf("%d", r.Serial)))
			b.WriteString(kvLine("Refresh", dns.FormatTTL(r.Refresh)))
			b.WriteString(kvLine("Retry", dns.FormatTTL(r.Retry)))
			b.WriteString(kvLine("Expire", dns.FormatTTL(r.Expire)))
			b.WriteString(kvLine("Min TTL", dns.FormatTTL(r.MinTTL)))
			b.WriteString(kvLine("TTL", dns.FormatTTL(r.TTL)))
		}
	}

	if len(rs.SRV) > 0 {
		ws := []int{24, 8, 10, 10}
		b.WriteString(sectionHeader("SRV", len(rs.SRV)))
		b.WriteString(tableHeader(ws, []string{"Target", "Port", "Priority", "Weight", "TTL"}))
		for _, r := range rs.SRV {
			b.WriteString(tableDataRow(ws, []string{r.Target, fmt.Sprintf("%d", r.Port), fmt.Sprintf("%d", r.Priority), fmt.Sprintf("%d", r.Weight), dns.FormatTTL(r.TTL)}))
		}
	}

	if len(rs.CAA) > 0 {
		b.WriteString(sectionHeader("CAA", len(rs.CAA)))
		b.WriteString(tableHeader(w3, []string{"Tag", "Value", "TTL"}))
		for _, r := range rs.CAA {
			b.WriteString(tableDataRow(w3, []string{r.Tag, r.Value, dns.FormatTTL(r.TTL)}))
		}
	}

	if len(rs.PTR) > 0 {
		b.WriteString(sectionHeader("PTR", len(rs.PTR)))
		b.WriteString(tableHeader(w2, []string{"Host", "TTL"}))
		for _, r := range rs.PTR {
			b.WriteString(tableDataRow(w2, []string{r.Host, dns.FormatTTL(r.TTL)}))
		}
	}

	if len(rs.Errors) > 0 {
		b.WriteString("\n")
		b.WriteString(redBold.Render("  Warnings "))
		b.WriteString(dimStyle.Render(strings.Repeat("\u2500", viewWidth-11)))
		b.WriteString("\n\n")
		for typ, err := range rs.Errors {
			b.WriteString(yellowFg.Render(fmt.Sprintf("    %-10s", typ)))
			b.WriteString(dimStyle.Render(fmt.Sprintf("%v", err)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	return b.String()
}
