package dns

import "strings"

// Provider represents a known DNS/cloud service provider.
type Provider struct {
	Name string
	Icon string
}

var (
	// Well-known providers with terminal-friendly icons.
	ProviderGoogle     = Provider{"Google", "🔵"}
	ProviderAWS        = Provider{"AWS", "🟠"}
	ProviderCloudflare = Provider{"Cloudflare", "🟧"}
	ProviderMicrosoft  = Provider{"Microsoft", "🔷"}
	ProviderVercel     = Provider{"Vercel", "▲"}
	ProviderNetlify    = Provider{"Netlify", "◆"}
	ProviderGitHub     = Provider{"GitHub", "●"}
	ProviderFastly     = Provider{"Fastly", "⚡"}
	ProviderAkamai     = Provider{"Akamai", "☁"}
	ProviderDigitalOcean = Provider{"DigitalOcean", "🔹"}
	ProviderHetzner    = Provider{"Hetzner", "⬡"}
	ProviderProtonMail = Provider{"ProtonMail", "🟣"}
	ProviderZoho       = Provider{"Zoho", "✦"}
)

// hostPatterns maps hostname suffix patterns to providers.
var hostPatterns = []struct {
	suffix   string
	provider Provider
}{
	// Google / GCP
	{".google.com.", ProviderGoogle},
	{".google.com", ProviderGoogle},
	{".googleapis.com.", ProviderGoogle},
	{".googleapis.com", ProviderGoogle},
	{".googlehosted.com.", ProviderGoogle},
	{".googlehosted.com", ProviderGoogle},
	{".googledomains.com.", ProviderGoogle},
	{".googledomains.com", ProviderGoogle},
	{".googleplex.com.", ProviderGoogle},
	{".googleplex.com", ProviderGoogle},
	{".1e100.net.", ProviderGoogle},
	{".1e100.net", ProviderGoogle},

	// AWS
	{".amazonaws.com.", ProviderAWS},
	{".amazonaws.com", ProviderAWS},
	{".cloudfront.net.", ProviderAWS},
	{".cloudfront.net", ProviderAWS},
	{".awsdns-", ProviderAWS},
	{".elasticbeanstalk.com.", ProviderAWS},
	{".elasticbeanstalk.com", ProviderAWS},

	// Cloudflare
	{".cloudflare.com.", ProviderCloudflare},
	{".cloudflare.com", ProviderCloudflare},
	{".cloudflaressl.com.", ProviderCloudflare},
	{".cloudflaressl.com", ProviderCloudflare},
	{".ns.cloudflare.com.", ProviderCloudflare},
	{".ns.cloudflare.com", ProviderCloudflare},

	// Microsoft / Azure
	{".azure-dns.com.", ProviderMicrosoft},
	{".azure-dns.com", ProviderMicrosoft},
	{".azure-dns.net.", ProviderMicrosoft},
	{".azure-dns.net", ProviderMicrosoft},
	{".azure-dns.org.", ProviderMicrosoft},
	{".azure-dns.org", ProviderMicrosoft},
	{".azure-dns.info.", ProviderMicrosoft},
	{".azure-dns.info", ProviderMicrosoft},
	{".azurewebsites.net.", ProviderMicrosoft},
	{".azurewebsites.net", ProviderMicrosoft},
	{".trafficmanager.net.", ProviderMicrosoft},
	{".trafficmanager.net", ProviderMicrosoft},
	{".mail.protection.outlook.com.", ProviderMicrosoft},
	{".mail.protection.outlook.com", ProviderMicrosoft},
	{".outlook.com.", ProviderMicrosoft},
	{".outlook.com", ProviderMicrosoft},

	// Vercel
	{".vercel-dns.com.", ProviderVercel},
	{".vercel-dns.com", ProviderVercel},
	{".vercel.app.", ProviderVercel},
	{".vercel.app", ProviderVercel},

	// Netlify
	{".netlify.app.", ProviderNetlify},
	{".netlify.app", ProviderNetlify},
	{".netlify.com.", ProviderNetlify},
	{".netlify.com", ProviderNetlify},

	// GitHub
	{".github.io.", ProviderGitHub},
	{".github.io", ProviderGitHub},
	{".github.com.", ProviderGitHub},
	{".github.com", ProviderGitHub},

	// Fastly
	{".fastly.net.", ProviderFastly},
	{".fastly.net", ProviderFastly},

	// Akamai
	{".akamaiedge.net.", ProviderAkamai},
	{".akamaiedge.net", ProviderAkamai},
	{".akamai.net.", ProviderAkamai},
	{".akamai.net", ProviderAkamai},
	{".akadns.net.", ProviderAkamai},
	{".akadns.net", ProviderAkamai},

	// DigitalOcean
	{".digitalocean.com.", ProviderDigitalOcean},
	{".digitalocean.com", ProviderDigitalOcean},

	// Hetzner
	{".hetzner.com.", ProviderHetzner},
	{".hetzner.com", ProviderHetzner},
	{".your-server.de.", ProviderHetzner},
	{".your-server.de", ProviderHetzner},

	// ProtonMail
	{".protonmail.ch.", ProviderProtonMail},
	{".protonmail.ch", ProviderProtonMail},

	// Zoho
	{".zoho.com.", ProviderZoho},
	{".zoho.com", ProviderZoho},
	{".zoho.eu.", ProviderZoho},
	{".zoho.eu", ProviderZoho},
}

// DetectProvider returns the provider for a given hostname, or nil if unknown.
func DetectProvider(hostname string) *Provider {
	h := strings.ToLower(hostname)
	for _, p := range hostPatterns {
		if strings.HasSuffix(h, p.suffix) || strings.Contains(h, p.suffix) {
			return &p.provider
		}
	}
	return nil
}

// ProviderLabel returns a formatted "icon Name" string for display, or empty if unknown.
func ProviderLabel(hostname string) string {
	p := DetectProvider(hostname)
	if p == nil {
		return ""
	}
	return p.Icon + " " + p.Name
}
