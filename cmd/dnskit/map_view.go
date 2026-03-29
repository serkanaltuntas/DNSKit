package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/serkanaltuntas/dnskit/pkg/dns"
)

// World map projection parameters.
// Equirectangular: 170°W–190°E, 80°N–60°S → 120×36 character grid.
// col = (lon + 170) / 3       → each column ≈ 3° longitude
// row = (80 − lat) × 36 / 140 → each row ≈ 3.9° latitude
// Braille resolution: 240×144 pixels (2×4 dots per character).
const (
	mapWidth  = 120
	mapHeight = 36
	mapLatN   = 80.0
	mapLatS   = -60.0
	mapLonW   = -170.0
	mapLonE   = 190.0

	brailleW = mapWidth * 2  // 240 pixels
	brailleH = mapHeight * 4 // 144 pixels
)

// landRanges defines continent shapes as filled horizontal spans {row, colStart, colEnd}.
// Coastlines are extracted automatically via bilinear interpolation + edge detection.
//
// Key sea gaps for recognition:
//   - North Sea (UK ↔ Scandinavia)
//   - English Channel (UK ↔ France)
//   - Mediterranean (S.Europe ↔ N.Africa)
//   - Hudson Bay (inside Canada)
//   - Gulf of Bothnia (Sweden ↔ Finland)
//   - Baltic Sea (Scandinavia ↔ continental Europe)
//   - Black Sea (Turkey ↔ Ukraine)        — 5 cols wide
//   - Caspian Sea (Russia ↔ Iran)         — 2 cols wide
//   - Mediterranean (S.Europe ↔ N.Africa) — opens to Atlantic at Gibraltar
//   - Red Sea (Africa ↔ Arabia)           — 2 cols wide, opens to Gulf of Aden
//   - Persian Gulf (Arabia ↔ Iran)        — 1 col wide
//   - Gulf of Mexico (USA ↔ Mexico)       — with Yucatan & Cuba
//   - Sea of Japan (Korea ↔ Japan)
var landRanges = []struct{ row, c1, c2 int }{
	// ── Arctic & Northern Regions ──────────────────────────────────────
	// Row 0 (80°N): Arctic Canada, Northern Russia
	{0, 27, 33}, {0, 67, 108},
	// Row 1 (76°N): Arctic Canada, Greenland, Russia
	{1, 19, 35}, {1, 40, 46}, {1, 65, 114},
	// Row 2 (72°N): Alaska, Canada, Greenland, Iceland, Russia
	{2, 5, 10}, {2, 15, 35}, {2, 40, 47}, {2, 51, 52}, {2, 63, 115},
	// Row 3 (68°N): Alaska, Canada, Greenland, Iceland, Scandinavia, Russia
	{3, 4, 10}, {3, 12, 34}, {3, 41, 46}, {3, 51, 52}, {3, 58, 62}, {3, 63, 112},
	// Row 4 (64°N): Alaska, Canada, Greenland, UK tip, Scandinavia, [Gulf of Bothnia], Finland-Russia
	{4, 5, 9}, {4, 12, 33}, {4, 42, 45}, {4, 53, 54}, {4, 58, 62}, {4, 64, 107},
	// Row 5 (61°N): Alaska, Canada W, [Hudson Bay], Canada E, Greenland, Ireland/UK, Sweden, [Bothnia], Finland-Russia
	{5, 5, 8}, {5, 13, 24}, {5, 31, 33}, {5, 42, 44}, {5, 52, 55}, {5, 58, 62}, {5, 64, 105},

	// ── Europe & Northern Asia ─────────────────────────────────────────
	// Row 6 (57°N): Canada W, [Hudson Bay], Canada E, Ireland/UK, S.Scandinavia, [Baltic], Europe-Russia
	{6, 14, 24}, {6, 31, 33}, {6, 52, 55}, {6, 57, 61}, {6, 63, 103},
	// Row 7 (53°N): USA/Canada W, [Hudson Bay S.], Canada E, UK, Denmark, [Baltic/Kattegat], Europe-Russia
	{7, 16, 25}, {7, 31, 33}, {7, 53, 55}, {7, 57, 60}, {7, 62, 100},
	// Row 8 (49°N): USA/Canada, France-Germany-Poland-Ukraine-Russia-CentralAsia, Kamchatka
	{8, 17, 32}, {8, 56, 98}, {8, 109, 110},
	// Row 9 (45°N): USA, Iberia-France, Italy-Balkans, [Black Sea], Caucasus, [Caspian], CentralAsia-China, Kamchatka
	{9, 18, 31}, {9, 55, 58}, {9, 59, 65}, {9, 71, 72}, {9, 75, 98}, {9, 109, 110},

	// ── Mediterranean, Black Sea & Middle East ─────────────────────────
	// Row 10 (41°N): USA, Iberia, [W.Med], Italy, [Adriatic], Greece, [Black Sea], Turkey, [Caspian], Iran-CentralAsia-China, [Sea of Japan], Japan
	{10, 19, 30}, {10, 55, 58}, {10, 60, 61}, {10, 63, 65}, {10, 71, 72}, {10, 75, 100}, {10, 103, 105},
	// Row 11 (37°N): USA, S.Spain tip, [Mediterranean], Turkey, [Caspian], Iran-Afghan, China-Korea, [Sea of Japan], Japan
	{11, 20, 30}, {11, 56, 56}, {11, 65, 72}, {11, 75, 80}, {11, 83, 101}, {11, 103, 106},
	// Row 12 (33°N): USA, Morocco-Algeria, [Med south basin 57-66], Egypt-Levant-Iraq-Iran-Pakistan, India N, China
	{12, 21, 29}, {12, 53, 56}, {12, 67, 81}, {12, 82, 86}, {12, 88, 101},

	// ── Africa, Arabia, Americas & South Asia ──────────────────────────
	// Row 13 (29°N): Mexico-Texas-Louisiana, Florida, Sahara, [Red Sea 68-69], Arabia-Iraq-Iran-Pakistan, India, China
	{13, 18, 27}, {13, 30, 30}, {13, 52, 67}, {13, 70, 81}, {13, 82, 86}, {13, 88, 99},
	// Row 14 (26°N): Mexico, [Gulf of Mexico], Florida, Sahara, [Red Sea 68-69], Arabia, [Persian Gulf 74], Iran-Pakistan, India, Myanmar-China
	{14, 17, 23}, {14, 30, 30}, {14, 51, 67}, {14, 70, 73}, {14, 75, 81}, {14, 82, 87}, {14, 89, 97},
	// Row 15 (22°N): Mexico, Yucatan, [Yucatan Channel], Cuba, Sahara/Sahel, [Red Sea 68-69], Arabia/Yemen, India, SE Asia
	{15, 17, 23}, {15, 26, 28}, {15, 30, 31}, {15, 51, 67}, {15, 70, 76}, {15, 80, 86}, {15, 90, 95},
	// Row 16 (18°N): C.America, Caribbean islands, W.Africa-Sahel-Sudan, [Red Sea 68-69], Yemen, India, SE Asia, Philippines
	{16, 18, 22}, {16, 28, 29}, {16, 51, 67}, {16, 70, 74}, {16, 80, 85}, {16, 90, 95}, {16, 98, 99},
	// Row 17 (14°N): C.America, Africa to Djibouti, [Bab el-Mandeb + Gulf of Aden 68-72], Horn of Africa, India, SE Asia, Philippines
	{17, 19, 22}, {17, 51, 67}, {17, 73, 75}, {17, 80, 84}, {17, 90, 95}, {17, 98, 99},
	// Row 18 (10°N): Colombia, W/C/E Africa, [Gulf of Aden south 72-73], Somalia tip, India tip, SE Asia
	{18, 20, 24}, {18, 51, 71}, {18, 74, 75}, {18, 81, 84}, {18, 90, 94},

	// ── Equatorial Africa & South America ──────────────────────────────
	// Row 19 (6°N): Colombia/Venezuela, W/C/E Africa, Sri Lanka, Malay/Indonesia
	{19, 23, 31}, {19, 52, 74}, {19, 83, 84}, {19, 91, 99},
	// Row 20 (2°N): N.S.America, C/E Africa, Sumatra, Borneo
	{20, 24, 34}, {20, 54, 72}, {20, 92, 93}, {20, 95, 100},
	// Row 21 (−2°S): Brazil, C/E Africa, Sumatra, Borneo/Sulawesi
	{21, 25, 35}, {21, 56, 71}, {21, 92, 93}, {21, 95, 100},
	// Row 22 (−6°S): Brazil, E.Africa, Java, Borneo, Papua New Guinea
	{22, 26, 35}, {22, 58, 70}, {22, 93, 94}, {22, 96, 99}, {22, 104, 107},
	// Row 23 (−9°S): Brazil, E.Africa/Tanzania, Indonesia, PNG
	{23, 27, 35}, {23, 59, 70}, {23, 95, 97}, {23, 99, 99}, {23, 105, 108},
	// Row 24 (−13°S): Brazil, E.Africa/Mozambique, Timor, PNG
	{24, 27, 34}, {24, 61, 70}, {24, 97, 98}, {24, 106, 108},

	// ── Southern Africa, S.America & Australia ─────────────────────────
	// Row 25 (−17°S): S.America, SE Africa, Madagascar, N.Australia
	{25, 27, 33}, {25, 63, 69}, {25, 71, 73}, {25, 100, 110},
	// Row 26 (−21°S): S.America, S.Africa, Madagascar, Australia
	{26, 27, 32}, {26, 63, 69}, {26, 71, 73}, {26, 99, 112},
	// Row 27 (−25°S): S.America, S.Africa/Botswana, Australia
	{27, 26, 31}, {27, 63, 68}, {27, 98, 112},
	// Row 28 (−29°S): Chile/Argentina, S.Africa, Australia
	{28, 26, 30}, {28, 64, 67}, {28, 99, 111},
	// Row 29 (−33°S): Chile/Argentina, S.Africa tip, Australia
	{29, 26, 29}, {29, 64, 67}, {29, 100, 110},
	// Row 30 (−37°S): Chile/Argentina, S.Australia, New Zealand
	{30, 26, 29}, {30, 101, 109}, {30, 115, 116},
	// Row 31 (−41°S): Argentina, S.Australia, New Zealand
	{31, 26, 28}, {31, 102, 108}, {31, 115, 116},
	// Row 32 (−44°S): Patagonia, Tasmania
	{32, 26, 28}, {32, 108, 109},
	// Row 33 (−48°S): Patagonia
	{33, 26, 27},
	// Row 34 (−52°S): Tierra del Fuego
	{34, 26, 27},
}

