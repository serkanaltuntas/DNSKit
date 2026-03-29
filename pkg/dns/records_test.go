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

import "testing"

func TestFormatTTL(t *testing.T) {
	tests := []struct {
		input uint32
		want  string
	}{
		{0, "0"},
		{30, "30"},
		{60, "60 (1m)"},
		{300, "300 (5m)"},
		{3600, "3600 (1h)"},
		{7200, "7200 (2h)"},
		{86400, "86400 (1d)"},
		{172800, "172800 (2d)"},
		{1209600, "1209600 (14d)"},
		{3601, "3601"},
		{90, "90"},
	}

	for _, tt := range tests {
		got := FormatTTL(tt.input)
		if got != tt.want {
			t.Errorf("FormatTTL(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
