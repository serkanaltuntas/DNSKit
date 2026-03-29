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
	"testing"
)

func TestLookupAll_ExampleCom(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	r := NewResolver("", 0)
	rs, err := r.LookupAll(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("LookupAll(example.com) error: %v", err)
	}

	if len(rs.A) == 0 {
		t.Error("expected at least one A record for example.com")
	}
	if len(rs.NS) == 0 {
		t.Error("expected at least one NS record for example.com")
	}
	if len(rs.SOA) == 0 {
		t.Error("expected SOA record for example.com")
	}
}

func TestLookupAll_NXDOMAIN(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	r := NewResolver("", 0)
	_, err := r.LookupAll(context.Background(), "this-domain-does-not-exist-xyz123.example")
	if !errors.Is(err, ErrNXDOMAIN) {
		t.Errorf("expected ErrNXDOMAIN, got: %v", err)
	}
}

func TestNewResolver_Defaults(t *testing.T) {
	r := NewResolver("", 0)
	if r.Server != "8.8.8.8:53" {
		t.Errorf("expected default server 8.8.8.8:53, got %s", r.Server)
	}
	if r.Timeout != 5e9 {
		t.Errorf("expected default timeout 5s, got %v", r.Timeout)
	}
}
