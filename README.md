# DNSKit

> Sleek DNS record lookup and propagation tracker written in Go.

DNSKit is a modern toolkit and interactive Terminal User Interface (TUI) for querying DNS records and checking real-time propagation status across multiple global resolvers.

## Features

- **Comprehensive Lookups**: Quickly query a wide range of DNS record types (A, AAAA, CNAME, MX, NS, TXT, SOA, SRV, CAA, PTR).
- **Interactive TUI**: A sleek, dual-tab console interface built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), allowing you to easily switch between standard lookups and propagation tracking.
- **Propagation Tracker**: Checks DNS resolution in parallel across major global providers (Google, Cloudflare, Quad9, OpenDNS) and highlights mismatches out of the box.
- **Smart Estimation**: Estimates propagation completion times by calculating remaining TTL values.
- **Go Library**: Can be integrated into your own Go projects via the `pkg/dns` package.

## Installation

Ensure you have Go 1.25.5 or later installed, then run:

```bash
go install github.com/serkanaltuntas/dnskit/cmd/dnskit@latest
```

## Usage

Simply pass the domain you want to query:

```bash
dnskit example.com
```

### Options

Navigation inside the TUI is straightforward. Use `Tab` or `←/→` to switch between tabs, `↑/↓` or `j/k` to scroll through records, and `q` or `Ctrl+C` to quit.

You can also customize the query behavior through CLI flags:

```text
Usage: dnskit [flags] <domain>

Flags:
  -server string
        DNS server address (host:port) for the Records tab (default "8.8.8.8:53")
  -timeout duration
        query timeout (default 5s)
```

## Build from Source

```bash
git clone https://github.com/serkanaltuntas/dnskit.git
cd dnskit
go build -o dnskit ./cmd/dnskit
```

## License

This project is licensed under the [GNU General Public License v3.0](LICENSE).
