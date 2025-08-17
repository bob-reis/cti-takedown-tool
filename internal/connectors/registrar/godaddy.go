package registrar

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/cti-team/takedown/internal/state"
	"github.com/cti-team/takedown/pkg/models"
)

// GoDaddyConnector implementa connector para GoDaddy
type GoDaddyConnector struct {
	smtpConfig SMTPConfig
	templates  map[string]string
}

// SMTPConfig configuração para envio de emails
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// NewGoDaddyConnector cria um novo connector para GoDaddy
func NewGoDaddyConnector(smtpConfig SMTPConfig) *GoDaddyConnector {
	connector := &GoDaddyConnector{
		smtpConfig: smtpConfig,
		templates:  make(map[string]string),
	}
	connector.loadTemplates()
	return connector
}

// GetType retorna o tipo do connector
func (g *GoDaddyConnector) GetType() string {
	return "registrar"
}

// Submit submete um takedown request para GoDaddy
func (g *GoDaddyConnector) Submit(ctx context.Context, request *models.TakedownRequest, evidence *models.EvidencePack) error {
	// Verificar se é realmente GoDaddy
	if !strings.Contains(strings.ToLower(request.Target.Entity), "godaddy") {
		return fmt.Errorf("this connector is only for GoDaddy requests")
	}

	// Preparar email baseado no template
	subject, body, err := g.prepareEmail(request, evidence)
	if err != nil {
		return fmt.Errorf("failed to prepare email: %w", err)
	}

	// Enviar email para abuse@godaddy.com
	abuseEmail := "abuse@godaddy.com"
	if request.Target.Email != "" {
		abuseEmail = request.Target.Email
	}

	err = g.sendEmail(abuseEmail, subject, body)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Registrar submissão
	request.AddEvent("email_sent", "email", abuseEmail, fmt.Sprintf("Sent takedown request to %s", abuseEmail))

	return nil
}

// CheckStatus verifica o status de um request no GoDaddy
func (g *GoDaddyConnector) CheckStatus(ctx context.Context, request *models.TakedownRequest) (*state.StatusUpdate, error) {
	// GoDaddy normalmente responde por email, então não temos API para verificar status
	// Este método seria implementado se houvesse uma API específica

	// Por enquanto, apenas sugerimos um follow-up baseado no tempo
	now := time.Now()
	nextFollowUp := now.Add(48 * time.Hour) // GoDaddy SLA é 48h

	return &state.StatusUpdate{
		Status:       models.StatusFollowUp,
		Notes:        "Awaiting response from GoDaddy (48h SLA)",
		NextFollowUp: &nextFollowUp,
	}, nil
}

// prepareEmail prepara o email de takedown baseado no template
func (g *GoDaddyConnector) prepareEmail(request *models.TakedownRequest, evidence *models.EvidencePack) (string, string, error) {
	// Determinar categoria
	category := "abuse"
	if request.Tags != nil {
		for _, tag := range request.Tags {
			if tag == "phishing" {
				category = "phishing"
				break
			} else if tag == "malware" {
				category = "malware"
				break
			}
		}
	}

	template, exists := g.templates[category]
	if !exists {
		template = g.templates["default"]
	}

	// Substituir placeholders
	domain := evidence.Defanged
	if domain == "" {
		domain = "suspicious-domain[.]com" // fallback
	}

	subject := fmt.Sprintf("[Urgente] Solicitação de suspensão de domínio por %s — %s", category, domain)

	body := template
	body = strings.ReplaceAll(body, "{domain}", domain)
	body = strings.ReplaceAll(body, "{category}", category)
	body = strings.ReplaceAll(body, "{case_id}", request.CaseID)
	body = strings.ReplaceAll(body, "{evidence_id}", request.EvidenceID)
	body = strings.ReplaceAll(body, "{first_seen}", evidence.CollectedAt.Format("2006-01-02 15:04:05 UTC"))
	body = strings.ReplaceAll(body, "{risk_score}", fmt.Sprintf("%d", evidence.Risk.Score))
	body = strings.ReplaceAll(body, "{rationale}", evidence.Risk.Rationale)

	return subject, body, nil
}

// sendEmail envia um email via SMTP
func (g *GoDaddyConnector) sendEmail(to, subject, body string) error {
	// Configurar conexão SMTP
	auth := smtp.PlainAuth("", g.smtpConfig.Username, g.smtpConfig.Password, g.smtpConfig.Host)

	// Preparar headers e corpo do email
	msg := fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", to, g.smtpConfig.From, subject, body)

	// Enviar email
	addr := fmt.Sprintf("%s:%d", g.smtpConfig.Host, g.smtpConfig.Port)
	return smtp.SendMail(addr, auth, g.smtpConfig.From, []string{to}, []byte(msg))
}

// loadTemplates carrega templates de email
func (g *GoDaddyConnector) loadTemplates() {
	g.templates["phishing"] = `Prezados,

Identificamos atividade de phishing no domínio {domain}, registrado através de GoDaddy.
Solicitamos SUSPENSÃO IMEDIATA conforme violação de Terms of Service e DNS Abuse Policy.

DADOS DO CASO:
- Case ID: {case_id}
- Evidence ID: {evidence_id}
- Domínio (defanged): {domain}
- Categoria: {category}
- Risk Score: {risk_score}/100
- Primeiro visto: {first_seen}
- Análise: {rationale}

AÇÃO SOLICITADA:
Suspensão imediata do domínio por violação de ToS (phishing/fraud).

IMPACTO:
O domínio está sendo usado para coletar credenciais de usuários legítimos, 
causando danos financeiros e comprometimento de dados pessoais.

Solicitamos confirmação de recebimento e número de caso para acompanhamento.

Contato técnico para esclarecimentos: security@example.com

Atenciosamente,
CTI Security Team`

	g.templates["malware"] = `Prezados,

Detectamos distribuição de malware através do domínio {domain}, registrado via GoDaddy.
Solicitamos SUSPENSÃO conforme violação de Terms of Service.

DADOS DO CASO:
- Case ID: {case_id}
- Evidence ID: {evidence_id}
- Domínio (defanged): {domain}
- Categoria: {category}
- Risk Score: {risk_score}/100
- Primeiro visto: {first_seen}
- Análise: {rationale}

AÇÃO SOLICITADA:
Suspensão do domínio por distribuição de malware.

Favor confirmar recebimento e fornecer número de caso.

Atenciosamente,
CTI Security Team`

	g.templates["default"] = `Prezados,

Identificamos atividade maliciosa no domínio {domain}, registrado através de GoDaddy.
Solicitamos análise e ação apropriada conforme Terms of Service.

DADOS DO CASO:
- Case ID: {case_id}
- Evidence ID: {evidence_id}
- Domínio (defanged): {domain}
- Categoria: {category}
- Risk Score: {risk_score}/100
- Primeiro visto: {first_seen}
- Análise: {rationale}

Favor analisar e tomar ação apropriada.

Atenciosamente,
CTI Security Team`
}
