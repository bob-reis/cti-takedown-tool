package hosting

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cti-team/takedown/internal/connectors/registrar"
	"github.com/cti-team/takedown/internal/state"
	"github.com/cti-team/takedown/pkg/models"
)

// GenericHostingConnector implementa connector genérico para provedores de hosting.
type GenericHostingConnector struct {
	smtpConfig registrar.SMTPConfig
	templates  map[string]string
}

// NewGenericHostingConnector cria um novo connector genérico para hosting.
func NewGenericHostingConnector(smtpConfig registrar.SMTPConfig) *GenericHostingConnector {
	connector := &GenericHostingConnector{
		smtpConfig: smtpConfig,
		templates:  make(map[string]string),
	}
	connector.loadTemplates()

	return connector
}

// GetType retorna o tipo do connector.
func (g *GenericHostingConnector) GetType() string {
	return "hosting"
}

// Submit submete um takedown request para o provedor de hosting.
func (g *GenericHostingConnector) Submit(_ context.Context, request *models.TakedownRequest, evidence *models.EvidencePack) error {
	// Preparar email baseado no template
	subject, body, err := g.prepareEmail(request, evidence)
	if err != nil {
		return fmt.Errorf("failed to prepare email: %w", err)
	}

	// Determinar email de destino
	abuseEmail := request.Target.Email
	if abuseEmail == "" {
		abuseEmail = g.guessAbuseEmail(request.Target.Entity)
	}

	// Enviar email
	err = g.sendEmail(abuseEmail, subject, body)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Registrar submissão
	request.AddEvent("email_sent", "email", abuseEmail,
		fmt.Sprintf("Sent content removal request to %s", abuseEmail))

	return nil
}

// CheckStatus verifica o status junto ao provedor de hosting.
func (g *GenericHostingConnector) CheckStatus(_ context.Context, request *models.TakedownRequest) (*state.StatusUpdate, error) {
	// A maioria dos provedores não tem API pública para status
	// Retornamos status baseado no SLA do hosting

	now := time.Now()
	nextFollowUp := now.Add(24 * time.Hour) // Hosting geralmente responde em 24h

	return &state.StatusUpdate{
		Status:       models.StatusFollowUp,
		Notes:        "Awaiting response from hosting provider (24-48h SLA)",
		NextFollowUp: &nextFollowUp,
	}, nil
}

// prepareEmail prepara o email para o provedor de hosting.
func (g *GenericHostingConnector) prepareEmail(request *models.TakedownRequest, evidence *models.EvidencePack) (string, string, error) {
	// Determinar categoria
	category := "malicious_content"
	if request.Tags != nil {
		for _, tag := range request.Tags {
			if tag == "phishing" {
				category = "phishing"
				break
			} else if tag == "malware" {
				category = "malware"
				break
			} else if tag == "c2" {
				category = "c2"
				break
			}
		}
	}

	template, exists := g.templates[category]
	if !exists {
		template = g.templates["default"]
	}

	// Extrair informações do evidence
	domain := g.extractDomain(evidence.Defanged)
	ip := "unknown"
	if len(evidence.DNS.A) > 0 {
		ip = evidence.DNS.A[0]
	}

	// Preparar subject
	title := strings.ToUpper(category[:1]) + category[1:]
	subject := fmt.Sprintf("[Abuse] %s content hosted on your network — %s",
		title, ip)

	// Substituir placeholders
	body := template
	body = strings.ReplaceAll(body, "{case_id}", request.CaseID)
	body = strings.ReplaceAll(body, "{evidence_id}", request.EvidenceID)
	body = strings.ReplaceAll(body, "{domain}", domain)
	body = strings.ReplaceAll(body, "{ip}", ip)
	body = strings.ReplaceAll(body, "{category}", category)
	body = strings.ReplaceAll(body, "{provider}", request.Target.Entity)
	body = strings.ReplaceAll(body, "{first_seen}", evidence.CollectedAt.Format("2006-01-02 15:04:05 UTC"))
	body = strings.ReplaceAll(body, "{risk_score}", fmt.Sprintf("%d", evidence.Risk.Score))
	body = strings.ReplaceAll(body, "{rationale}", evidence.Risk.Rationale)
	body = strings.ReplaceAll(body, "{defanged_url}", evidence.Defanged)

	return subject, body, nil
}

// sendEmail envia email usando a mesma função do registrar connector
func (g *GenericHostingConnector) sendEmail(to, subject, body string) error {
	// Reutilizar função de envio de email
	// Em uma implementação real, isso seria um serviço compartilhado
	return fmt.Errorf("SMTP not implemented in this demo - would send to %s", to)
}

