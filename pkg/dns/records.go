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

import "fmt"

// RecordSet holds all DNS results for a single domain.
type RecordSet struct {
	Domain string
	A      []ARecord
	AAAA   []AAAARecord
	CNAME  []CNAMERecord
	MX     []MXRecord
	NS     []NSRecord
	TXT    []TXTRecord
	SOA    []SOARecord
	SRV    []SRVRecord
	CAA    []CAARecord
	PTR    []PTRRecord
	Errors map[string]error
}

type ARecord struct {
	Address string
	TTL     uint32
}

type AAAARecord struct {
	Address string
	TTL     uint32
}

type CNAMERecord struct {
	Target string
	TTL    uint32
}

type MXRecord struct {
	Host       string
	Preference uint16
	TTL        uint32
}

type NSRecord struct {
	Nameserver string
	TTL        uint32
}

type TXTRecord struct {
	Text string
	TTL  uint32
}

type SOARecord struct {
	PrimaryNS string
	Mailbox   string
	Serial    uint32
	Refresh   uint32
	Retry     uint32
	Expire    uint32
	MinTTL    uint32
	TTL       uint32
}

type SRVRecord struct {
	Target   string
	Port     uint16
	Priority uint16
	Weight   uint16
	TTL      uint32
}

type CAARecord struct {
	Flag  uint8
	Tag   string
	Value string
	TTL   uint32
}

type PTRRecord struct {
	Host string
	TTL  uint32
}

// FormatTTL converts a TTL in seconds to a human-friendly string.
func FormatTTL(seconds uint32) string {
	switch {
	case seconds >= 86400 && seconds%86400 == 0:
		return fmt.Sprintf("%d (%dd)", seconds, seconds/86400)
	case seconds >= 3600 && seconds%3600 == 0:
		return fmt.Sprintf("%d (%dh)", seconds, seconds/3600)
	case seconds >= 60 && seconds%60 == 0:
		return fmt.Sprintf("%d (%dm)", seconds, seconds/60)
	default:
		return fmt.Sprintf("%d", seconds)
	}
}