// brailleDotBit maps pixel offset (dx 0-1, dy 0-3) within a braille cell to bit.
var brailleDotBit = [2][4]rune{
	{0x01, 0x02, 0x04, 0x40}, // left column  (dx=0): dots 1,2,3,7
	{0x08, 0x10, 0x20, 0x80}, // right column (dx=1): dots 4,5,6,8
}

var markerColors = []lipgloss.Style{
	lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
	lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true),
	lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true),
	lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true),
	lipgloss.NewStyle().Foreground(lipgloss.Color("201")).Bold(true),
	lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true),
	lipgloss.NewStyle().Foreground(lipgloss.Color("118")).Bold(true),
	lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Bold(true),
}

var (
	coastStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("65"))
	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("239"))
	legendLabel = lipgloss.NewStyle().Faint(true)
)

type mapMarker struct {
	row, col int
	symbol   string
	label    string
	style    lipgloss.Style
}

// buildLandGrid fills a boolean grid from landRanges.
func buildLandGrid() [mapHeight][mapWidth]bool {
	var grid [mapHeight][mapWidth]bool
	for _, r := range landRanges {
		for c := r.c1; c <= r.c2 && c < mapWidth; c++ {
			if r.row < mapHeight {
				grid[r.row][c] = true
			}
		}
	}
	return grid
}

