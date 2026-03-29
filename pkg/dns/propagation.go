package dns

import (
	"context"
	"sync"
	"time"
)

// ResolverInfo identifies a named DNS resolver.
type ResolverInfo struct {
	Name    string
	Address string
}

// DefaultResolvers is the fixed set of public resolvers used for propagation checks.
var DefaultResolvers = []ResolverInfo{
	{Name: "Google", Address: "8.8.8.8:53"},
	{Name: "Cloudflare", Address: "1.1.1.1:53"},
	{Name: "Quad9", Address: "9.9.9.9:53"},
	{Name: "OpenDNS", Address: "208.67.222.222:53"},
}

// PropagationResult holds the query result from a single resolver.
type PropagationResult struct {
	Resolver  ResolverInfo
	Records   *RecordSet
	Error     error
	QueryTime time.Duration
}

// PropagationReport holds results from all resolvers for a domain.
type PropagationReport struct {
	Domain  string
	Results []PropagationResult
}

// CheckPropagation queries all default resolvers in parallel and returns a report.
func CheckPropagation(ctx context.Context, domain string, timeout time.Duration) *PropagationReport {
	report := &PropagationReport{
		Domain:  domain,
		Results: make([]PropagationResult, len(DefaultResolvers)),
	}

	var wg sync.WaitGroup
	wg.Add(len(DefaultResolvers))

	for i, ri := range DefaultResolvers {
		go func(idx int, info ResolverInfo) {
			defer wg.Done()

			r := NewResolver(info.Address, timeout)
			start := time.Now()
			rs, err := r.LookupAll(ctx, domain)
			elapsed := time.Since(start)

			report.Results[idx] = PropagationResult{
				Resolver:  info,
				Records:   rs,
				Error:     err,
				QueryTime: elapsed,
			}
		}(i, ri)
	}

	wg.Wait()
	return report
}
