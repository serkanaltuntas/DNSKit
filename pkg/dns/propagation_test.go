package dns

import (
	"context"
	"testing"
	"time"
)

func TestCheckPropagation_ExampleCom(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test in short mode")
	}

	report := CheckPropagation(context.Background(), "example.com", 5*time.Second)

	if report.Domain != "example.com" {
		t.Errorf("expected domain example.com, got %s", report.Domain)
	}

	if len(report.Results) != len(DefaultResolvers) {
		t.Errorf("expected %d results, got %d", len(DefaultResolvers), len(report.Results))
	}

	for _, r := range report.Results {
		if r.Error != nil {
			t.Errorf("resolver %s (%s) returned error: %v", r.Resolver.Name, r.Resolver.Address, r.Error)
			continue
		}
		if len(r.Records.A) == 0 {
			t.Errorf("resolver %s: expected at least one A record", r.Resolver.Name)
		}
	}
}

func TestDefaultResolvers(t *testing.T) {
	expected := []struct {
		name    string
		address string
	}{
		{"Google", "8.8.8.8:53"},
		{"Cloudflare", "1.1.1.1:53"},
		{"Quad9", "9.9.9.9:53"},
		{"OpenDNS", "208.67.222.222:53"},
	}

	if len(DefaultResolvers) != len(expected) {
		t.Fatalf("expected %d resolvers, got %d", len(expected), len(DefaultResolvers))
	}

	for i, e := range expected {
		if DefaultResolvers[i].Name != e.name {
			t.Errorf("resolver %d: expected name %s, got %s", i, e.name, DefaultResolvers[i].Name)
		}
		if DefaultResolvers[i].Address != e.address {
			t.Errorf("resolver %d: expected address %s, got %s", i, e.address, DefaultResolvers[i].Address)
		}
	}
}
