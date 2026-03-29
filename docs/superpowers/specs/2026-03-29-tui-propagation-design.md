# DNSKit TUI + Propagation Check Design

## Summary

Transform DNSKit from plain stdout output to a Bubble Tea TUI with two tabs:
1. **DNS Records** — current query results in TUI format
2. **Propagation** — parallel multi-resolver comparison with estimation

## Tab 1: DNS Records

Renders the existing DNS query output inside the TUI. Layout matches current output:

- Header: domain name + server address
- Sections per record type (A, AAAA, CNAME, MX, NS, TXT, SOA, SRV, CAA, PTR)
- Only types with results are shown
- TTL formatting unchanged (`300 (5m)`)

Data source: the resolver specified by `-server` flag (default `8.8.8.8:53`).

## Tab 2: Propagation

Queries all record types from 4 fixed resolvers in parallel. Displays a table per record type that has results from any resolver.

### Resolver List (Fixed)

| Resolver   | Address          | Label      |
|------------|------------------|------------|
| Google     | 8.8.8.8:53       | Google     |
| Cloudflare | 1.1.1.1:53       | Cloudflare |
| Quad9      | 9.9.9.9:53       | Quad9      |
| OpenDNS    | 208.67.222.222:53| OpenDNS    |

### Table Format

Per record type:

```
A Record
┌──────────────────────┬───────────────────┬─────────────┐
│ Resolver             │ Value             │ TTL         │
├──────────────────────┼───────────────────┼─────────────┤
│ 8.8.8.8 (Google)     │ 34.160.255.155    │ 300 (5m)    │
│ 1.1.1.1 (Cloudflare) │ 34.160.255.155    │ 245 (4m 5s) │
│ 9.9.9.9 (Quad9)      │ 34.160.255.155    │ 180 (3m)    │
│ 208.67.222.222 (Open)│ 34.160.255.155    │ 300 (5m)    │
└──────────────────────┴───────────────────┴─────────────┘
```

### Value Comparison Highlighting

When resolver values differ (indicating propagation in progress):
- **Majority value** — rendered normally
- **Minority/different value** — highlighted in red to draw attention

### Propagation Estimation

Below each record type table, one of three states is shown:

**All values identical:**
```
✓ Fully propagated across all resolvers (max TTL: 300s / 5m)
```

**Values differ (propagation in progress):**
```
⏱ Propagation in progress — 3/4 resolvers updated
   Remaining: Quad9 (TTL: 180s, ~3 minutes)
```

**Estimation logic:**
- Max TTL across all resolvers for that record type = worst-case propagation time
- When values differ, show which resolvers still have old values and their remaining TTL
- Remaining TTL is approximate — it's the TTL returned by the resolver, representing time until cache expiry

**Error from a resolver:**
```
⚠ 9.9.9.9 (Quad9): timeout
```

Errors are shown inline in the table row, not counted as "different value."

## Navigation

| Key            | Action           |
|----------------|------------------|
| `Tab` or `←/→` | Switch tabs      |
| `↑/↓` or `j/k` | Scroll content   |
| `q` or `Ctrl+C` | Quit             |

Active tab is visually highlighted in the tab bar.

## Technical Architecture

### Dependencies

**Add:**
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — styling
- `github.com/charmbracelet/bubbles/table` — table component
- `github.com/charmbracelet/bubbles/viewport` — scrollable content

**Remove:**
- `github.com/fatih/color` — replaced by Lip Gloss

### Package Structure

```
pkg/dns/
  resolver.go       # existing — unchanged
  records.go        # existing — unchanged
  propagation.go    # NEW — multi-resolver parallel query + estimation logic

cmd/dnskit/
  main.go           # rewrite — Bubble Tea TUI replaces stdout printing
```

### Propagation Logic (`pkg/dns/propagation.go`)

```go
type ResolverInfo struct {
    Name    string // "Google", "Cloudflare", etc.
    Address string // "8.8.8.8:53"
}

type PropagationResult struct {
    Resolver  ResolverInfo
    Records   RecordSet
    Error     error
    QueryTime time.Duration
}

type PropagationReport struct {
    Domain  string
    Results []PropagationResult
}

// CheckPropagation queries all fixed resolvers in parallel
func CheckPropagation(domain string, timeout time.Duration) PropagationReport
```

- Uses goroutines + `sync.WaitGroup` for parallel queries
- Each resolver gets its own `Resolver` instance
- Results collected via channel
- Timeout applies per-resolver

### Estimation Logic

```go
type EstimationStatus int

const (
    FullyPropagated EstimationStatus = iota
    InProgress
    Error
)

type Estimation struct {
    Status       EstimationStatus
    MaxTTL       uint32
    Updated      int // count of resolvers with majority value
    Total        int // total resolver count
    Remaining    []RemainingResolver // resolvers with old/different values
}

type RemainingResolver struct {
    Resolver ResolverInfo
    TTL      uint32
    Value    string
}
```

For each record type: group resolver results by value, identify majority, flag minority as "not yet propagated."

### TUI Model (`cmd/dnskit/main.go`)

Bubble Tea model with:
- `activeTab` (0 = Records, 1 = Propagation)
- `recordsViewport` (scrollable DNS records content)
- `propagationViewport` (scrollable propagation content)
- `recordSet` (from primary resolver)
- `propagationReport` (from all resolvers)

Both queries run concurrently on startup. Tab content is pre-rendered into viewports.

## CLI Interface

No changes to flags:

```
dnskit [flags] <domain>

Flags:
  -server string     DNS server for Records tab (default "8.8.8.8:53")
  -timeout duration  query timeout (default 5s)
```

The `-server` flag affects only the Records tab. Propagation tab always uses the 4 fixed resolvers.

## What Stays the Same

- `pkg/dns/resolver.go` — untouched
- `pkg/dns/records.go` — untouched
- All record types and TTL formatting
- CLI flags (`-server`, `-timeout`)
- `pkg/dns` remains usable as a library
