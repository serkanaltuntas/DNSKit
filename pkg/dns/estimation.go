package dns

import (
	"fmt"
	"sort"
	"strings"
)

// EstimationStatus represents the propagation state for a record type.
type EstimationStatus int

const (
	FullyPropagated EstimationStatus = iota
	InProgress
)

// RemainingResolver describes a resolver that hasn't propagated yet.
type RemainingResolver struct {
	Resolver ResolverInfo
	TTL      uint32
	Value    string
}

// Estimation holds the propagation estimate for one record type.
type Estimation struct {
	Status    EstimationStatus
	MaxTTL    uint32
	Updated   int
	Total     int
	Remaining []RemainingResolver
}

// TypedValue is a displayable record value with its TTL.
type TypedValue struct {
	Display string
	TTL     uint32
}

// RecordTypes returns all supported DNS record type names in display order.
func RecordTypes() []string {
	return []string{"A", "AAAA", "CNAME", "MX", "NS", "TXT", "SOA", "SRV", "CAA", "PTR"}
}

// RecordValues extracts displayable values from a RecordSet for the given type.
func RecordValues(rs *RecordSet, recordType string) []TypedValue {
	if rs == nil {
		return nil
	}

	switch recordType {
	case "A":
		vals := make([]TypedValue, len(rs.A))
		for i, r := range rs.A {
			vals[i] = TypedValue{Display: r.Address, TTL: r.TTL}
		}
		return vals
	case "AAAA":
		vals := make([]TypedValue, len(rs.AAAA))
		for i, r := range rs.AAAA {
			vals[i] = TypedValue{Display: r.Address, TTL: r.TTL}
		}
		return vals
	case "CNAME":
		vals := make([]TypedValue, len(rs.CNAME))
		for i, r := range rs.CNAME {
			vals[i] = TypedValue{Display: r.Target, TTL: r.TTL}
		}
		return vals
	case "MX":
		vals := make([]TypedValue, len(rs.MX))
		for i, r := range rs.MX {
			vals[i] = TypedValue{Display: fmt.Sprintf("%s (priority: %d)", r.Host, r.Preference), TTL: r.TTL}
		}
		return vals
	case "NS":
		vals := make([]TypedValue, len(rs.NS))
		for i, r := range rs.NS {
			vals[i] = TypedValue{Display: r.Nameserver, TTL: r.TTL}
		}
		return vals
	case "TXT":
		vals := make([]TypedValue, len(rs.TXT))
		for i, r := range rs.TXT {
			vals[i] = TypedValue{Display: r.Text, TTL: r.TTL}
		}
		return vals
	case "SOA":
		vals := make([]TypedValue, len(rs.SOA))
		for i, r := range rs.SOA {
			vals[i] = TypedValue{Display: fmt.Sprintf("%s %s %d", r.PrimaryNS, r.Mailbox, r.Serial), TTL: r.TTL}
		}
		return vals
	case "SRV":
		vals := make([]TypedValue, len(rs.SRV))
		for i, r := range rs.SRV {
			vals[i] = TypedValue{Display: fmt.Sprintf("%s:%d (pri:%d w:%d)", r.Target, r.Port, r.Priority, r.Weight), TTL: r.TTL}
		}
		return vals
	case "CAA":
		vals := make([]TypedValue, len(rs.CAA))
		for i, r := range rs.CAA {
			vals[i] = TypedValue{Display: fmt.Sprintf("%s %s", r.Tag, r.Value), TTL: r.TTL}
		}
		return vals
	case "PTR":
		vals := make([]TypedValue, len(rs.PTR))
		for i, r := range rs.PTR {
			vals[i] = TypedValue{Display: r.Host, TTL: r.TTL}
		}
		return vals
	default:
		return nil
	}
}

// Estimate computes the propagation estimation for a given record type across resolver results.
func Estimate(results []PropagationResult, recordType string) Estimation {
	type resolverEntry struct {
		resolver    ResolverInfo
		fingerprint string
		maxTTL      uint32
	}

	var entries []resolverEntry

	for _, r := range results {
		if r.Error != nil || r.Records == nil {
			continue
		}
		vals := RecordValues(r.Records, recordType)
		fp := fingerprint(vals)
		var maxTTL uint32
		for _, v := range vals {
			if v.TTL > maxTTL {
				maxTTL = v.TTL
			}
		}
		entries = append(entries, resolverEntry{
			resolver:    r.Resolver,
			fingerprint: fp,
			maxTTL:      maxTTL,
		})
	}

	if len(entries) == 0 {
		return Estimation{Status: FullyPropagated}
	}

	// Find majority fingerprint
	fpCount := make(map[string]int)
	for _, e := range entries {
		fpCount[e.fingerprint]++
	}

	majorityFP := ""
	majorityCount := 0
	for fp, count := range fpCount {
		if count > majorityCount {
			majorityCount = count
			majorityFP = fp
		}
	}

	var globalMaxTTL uint32
	for _, e := range entries {
		if e.maxTTL > globalMaxTTL {
			globalMaxTTL = e.maxTTL
		}
	}

	est := Estimation{
		MaxTTL:  globalMaxTTL,
		Updated: majorityCount,
		Total:   len(entries),
	}

	if majorityCount == len(entries) {
		est.Status = FullyPropagated
		return est
	}

	// When no strict majority exists (no fingerprint held by more than half
	// the resolvers), the differences are likely anycast / load-balanced
	// responses rather than stale caches. Treat as fully propagated.
	if majorityCount*2 <= len(entries) {
		est.Status = FullyPropagated
		est.Updated = len(entries)
		return est
	}

	est.Status = InProgress
	for _, e := range entries {
		if e.fingerprint != majorityFP {
			est.Remaining = append(est.Remaining, RemainingResolver{
				Resolver: e.resolver,
				TTL:      e.maxTTL,
				Value:    e.fingerprint,
			})
		}
	}

	return est
}

// fingerprint creates a comparable string from a set of typed values.
func fingerprint(vals []TypedValue) string {
	displays := make([]string, len(vals))
	for i, v := range vals {
		displays[i] = v.Display
	}
	sort.Strings(displays)
	return strings.Join(displays, "|")
}