// guessAbuseEmail tenta adivinhar o email de abuse do provedor
func (g *GenericHostingConnector) guessAbuseEmail(providerName string) string {
	// Mapeamento de provedores conhecidos
	knownProviders := map[string]string{
		"digitalocean": "abuse@digitalocean.com",
		"amazon":       "abuse@amazonaws.com",
		"google":       "network-abuse@google.com",
		"microsoft":    "abuse@microsoft.com",
		"cloudflare":   "abuse@cloudflare.com",
		"ovh":          "abuse@ovh.net",
		"hetzner":      "abuse@hetzner.de",
		"vultr":        "abuse@vultr.com",
		"linode":       "abuse@linode.com",
	}

	providerLower := strings.ToLower(providerName)
	for provider, email := range knownProviders {
		if strings.Contains(providerLower, provider) {
			return email
		}
	}

	// Fallback: abuse@domain
	domain := g.extractDomainFromProvider(providerName)
	return fmt.Sprintf("abuse@%s", domain)
}

// extractDomainFromProvider extrai domínio do nome do provedor
func (g *GenericHostingConnector) extractDomainFromProvider(provider string) string {
	provider = strings.ToLower(provider)
	provider = strings.ReplaceAll(provider, " ", "")
	provider = strings.ReplaceAll(provider, ",", "")
	provider = strings.ReplaceAll(provider, "llc", "")
	provider = strings.ReplaceAll(provider, "inc", "")
	provider = strings.ReplaceAll(provider, "ltd", "")

	if provider == "" {
		return "example.com"
	}

	return provider + ".com"
}

// extractDomain extrai domínio de URL defanged
func (g *GenericHostingConnector) extractDomain(defanged string) string {
	clean := strings.ReplaceAll(defanged, "[.]", ".")
	clean = strings.ReplaceAll(clean, "hxxp://", "http://")
	clean = strings.ReplaceAll(clean, "hxxps://", "https://")

	// Extrair apenas o hostname
	if strings.Contains(clean, "://") {
		parts := strings.Split(clean, "://")
		if len(parts) > 1 {
			hostPath := parts[1]
			if strings.Contains(hostPath, "/") {
				return strings.Split(hostPath, "/")[0]
			}
			return hostPath
		}
	}

	return clean
}

// loadTemplates carrega templates de email para hosting
func (g *GenericHostingConnector) loadTemplates() {
	g.templates["phishing"] = `Hello Abuse Team,

We detected phishing content being served from your network.
Please REMOVE the content and notify the customer.

CASE DETAILS:
- Case ID: {case_id}
- Evidence ID: {evidence_id}
- Domain: {domain}
- IP Address: {ip}
- Provider: {provider}
- Category: Phishing
- Risk Score: {risk_score}/100
- First seen: {first_seen}

EVIDENCE:
- URLs (defanged): {defanged_url}
- Analysis: {rationale}

REQUESTED ACTION:
Immediate removal of phishing content and customer notification.

IMPACT:
The content is actively harvesting user credentials, causing financial 
damage and personal data compromise.

Please provide a ticket ID and status update.

Regards,
CTI Security Team
security@example.com`

	g.templates["malware"] = `Hello Abuse Team,

We detected malware distribution from your network infrastructure.
Please REMOVE the malicious content immediately.

CASE DETAILS:
- Case ID: {case_id}
- Evidence ID: {evidence_id}
- Domain: {domain}
- IP Address: {ip}
- Provider: {provider}
- Category: Malware Distribution
- Risk Score: {risk_score}/100
- First seen: {first_seen}

EVIDENCE:
- URLs (defanged): {defanged_url}
- Analysis: {rationale}

REQUESTED ACTION:
Immediate removal of malware payload and customer notification.

Please confirm receipt and provide ticket ID.

Regards,
CTI Security Team`

	g.templates["c2"] = `Hello Abuse Team,

We identified Command & Control (C2) infrastructure on your network.
Please TAKE DOWN the malicious infrastructure immediately.

CASE DETAILS:
- Case ID: {case_id}
- Evidence ID: {evidence_id}
- Domain: {domain}
- IP Address: {ip}
- Provider: {provider}
- Category: C2 Infrastructure
- Risk Score: {risk_score}/100
- First seen: {first_seen}

EVIDENCE:
- URLs (defanged): {defanged_url}
- Analysis: {rationale}

REQUESTED ACTION:
Immediate takedown of C2 infrastructure.

URGENCY: HIGH - Active malware campaigns depend on this infrastructure.

Please confirm immediate action and provide ticket reference.

Regards,
CTI Security Team`

	g.templates["default"] = `Hello Abuse Team,

We detected malicious content hosted on your network infrastructure.
Please investigate and take appropriate action.

CASE DETAILS:
- Case ID: {case_id}
- Evidence ID: {evidence_id}
- Domain: {domain}
- IP Address: {ip}
- Provider: {provider}
- Category: {category}
- Risk Score: {risk_score}/100
- First seen: {first_seen}

EVIDENCE:
- URLs (defanged): {defanged_url}
- Analysis: {rationale}

Please investigate and take appropriate action per your AUP.

Regards,
CTI Security Team`
}