// buildBrailleCoastline produces a high-resolution coastline grid using
// bilinear interpolation of the land grid for smooth edges.
func buildBrailleCoastline() [brailleH][brailleW]bool {
	land := buildLandGrid()

	// Bilinear-interpolated high-res land grid.
	var hires [brailleH][brailleW]bool
	for py := 0; py < brailleH; py++ {
		for px := 0; px < brailleW; px++ {
			gy := float64(py) * float64(mapHeight-1) / float64(brailleH-1)
			gx := float64(px) * float64(mapWidth-1) / float64(brailleW-1)

			y0, x0 := int(gy), int(gx)
			if y0 >= mapHeight-1 {
				y0 = mapHeight - 2
			}
			if x0 >= mapWidth-1 {
				x0 = mapWidth - 2
			}

			fy := gy - float64(y0)
			fx := gx - float64(x0)

			b2f := func(b bool) float64 {
				if b {
					return 1
				}
				return 0
			}

			val := b2f(land[y0][x0])*(1-fx)*(1-fy) +
				b2f(land[y0][x0+1])*fx*(1-fy) +
				b2f(land[y0+1][x0])*(1-fx)*fy +
				b2f(land[y0+1][x0+1])*fx*fy

			hires[py][px] = val > 0.5
		}
	}

	// Extract coastline: land pixels with at least one water neighbor.
	var coast [brailleH][brailleW]bool
	for py := 0; py < brailleH; py++ {
		for px := 0; px < brailleW; px++ {
			if !hires[py][px] {
				continue
			}
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dy == 0 && dx == 0 {
						continue
					}
					ny, nx := py+dy, px+dx
					if ny < 0 || ny >= brailleH || nx < 0 || nx >= brailleW || !hires[ny][nx] {
						coast[py][px] = true
					}
				}
			}
		}
	}

	return coast
}

