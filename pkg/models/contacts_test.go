package models

import (
	"testing"
)

func TestAbuseContact_GetPrimaryAbuseEmail(t *testing.T) {
	tests := []struct {
		name     string
		contact  *AbuseContact
		expected string
	}{
		{
			name: "Direct abuse email available",
			contact: &AbuseContact{
				Abuse: ContactInfo{
					Email: "abuse@example.com",
				},
			},
			expected: "abuse@example.com",
		},
		{
			name: "No direct abuse email, use registrar fallback",
			contact: &AbuseContact{
				Abuse: ContactInfo{},
				Registrar: &RegistrarInfo{
					Name: "GoDaddy.com, LLC",
				},
			},
			expected: "abuse@godaddy.com",
		},
		{
			name: "Unknown registrar fallback",
			contact: &AbuseContact{
				Abuse: ContactInfo{},
				Registrar: &RegistrarInfo{
					Name: "Unknown Registrar Inc",
				},
			},
			expected: "",
		},
		{
			name: "No abuse email and no registrar",
			contact: &AbuseContact{
				Abuse: ContactInfo{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.contact.GetPrimaryAbuseEmail()
			if result != tt.expected {
				t.Errorf("GetPrimaryAbuseEmail() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAbuseContact_GetTargets(t *testing.T) {
	contact := &AbuseContact{
		Domain: "malicious.com",
		Registrar: &RegistrarInfo{
			Name:   "GoDaddy.com, LLC",
			IANAID: 146,
		},
		Abuse: ContactInfo{
			Email: "abuse@godaddy.com",
		},
		Hosting: &HostingInfo{
			ASN:  12345,
			Name: "Example Hosting",
			Abuse: ContactInfo{
				Email: "abuse@examplehosting.com",
			},
		},
		CDN: &CDNInfo{
			Name: "Cloudflare",
			Abuse: ContactInfo{
				Email: "abuse@cloudflare.com",
			},
			Webform: "https://www.cloudflare.com/abuse/form",
		},
	}

	tests := []struct {
		name          string
		category      string
		expectedCount int
		expectedTypes []string
	}{
		{
			name:          "Phishing - all targets",
			category:      "phishing",
			expectedCount: 3,
			expectedTypes: []string{"registrar", "hosting", "cdn"},
		},
		{
			name:          "Malware - hosting and registrar",
			category:      "malware",
			expectedCount: 3,
			expectedTypes: []string{"registrar", "hosting", "cdn"},
		},
		{
			name:          "C2 - hosting and registrar",
			category:      "c2",
			expectedCount: 3,
			expectedTypes: []string{"registrar", "hosting", "cdn"},
		},
		{
			name:          "Brand dispute - registrar and CDN",
			category:      "brand",
			expectedCount: 2,
			expectedTypes: []string{"registrar", "cdn"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targets := contact.GetTargets(tt.category)

			if len(targets) != tt.expectedCount {
				t.Errorf("Expected %d targets, got %d", tt.expectedCount, len(targets))
			}

			// Check that expected types are present
			foundTypes := make(map[string]bool)
			for _, target := range targets {
				foundTypes[target.Type] = true
			}

			for _, expectedType := range tt.expectedTypes {
				if !foundTypes[expectedType] {
					t.Errorf("Expected target type %s not found", expectedType)
				}
			}
		})
	}
}

func TestRegistrarInfo_Structure(t *testing.T) {
	registrar := RegistrarInfo{
		Name:   "GoDaddy.com, LLC",
		IANAID: 146,
	}

	if registrar.Name != "GoDaddy.com, LLC" {
		t.Errorf("Name should be 'GoDaddy.com, LLC', got %s", registrar.Name)
	}
	if registrar.IANAID != 146 {
		t.Errorf("IANAID should be 146, got %d", registrar.IANAID)
	}
}

func TestContactInfo_Structure(t *testing.T) {
	contact := ContactInfo{
		Email:   "abuse@example.com",
		Phone:   "+1-555-0123",
		Webform: "https://example.com/abuse",
	}

	if contact.Email != "abuse@example.com" {
		t.Errorf("Email not correct")
	}
	if contact.Phone != "+1-555-0123" {
		t.Errorf("Phone not correct")
	}
	if contact.Webform != "https://example.com/abuse" {
		t.Errorf("Webform not correct")
	}
}

func TestHostingInfo_Structure(t *testing.T) {
	hosting := HostingInfo{
		ASN:  64512,
		Name: "Example Cloud Services",
		Abuse: ContactInfo{
			Email: "abuse@examplecloud.com",
		},
	}

	if hosting.ASN != 64512 {
		t.Errorf("ASN should be 64512, got %d", hosting.ASN)
	}
	if hosting.Name != "Example Cloud Services" {
		t.Errorf("Name not correct")
	}
	if hosting.Abuse.Email != "abuse@examplecloud.com" {
		t.Errorf("Abuse email not correct")
	}
}

func TestCDNInfo_Structure(t *testing.T) {
	cdn := CDNInfo{
		Name: "Cloudflare",
		Abuse: ContactInfo{
			Email: "abuse@cloudflare.com",
		},
		Webform: "https://www.cloudflare.com/abuse/form",
	}

	if cdn.Name != "Cloudflare" {
		t.Errorf("Name should be 'Cloudflare', got %s", cdn.Name)
	}
	if cdn.Abuse.Email != "abuse@cloudflare.com" {
		t.Errorf("Abuse email not correct")
	}
	if cdn.Webform != "https://www.cloudflare.com/abuse/form" {
		t.Errorf("Webform not correct")
	}
}

func TestGetRegistrarAbuseEmail(t *testing.T) {
	tests := []struct {
		registrarName string
		expected      string
	}{
		{"GoDaddy.com, LLC", "abuse@godaddy.com"},
		{"NameCheap, Inc.", "abuse@namecheap.com"},
		{"Registro.br", "abuse@registro.br"},
		{"Amazon Registrar, Inc.", "legal@amazon.com"},
		{"Google LLC", "domain-abuse@google.com"},
		{"Cloudflare, Inc.", "abuse@cloudflare.com"},
		{"Unknown Registrar", ""},
	}

	for _, tt := range tests {
		t.Run(tt.registrarName, func(t *testing.T) {
			result := getRegistrarAbuseEmail(tt.registrarName)
			if result != tt.expected {
				t.Errorf("getRegistrarAbuseEmail(%s) = %v, want %v",
					tt.registrarName, result, tt.expected)
			}
		})
	}
}

func TestAbuseContact_CompleteStructure(t *testing.T) {
	contact := &AbuseContact{
		Domain: "suspicious.com",
		Registrar: &RegistrarInfo{
			Name:   "Example Registrar",
			IANAID: 999,
		},
		Abuse: ContactInfo{
			Email:   "abuse@example.com",
			Phone:   "+1-555-ABUSE",
			Webform: "https://example.com/report",
		},
		Hosting: &HostingInfo{
			ASN:  12345,
			Name: "Cloud Provider Inc",
			Abuse: ContactInfo{
				Email: "security@cloudprovider.com",
			},
		},
		CDN: &CDNInfo{
			Name: "FastCDN",
			Abuse: ContactInfo{
				Email: "abuse@fastcdn.com",
			},
		},
		Privacy: true,
	}

	// Test all fields are accessible
	if contact.Domain != "suspicious.com" {
		t.Errorf("Domain not correct")
	}
	if contact.Registrar.Name != "Example Registrar" {
		t.Errorf("Registrar name not correct")
	}
	if contact.Hosting.ASN != 12345 {
		t.Errorf("Hosting ASN not correct")
	}
	if contact.CDN.Name != "FastCDN" {
		t.Errorf("CDN name not correct")
	}
	if !contact.Privacy {
		t.Errorf("Privacy flag should be true")
	}

	// Test primary abuse email
	primaryEmail := contact.GetPrimaryAbuseEmail()
	if primaryEmail != "abuse@example.com" {
		t.Errorf("Primary abuse email not correct")
	}

	// Test target generation
	targets := contact.GetTargets("phishing")
	if len(targets) == 0 {
		t.Errorf("Should have targets for phishing")
	}
}
