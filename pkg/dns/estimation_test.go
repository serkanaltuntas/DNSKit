package dns

import (
	"testing"
)

func TestRecordValues_A(t *testing.T) {
	rs := &RecordSet{
		A: []ARecord{
			{Address: "1.2.3.4", TTL: 300},
			{Address: "5.6.7.8", TTL: 300},
		},
	}
	vals := RecordValues(rs, "A")
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d", len(vals))
	}
	if vals[0].Display != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got %s", vals[0].Display)
	}
}

func TestRecordValues_Empty(t *testing.T) {
	rs := &RecordSet{}
	vals := RecordValues(rs, "A")
	if len(vals) != 0 {
		t.Errorf("expected 0 values, got %d", len(vals))
	}
}

func TestRecordValues_Nil(t *testing.T) {
	vals := RecordValues(nil, "A")
	if vals != nil {
		t.Errorf("expected nil for nil RecordSet, got %v", vals)
	}
}

func TestEstimate_FullyPropagated(t *testing.T) {
	results := []PropagationResult{
		{Resolver: ResolverInfo{Name: "Google", Address: "8.8.8.8:53"}, Records: &RecordSet{
			A: []ARecord{{Address: "1.2.3.4", TTL: 300}},
		}},
		{Resolver: ResolverInfo{Name: "Cloudflare", Address: "1.1.1.1:53"}, Records: &RecordSet{
			A: []ARecord{{Address: "1.2.3.4", TTL: 200}},
		}},
	}

	est := Estimate(results, "A")
	if est.Status != FullyPropagated {
		t.Errorf("expected FullyPropagated, got %d", est.Status)
	}
	if est.MaxTTL != 300 {
		t.Errorf("expected MaxTTL 300, got %d", est.MaxTTL)
	}
	if est.Updated != 2 || est.Total != 2 {
		t.Errorf("expected 2/2, got %d/%d", est.Updated, est.Total)
	}
}

func TestEstimate_InProgress(t *testing.T) {
	results := []PropagationResult{
		{Resolver: ResolverInfo{Name: "Google"}, Records: &RecordSet{
			A: []ARecord{{Address: "1.2.3.4", TTL: 300}},
		}},
		{Resolver: ResolverInfo{Name: "Cloudflare"}, Records: &RecordSet{
			A: []ARecord{{Address: "1.2.3.4", TTL: 200}},
		}},
		{Resolver: ResolverInfo{Name: "Quad9"}, Records: &RecordSet{
			A: []ARecord{{Address: "9.9.9.9", TTL: 180}},
		}},
	}

	est := Estimate(results, "A")
	if est.Status != InProgress {
		t.Errorf("expected InProgress, got %d", est.Status)
	}
	if est.Updated != 2 {
		t.Errorf("expected 2 updated, got %d", est.Updated)
	}
	if est.Total != 3 {
		t.Errorf("expected 3 total, got %d", est.Total)
	}
	if len(est.Remaining) != 1 {
		t.Fatalf("expected 1 remaining, got %d", len(est.Remaining))
	}
	if est.Remaining[0].Resolver.Name != "Quad9" {
		t.Errorf("expected remaining resolver Quad9, got %s", est.Remaining[0].Resolver.Name)
	}
	if est.Remaining[0].TTL != 180 {
		t.Errorf("expected remaining TTL 180, got %d", est.Remaining[0].TTL)
	}
}

func TestEstimate_WithErrors(t *testing.T) {
	results := []PropagationResult{
		{Resolver: ResolverInfo{Name: "Google"}, Records: &RecordSet{
			A: []ARecord{{Address: "1.2.3.4", TTL: 300}},
		}},
		{Resolver: ResolverInfo{Name: "Cloudflare"}, Error: ErrServerFailure},
	}

	est := Estimate(results, "A")
	if est.Status != FullyPropagated {
		t.Errorf("expected FullyPropagated (error excluded), got %d", est.Status)
	}
	if est.Total != 1 {
		t.Errorf("expected 1 total (error excluded), got %d", est.Total)
	}
}

func TestEstimate_NoRecords(t *testing.T) {
	results := []PropagationResult{
		{Resolver: ResolverInfo{Name: "Google"}, Records: &RecordSet{}},
		{Resolver: ResolverInfo{Name: "Cloudflare"}, Records: &RecordSet{}},
	}

	est := Estimate(results, "A")
	if est.Status != FullyPropagated {
		t.Errorf("expected FullyPropagated for no records, got %d", est.Status)
	}
}

func TestRecordTypes(t *testing.T) {
	types := RecordTypes()
	if len(types) != 10 {
		t.Errorf("expected 10 record types, got %d", len(types))
	}
	if types[0] != "A" {
		t.Errorf("expected first type to be A, got %s", types[0])
	}
}