func geoToMapPos(lat, lon float64) (row, col int) {
	if lat > mapLatN {
		lat = mapLatN
	}
	if lat < mapLatS {
		lat = mapLatS
	}
	for lon < mapLonW {
		lon += 360
	}
	for lon > mapLonE {
		lon -= 360
	}

	col = int((lon - mapLonW) / (mapLonE - mapLonW) * float64(mapWidth))
	row = int((mapLatN - lat) / (mapLatN - mapLatS) * float64(mapHeight))

	if col < 0 {
		col = 0
	} else if col >= mapWidth {
		col = mapWidth - 1
	}
	if row < 0 {
		row = 0
	} else if row >= mapHeight {
		row = mapHeight - 1
	}
	return row, col
}

func buildMarkers(geos []dns.GeoLocation) []mapMarker {
	symbols := []string{"❶", "❷", "❸", "❹", "❺", "❻", "❼", "❽"}
	occupied := make(map[[2]int]bool)
	var markers []mapMarker

	for i, geo := range geos {
		if i >= len(symbols) {
			break
		}
		row, col := geoToMapPos(geo.Lat, geo.Lon)

		key := [2]int{row, col}
		for occupied[key] {
			col++
			if col >= mapWidth {
				col = mapWidth - 1
				break
			}
			key = [2]int{row, col}
		}
		occupied[key] = true

		city := geo.City
		if city == "" {
			city = geo.Country
		}

		markers = append(markers, mapMarker{
			row:    row,
			col:    col,
			symbol: symbols[i],
			label:  fmt.Sprintf("%s  %s, %s", geo.IP, city, geo.Country),
			style:  markerColors[i%len(markerColors)],
		})
	}

	return markers
}

func renderMapTab(geos []dns.GeoLocation, domain string) string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(cyanBold.Render(fmt.Sprintf("  Server locations for %s", domain)))
	b.WriteString("\n\n")

	if len(geos) == 0 {
		b.WriteString(dimStyle.Render("  No geolocation data available.\n"))
		return b.String()
	}

	markers := buildMarkers(geos)
	coast := buildBrailleCoastline()

	markerAt := make(map[[2]int]*mapMarker)
	for i := range markers {
		markerAt[[2]int{markers[i].row, markers[i].col}] = &markers[i]
	}

	// Top border.
	b.WriteString("  ")
	b.WriteString(borderStyle.Render("╭" + strings.Repeat("─", mapWidth) + "╮"))
	b.WriteString("\n")

	// Map rows.
	for cy := 0; cy < mapHeight; cy++ {
		b.WriteString("  ")
		b.WriteString(borderStyle.Render("│"))

		for cx := 0; cx < mapWidth; cx++ {
			if m, ok := markerAt[[2]int{cy, cx}]; ok {
				b.WriteString(m.style.Render(m.symbol))
				continue
			}

			var bits rune
			for dx := 0; dx < 2; dx++ {
				for dy := 0; dy < 4; dy++ {
					px := cx*2 + dx
					py := cy*4 + dy
					if px < brailleW && py < brailleH && coast[py][px] {
						bits |= brailleDotBit[dx][dy]
					}
				}
			}

			if bits == 0 {
				b.WriteString(" ")
			} else {
				b.WriteString(coastStyle.Render(string(0x2800 + bits)))
			}
		}

		b.WriteString(borderStyle.Render("│"))
		b.WriteString("\n")
	}

	// Bottom border.
	b.WriteString("  ")
	b.WriteString(borderStyle.Render("╰" + strings.Repeat("─", mapWidth) + "╯"))
	b.WriteString("\n\n")

	// Legend.
	for _, m := range markers {
		b.WriteString(fmt.Sprintf("    %s  %s\n",
			m.style.Render(m.symbol),
			legendLabel.Render(m.label),
		))
	}

	b.WriteString("\n")
	return b.String()
}
