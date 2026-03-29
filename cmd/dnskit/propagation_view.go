package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/serkanaltuntas/dnskit/pkg/dns"
)

var (
	redFg     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	greenFg   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	warningFg = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

func hasRecordsForType(report *dns.PropagationReport, recordType string) bool {
	for _, r := range report.Results {
		if r.Error != nil || r.Records == nil {
			continue
		}
		vals := dns.RecordValues(r.Records, recordType)
		if len(vals) > 0 {
			return true
		}
	}
	return false
}

func renderEstimation(est dns.Estimation) string {
	var b strings.Builder
	b.WriteString("\n    ")
	switch est.Status {
	case dns.FullyPropagated:
		b.WriteString(greenFg.Render(fmt.Sprintf("\u2713 Fully propagated across all resolvers (max TTL: %s)", dns.FormatTTL(est.MaxTTL))))
	case dns.InProgress:
		b.WriteString(warningFg.Render(fmt.Sprintf("\u23f1 Propagation in progress \u2014 %d/%d resolvers updated", est.Updated, est.Total)))
		for _, rem := range est.Remaining {
			b.WriteString("\n      ")
			b.WriteString(dimStyle.Render(fmt.Sprintf("- %s: still serving old values (TTL: %s)", rem.Resolver.Name, dns.FormatTTL(rem.TTL))))
		}
	}
	b.WriteString("\n")
	return b.String()
}

func renderPropagationTab(report *dns.PropagationReport) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("  " + cyanBold.Render(fmt.Sprintf("Propagation check for %s", report.Domain)))
	b.WriteString("\n")

	// Compute majority fingerprint per record type for highlighting
	for _, recordType := range dns.RecordTypes() {
		if !hasRecordsForType(report, recordType) {
			continue
		}

		b.WriteString(sectionHeader(recordType, 0))

		// Table header
		b.WriteString(fmt.Sprintf("    %-26s%-30s%s\n",
			boldStyle.Render("Resolver"),
			boldStyle.Render("Value"),
			boldStyle.Render("TTL"),
		))
		b.WriteString("    " + dimStyle.Render(strings.Repeat("\u2500", 64)) + "\n")

		// Compute estimation to know which resolvers are "behind"
		est := dns.Estimate(report.Results, recordType)
		remainingMap := make(map[string]bool)
		for _, rem := range est.Remaining {
			remainingMap[rem.Resolver.Name] = true
		}

		for _, result := range report.Results {
			resolverName := result.Resolver.Name

			if result.Error != nil {
				b.WriteString(fmt.Sprintf("    %-26s%s\n",
					resolverName,
					warningFg.Render(fmt.Sprintf("error: %v", result.Error)),
				))
				continue
			}

			vals := dns.RecordValues(result.Records, recordType)
			if len(vals) == 0 {
				b.WriteString(fmt.Sprintf("    %-26s%s\n",
					resolverName,
					dimStyle.Render("(no records)"),
				))
				continue
			}

			isOutdated := remainingMap[resolverName]

			for i, v := range vals {
				label := resolverName
				if i > 0 {
					label = ""
				}
				valuePart := v.Display
				ttlPart := dns.FormatTTL(v.TTL)

				if isOutdated {
					b.WriteString(fmt.Sprintf("    %s%s%s\n",
						redFg.Render(fmt.Sprintf("%-26s", label)),
						redFg.Render(fmt.Sprintf("%-30s", valuePart)),
						redFg.Render(ttlPart),
					))
				} else {
					b.WriteString(fmt.Sprintf("    %-26s%-30s%s\n",
						label,
						valuePart,
						dimStyle.Render(ttlPart),
					))
				}
			}
		}

		b.WriteString(renderEstimation(est))
	}

	b.WriteString("\n")
	return b.String()
}
