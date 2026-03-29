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

package dns

import (
	"context"
	"errors"
	"strings"
	"time"

	mdns "github.com/miekg/dns"
)

var (
	ErrNXDOMAIN      = errors.New("domain does not exist (NXDOMAIN)")
	ErrServerFailure = errors.New("server failure (SERVFAIL)")
	ErrQueryRefused  = errors.New("query refused")
)

// Resolver performs DNS lookups against a specified server.
type Resolver struct {
	Server  string
	Timeout time.Duration
}

// NewResolver creates a Resolver with sensible defaults.
func NewResolver(server string, timeout time.Duration) *Resolver {
	if server == "" {
		server = "8.8.8.8:53"
	}
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Resolver{Server: server, Timeout: timeout}
}

// LookupAll queries all supported record types for the given domain.
func (r *Resolver) LookupAll(ctx context.Context, domain string) (*RecordSet, error) {
	fqdn := mdns.Fqdn(domain)

	rs := &RecordSet{
		Domain: strings.TrimSuffix(fqdn, "."),
		Errors: make(map[string]error),
	}

	types := []struct {
		name  string
		qtype uint16
	}{
		{"A", mdns.TypeA},
		{"AAAA", mdns.TypeAAAA},
		{"CNAME", mdns.TypeCNAME},
		{"MX", mdns.TypeMX},
		{"NS", mdns.TypeNS},
		{"TXT", mdns.TypeTXT},
		{"SOA", mdns.TypeSOA},
		{"SRV", mdns.TypeSRV},
		{"CAA", mdns.TypeCAA},
		{"PTR", mdns.TypePTR},
	}

	nxdomainCount := 0

	for _, t := range types {
		rrs, err := r.query(ctx, fqdn, t.qtype)
		if err != nil {
			if errors.Is(err, ErrNXDOMAIN) {
				nxdomainCount++
			} else {
				rs.Errors[t.name] = err
			}
			continue
		}
		r.parseRecords(rs, rrs)
	}

	if nxdomainCount == len(types) {
		return nil, ErrNXDOMAIN
	}

	return rs, nil
}

func (r *Resolver) query(ctx context.Context, fqdn string, qtype uint16) ([]mdns.RR, error) {
	msg := new(mdns.Msg)
	msg.SetQuestion(fqdn, qtype)
	msg.RecursionDesired = true

	client := &mdns.Client{Timeout: r.Timeout}
	resp, _, err := client.ExchangeContext(ctx, msg, r.Server)
	if err != nil {
		return nil, err
	}

	switch resp.Rcode {
	case mdns.RcodeNameError:
		return nil, ErrNXDOMAIN
	case mdns.RcodeServerFailure:
		return nil, ErrServerFailure
	case mdns.RcodeRefused:
		return nil, ErrQueryRefused
	}

	return resp.Answer, nil
}

func (r *Resolver) parseRecords(rs *RecordSet, rrs []mdns.RR) {
	for _, rr := range rrs {
		switch v := rr.(type) {
		case *mdns.A:
			rs.A = append(rs.A, ARecord{
				Address: v.A.String(),
				TTL:     v.Hdr.Ttl,
			})
		case *mdns.AAAA:
			rs.AAAA = append(rs.AAAA, AAAARecord{
				Address: v.AAAA.String(),
				TTL:     v.Hdr.Ttl,
			})
		case *mdns.CNAME:
			rs.CNAME = append(rs.CNAME, CNAMERecord{
				Target: v.Target,
				TTL:    v.Hdr.Ttl,
			})
		case *mdns.MX:
			rs.MX = append(rs.MX, MXRecord{
				Host:       v.Mx,
				Preference: v.Preference,
				TTL:        v.Hdr.Ttl,
			})
		case *mdns.NS:
			rs.NS = append(rs.NS, NSRecord{
				Nameserver: v.Ns,
				TTL:        v.Hdr.Ttl,
			})
		case *mdns.TXT:
			rs.TXT = append(rs.TXT, TXTRecord{
				Text: strings.Join(v.Txt, ""),
				TTL:  v.Hdr.Ttl,
			})
		case *mdns.SOA:
			rs.SOA = append(rs.SOA, SOARecord{
				PrimaryNS: v.Ns,
				Mailbox:   v.Mbox,
				Serial:    v.Serial,
				Refresh:   v.Refresh,
				Retry:     v.Retry,
				Expire:    v.Expire,
				MinTTL:    v.Minttl,
				TTL:       v.Hdr.Ttl,
			})
		case *mdns.SRV:
			rs.SRV = append(rs.SRV, SRVRecord{
				Target:   v.Target,
				Port:     v.Port,
				Priority: v.Priority,
				Weight:   v.Weight,
				TTL:      v.Hdr.Ttl,
			})
		case *mdns.CAA:
			rs.CAA = append(rs.CAA, CAARecord{
				Flag:  v.Flag,
				Tag:   v.Tag,
				Value: v.Value,
				TTL:   v.Hdr.Ttl,
			})
		case *mdns.PTR:
			rs.PTR = append(rs.PTR, PTRRecord{
				Host: v.Ptr,
				TTL:  v.Hdr.Ttl,
			})
		}
	}
}
