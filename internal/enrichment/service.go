package enrichment

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/cti-team/takedown/internal/contacts"
	"github.com/cti-team/takedown/pkg/models"
	"github.com/cti-team/takedown/pkg/rdap"
)

// Service enriquece IOCs com informações adicionais
type Service struct {
	rdapClient *rdap.Client
}

// NewService cria um novo serviço de enrichment
func NewService() *Service {
	return &Service{
		rdapClient: rdap.NewClient(),
	}
}

// EnrichIOC enriquece um IOC com informações de RDAP e ASN
func (s *Service) EnrichIOC(ctx context.Context, evidenceID string) (*models.AbuseContact, error) {
	// TODO: Carregar evidence pack pelo ID
	// Por enquanto, vamos simular com um domínio de exemplo
	domain := "suspicious-domain.com"

	// Buscar informações RDAP
	contact, err := s.rdapClient.LookupDomain(domain)
	if err != nil {
		return nil, fmt.Errorf("RDAP lookup failed: %w", err)
	}

	// Enriquecer com informações de hosting
	if err := s.enrichHosting(ctx, domain, contact); err != nil {
		// Log error but continue
		_, _ = fmt.Fprintf(os.Stderr, "Hosting enrichment failed: %v\n", err)
	}

	// Detectar CDN
	if err := s.detectCDN(ctx, domain, contact); err != nil {
		// Log error but continue
		_, _ = fmt.Fprintf(os.Stderr, "CDN detection failed: %v\n", err)
	}

	return contact, nil
}

// enrichHosting enriquece com informações do provedor de hosting
func (s *Service) enrichHosting(ctx context.Context, domain string, contact *models.AbuseContact) error {
	// Resolver IP do domínio
	ips, err := net.LookupHost(domain)
	if err != nil {
		return fmt.Errorf("IP lookup failed: %w", err)
	}

	if len(ips) == 0 {
		return fmt.Errorf("no IPs found for domain")
	}

	// Usar o primeiro IP para lookup de ASN
	ip := ips[0]

	// Lookup ASN (implementação simplificada)
	asn, asnName, err := s.lookupASN(ip)
	if err != nil {
		return fmt.Errorf("ASN lookup failed: %w", err)
	}

	contact.Hosting = &models.HostingInfo{
		ASN:  asn,
		Name: asnName,
		Abuse: models.ContactInfo{
			Email: contacts.GetASNAbuseEmail(asnName),
		},
	}

	return nil
}

// detectCDN detecta se o domínio usa CDN
func (s *Service) detectCDN(ctx context.Context, domain string, contact *models.AbuseContact) error {
	// Verificar CNAME para detectar CDNs conhecidos
	cname, err := net.LookupCNAME(domain)
	if err != nil {
		// Não é erro crítico
		return nil
	}

	cname = strings.ToLower(cname)

	// Detectar CDNs baseado em CNAME patterns
	cdnProviders := contacts.GetCDNProviders()

	for pattern, cdnInfo := range cdnProviders {
		if strings.Contains(cname, pattern) {
			contact.CDN = cdnInfo
			break
		}
	}

	return nil
}

// lookupASN realiza lookup de ASN para um IP (implementação simplificada)
func (s *Service) lookupASN(ip string) (int, string, error) {
	// Esta é uma implementação muito simplificada
	// Em produção, usaria serviços como Team Cymru ou MaxMind

	// Mapeamento estático para ASNs conhecidos (apenas para demo)
	asnMap := map[string]struct {
		ASN  int
		Name string
	}{
		"8.8.8.8":        {15169, "Google LLC"},
		"1.1.1.1":        {13335, "Cloudflare, Inc."},
		"208.67.222.222": {36692, "Cisco OpenDNS"},
	}

	if info, exists := asnMap[ip]; exists {
		return info.ASN, info.Name, nil
	}

	// Fallback genérico baseado na classe do IP
	if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
		return 0, "Private Network", nil
	}

	// Parse IP para determinar região aproximada
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return 0, "Unknown", fmt.Errorf("invalid IP format")
	}

	firstOctet, _ := strconv.Atoi(parts[0])

	switch {
	case firstOctet >= 1 && firstOctet <= 126:
		return 999999, "Generic Hosting Provider", nil
	case firstOctet >= 128 && firstOctet <= 191:
		return 888888, "International ISP", nil
	default:
		return 777777, "Unknown Provider", nil
	}
}
