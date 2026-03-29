package dns

import "testing"

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		hostname string
		wantName string
	}{
		{"aspmx.l.google.com.", "Google"},
		{"ns1.google.com", "Google"},
		{"ns-cloud-a1.googledomains.com.", "Google"},
		{"d111111abcdef8.cloudfront.net.", "AWS"},
		{"s3.amazonaws.com", "AWS"},
		{"ns-123.awsdns-45.com.", "AWS"},
		{"elb.us-east-1.amazonaws.com", "AWS"},
		{"ada.ns.cloudflare.com.", "Cloudflare"},
		{"bob.ns.cloudflare.com", "Cloudflare"},
		{"ns1-01.azure-dns.com.", "Microsoft"},
		{"myapp.azurewebsites.net", "Microsoft"},
		{"mail.protection.outlook.com.", "Microsoft"},
		{"cname.vercel-dns.com.", "Vercel"},
		{"mysite.netlify.app.", "Netlify"},
		{"user.github.io.", "GitHub"},
		{"dualstack.fastly.net", "Fastly"},
		{"a1234.dscb.akamaiedge.net.", "Akamai"},
		{"ns1.digitalocean.com.", "DigitalOcean"},
		{"robot.your-server.de.", "Hetzner"},
		{"mail.protonmail.ch.", "ProtonMail"},
		{"mx.zoho.com.", "Zoho"},
	}

	for _, tt := range tests {
		t.Run(tt.hostname, func(t *testing.T) {
			p := DetectProvider(tt.hostname)
			if p == nil {
				t.Fatalf("DetectProvider(%q) = nil, want %q", tt.hostname, tt.wantName)
			}
			if p.Name != tt.wantName {
				t.Errorf("DetectProvider(%q).Name = %q, want %q", tt.hostname, p.Name, tt.wantName)
			}
		})
	}
}

func TestDetectProviderUnknown(t *testing.T) {
	unknowns := []string{
		"ns1.example.com.",
		"mail.custom-domain.org",
		"192.168.1.1",
		"random-host.xyz",
	}
	for _, h := range unknowns {
		if p := DetectProvider(h); p != nil {
			t.Errorf("DetectProvider(%q) = %v, want nil", h, p.Name)
		}
	}
}

func TestProviderLabel(t *testing.T) {
	label := ProviderLabel("aspmx.l.google.com.")
	if label == "" {
		t.Fatal("ProviderLabel for google.com host returned empty")
	}
	if label != "🔵 Google" {
		t.Errorf("ProviderLabel = %q, want %q", label, "🔵 Google")
	}

	if label := ProviderLabel("unknown.example.com"); label != "" {
		t.Errorf("ProviderLabel for unknown host = %q, want empty", label)
	}
}
