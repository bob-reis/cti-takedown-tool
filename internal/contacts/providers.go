package contacts

import (
	"strings"

	"github.com/cti-team/takedown/pkg/models"
)

// GetCDNProviders retorna mapa de provedores CDN conhecidos
func GetCDNProviders() map[string]*models.CDNInfo {
	return map[string]*models.CDNInfo{
		"cloudflare": {
			Name:    "Cloudflare",
			Webform: "https://www.cloudflare.com/abuse/form",
			Abuse: models.ContactInfo{
				Email: "abuse@cloudflare.com",
			},
		},
		"fastly": {
			Name: "Fastly",
			Abuse: models.ContactInfo{
				Email: "abuse@fastly.com",
			},
		},
		"akamai": {
			Name: "Akamai",
			Abuse: models.ContactInfo{
				Email: "abuse@akamai.com",
			},
		},
		"amazon": {
			Name: "Amazon CloudFront",
			Abuse: models.ContactInfo{
				Email: "abuse@amazonaws.com",
			},
		},
	}
}

// GetASNAbuseEmailMap retorna mapa de emails de abuse para ASNs conhecidos
func GetASNAbuseEmailMap() map[string]string {
	return map[string]string{
		"google llc":            "network-abuse@google.com",
		"cloudflare, inc.":      "abuse@cloudflare.com",
		"amazon.com, inc.":      "abuse@amazonaws.com",
		"microsoft corporation": "abuse@microsoft.com",
		"digitalocean":          "abuse@digitalocean.com",
		"ovh":                   "abuse@ovh.net",
		"hetzner":               "abuse@hetzner.de",
		"vultr":                 "abuse@vultr.com",
		"linode":                "abuse@linode.com",
	}
}

// GetASNAbuseEmail retorna email de abuse para um ASN específico
func GetASNAbuseEmail(asnName string) string {
	asnName = strings.ToLower(asnName)
	abuseEmails := GetASNAbuseEmailMap()

	for provider, email := range abuseEmails {
		if strings.Contains(asnName, provider) {
			return email
		}
	}

	// Fallback genérico
	return "abuse@" + ExtractDomainFromName(asnName)
}

// ExtractDomainFromName extrai um possível domínio do nome do ASN
func ExtractDomainFromName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, ",", "")
	name = strings.ReplaceAll(name, "llc", "")
	name = strings.ReplaceAll(name, "inc", "")
	name = strings.ReplaceAll(name, "ltd", "")
	name = strings.ReplaceAll(name, "corporation", "")
	name = strings.TrimSpace(name)

	if name == "" {
		return "example.com"
	}

	return name + ".com"
}