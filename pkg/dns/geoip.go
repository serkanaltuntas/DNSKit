package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// GeoLocation holds geolocation data for an IP address.
type GeoLocation struct {
	IP      string
	Country string
	City    string
	Lat     float64
	Lon     float64
}

type ipAPIResponse struct {
	Status  string  `json:"status"`
	Country string  `json:"country"`
	City    string  `json:"city"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Query   string  `json:"query"`
}

// LookupGeo resolves geolocation for a single IP using ip-api.com.
func LookupGeo(ctx context.Context, ip string) (*GeoLocation, error) {
	// Skip private/loopback addresses.
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() {
		return nil, fmt.Errorf("non-routable IP: %s", ip)
	}

	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,city,lat,lon,query", ip)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data ipAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if data.Status != "success" {
		return nil, fmt.Errorf("geo lookup failed for %s", ip)
	}

	return &GeoLocation{
		IP:      data.Query,
		Country: data.Country,
		City:    data.City,
		Lat:     data.Lat,
		Lon:     data.Lon,
	}, nil
}

// LookupGeoAll looks up geolocation for all unique A/AAAA IPs in a RecordSet.
func LookupGeoAll(ctx context.Context, rs *RecordSet) []GeoLocation {
	if rs == nil {
		return nil
	}

	// Collect unique IPs.
	seen := make(map[string]bool)
	var ips []string
	for _, r := range rs.A {
		if !seen[r.Address] {
			seen[r.Address] = true
			ips = append(ips, r.Address)
		}
	}
	for _, r := range rs.AAAA {
		if !seen[r.Address] {
			seen[r.Address] = true
			ips = append(ips, r.Address)
		}
	}

	if len(ips) == 0 {
		return nil
	}

	var mu sync.Mutex
	var results []GeoLocation
	var wg sync.WaitGroup

	for _, ip := range ips {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			// Rate limit: ip-api.com allows 45 req/min for free tier.
			geo, err := LookupGeo(ctx, addr)
			if err != nil {
				return
			}
			mu.Lock()
			results = append(results, *geo)
			mu.Unlock()
		}(ip)
	}

	wg.Wait()
	return results
}
