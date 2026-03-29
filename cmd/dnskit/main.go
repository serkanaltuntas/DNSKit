// Copyright (C) 2026 Serkan Altuntaş
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/serkanaltuntas/dnskit/pkg/dns"
)

func main() {
	server := flag.String("server", "8.8.8.8:53", "DNS server address (host:port)")
	timeout := flag.Duration("timeout", 5*time.Second, "query timeout")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: dnskit [flags] <domain>\n\nFlags:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	domain := flag.Arg(0)

	ctx := context.Background()

	type recordResult struct {
		rs  *dns.RecordSet
		err error
	}
	type propagationResult struct {
		report *dns.PropagationReport
	}

	rCh := make(chan recordResult, 1)
	pCh := make(chan propagationResult, 1)

	go func() {
		resolver := dns.NewResolver(*server, *timeout)
		rs, err := resolver.LookupAll(ctx, domain)
		rCh <- recordResult{rs, err}
	}()

	go func() {
		report := dns.CheckPropagation(ctx, domain, *timeout)
		pCh <- propagationResult{report}
	}()

	rr := <-rCh
	if rr.err != nil {
		fmt.Fprintf(os.Stderr, "\n  Error: %v\n\n", rr.err)
		os.Exit(1)
	}

	// Fetch geolocation for IPs (runs in background while propagation finishes).
	geoCh := make(chan []dns.GeoLocation, 1)
	go func() {
		geoCh <- dns.LookupGeoAll(ctx, rr.rs)
	}()

	pr := <-pCh
	geoData := <-geoCh

	m := newModel(rr.rs, pr.report, geoData, *server)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
