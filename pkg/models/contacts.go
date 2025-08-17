package models

// RegistrarInfo representa informações do registrar
type RegistrarInfo struct {
	Name   string `json:"name"`
	IANAID int    `json:"iana_id"`
}

// ContactInfo representa informações de contato para abuse
type ContactInfo struct {
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Webform string `json:"webform,omitempty"`
}

// HostingInfo representa informações do provedor de hosting
type HostingInfo struct {
	ASN   int         `json:"asn"`
	Name  string      `json:"name"`
	Abuse ContactInfo `json:"abuse"`
}

// CDNInfo representa informações de CDN
type CDNInfo struct {
	Name    string      `json:"name"`
	Abuse   ContactInfo `json:"abuse"`
	Webform string      `json:"webform,omitempty"`
}

// AbuseContact representa contatos normalizados por RDAP conforme spec 8.3
type AbuseContact struct {
	Domain    string         `json:"domain"`
	Registrar *RegistrarInfo `json:"registrar,omitempty"`
	Abuse     ContactInfo    `json:"abuse"`
	Hosting   *HostingInfo   `json:"hosting,omitempty"`
	CDN       *CDNInfo       `json:"cdn,omitempty"`
	Privacy   bool           `json:"privacy"` // indica se usa privacy/proxy service
}

// GetPrimaryAbuseEmail retorna o email principal para contato
func (ac *AbuseContact) GetPrimaryAbuseEmail() string {
	if ac.Abuse.Email != "" {
		return ac.Abuse.Email
	}
	if ac.Registrar != nil && ac.Registrar.Name != "" {
		// Fallback para emails padrão de registrars conhecidos
		return getRegistrarAbuseEmail(ac.Registrar.Name)
	}
	return ""
}

// GetTargets retorna lista de targets para takedown baseado no tipo
func (ac *AbuseContact) GetTargets(category string) []TakedownTarget {
	var targets []TakedownTarget

	// Sempre incluir registrar para DNS abuse
	if ac.Registrar != nil {
		targets = append(targets, TakedownTarget{
			Type:   "registrar",
			Entity: ac.Registrar.Name,
			Email:  ac.GetPrimaryAbuseEmail(),
		})
	}

	// Incluir hosting para conteúdo malicioso
	if ac.Hosting != nil && (category == "phishing" || category == "malware" || category == "c2") {
		targets = append(targets, TakedownTarget{
			Type:   "hosting",
			Entity: ac.Hosting.Name,
			Email:  ac.Hosting.Abuse.Email,
		})
	}

	// Incluir CDN se presente
	if ac.CDN != nil {
		targets = append(targets, TakedownTarget{
			Type:    "cdn",
			Entity:  ac.CDN.Name,
			Email:   ac.CDN.Abuse.Email,
			Webform: ac.CDN.Webform,
		})
	}

	return targets
}

func getRegistrarAbuseEmail(registrarName string) string {
	// Mapeamento de registrars conhecidos para emails de abuse
	knownRegistrars := map[string]string{
		"GoDaddy.com, LLC":       "abuse@godaddy.com",
		"NameCheap, Inc.":        "abuse@namecheap.com",
		"Registro.br":            "abuse@registro.br",
		"Amazon Registrar, Inc.": "legal@amazon.com",
		"Google LLC":             "domain-abuse@google.com",
		"Cloudflare, Inc.":       "abuse@cloudflare.com",
		"Network Solutions, LLC": "abuse@networksolutions.com",
		"eNom, LLC":              "abuse@enom.com",
	}

	if email, exists := knownRegistrars[registrarName]; exists {
		return email
	}

	return ""
}
