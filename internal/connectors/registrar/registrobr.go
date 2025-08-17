package registrar

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cti-team/takedown/internal/state"
	"github.com/cti-team/takedown/pkg/models"
)

// RegistroBRConnector implementa connector para Registro.br
type RegistroBRConnector struct {
	httpClient *http.Client
	userAgent  string
}

// NewRegistroBRConnector cria um novo connector para Registro.br
func NewRegistroBRConnector() *RegistroBRConnector {
	return &RegistroBRConnector{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "CTI-Takedown/1.0",
	}
}

// GetType retorna o tipo do connector
func (r *RegistroBRConnector) GetType() string {
	return "registrar"
}

// Submit submete um takedown request para Registro.br
func (r *RegistroBRConnector) Submit(ctx context.Context, request *models.TakedownRequest, evidence *models.EvidencePack) error {
	// Verificar se é domínio .br
	domain := r.extractDomain(evidence.Defanged)
	if !strings.HasSuffix(strings.ToLower(domain), ".br") {
		return fmt.Errorf("this connector is only for .br domains")
	}

	// Determinar o tipo de solicitação baseado nas tags
	if r.isBrandDispute(request.Tags) {
		return r.submitBrandDispute(ctx, request, evidence)
	} else {
		return r.submitContentAbuse(ctx, request, evidence)
	}
}

// CheckStatus verifica o status no Registro.br
func (r *RegistroBRConnector) CheckStatus(ctx context.Context, request *models.TakedownRequest) (*state.StatusUpdate, error) {
	// Registro.br não tem API pública para verificar status
	// Retornamos um status genérico baseado no tempo

	now := time.Now()
	nextFollowUp := now.Add(72 * time.Hour) // 3 dias para .br

	return &state.StatusUpdate{
		Status:       models.StatusFollowUp,
		Notes:        "Awaiting response from Registro.br (72h SLA for .br domains)",
		NextFollowUp: &nextFollowUp,
	}, nil
}

// isBrandDispute verifica se é uma disputa de marca
func (r *RegistroBRConnector) isBrandDispute(tags []string) bool {
	for _, tag := range tags {
		if strings.Contains(tag, "brand") || strings.Contains(tag, "typosquatting") {
			return true
		}
	}
	return false
}

// submitBrandDispute submete uma disputa de marca para SACI-Adm
func (r *RegistroBRConnector) submitBrandDispute(ctx context.Context, request *models.TakedownRequest, evidence *models.EvidencePack) error {
	// Para disputas de marca em .br, o processo é através do SACI-Adm
	// Este é um processo mais complexo que requer documentação legal

	request.AddEvent("brand_dispute_identified", "system", "",
		"Brand dispute for .br domain - requires SACI-Adm process")

	// Preparar dados para SACI-Adm
	saciData := r.prepareSACIData(request, evidence)

	// Por enquanto, apenas logamos os dados
	// Em produção, isso seria integrado com o sistema SACI-Adm
	request.AddEvent("saci_data_prepared", "system", "",
		fmt.Sprintf("SACI-Adm data prepared: %s", saciData))

	return nil
}

// submitContentAbuse submete abuso de conteúdo
func (r *RegistroBRConnector) submitContentAbuse(ctx context.Context, request *models.TakedownRequest, evidence *models.EvidencePack) error {
	// Para conteúdo malicioso em .br, contactamos o hosting/ISP
	// Registro.br não remove conteúdo, apenas questões contratuais do domínio

	request.AddEvent("content_abuse_identified", "system", "",
		".br content abuse - targeting hosting provider instead of registry")

	// Redirecionar para hosting
	if len(evidence.DNS.A) > 0 {
		ip := evidence.DNS.A[0]
		request.AddEvent("redirect_to_hosting", "system", ip,
			fmt.Sprintf("Redirecting to hosting provider for IP: %s", ip))
	}

	// Opcionalmente, também notificar CERT.br para coordenação
	err := r.notifyCERTBR(ctx, request, evidence)
	if err != nil {
		request.AddEvent("cert_notification_failed", "system", "",
			fmt.Sprintf("Failed to notify CERT.br: %v", err))
	}

	return nil
}

// prepareSACIData prepara dados para submissão ao SACI-Adm
func (r *RegistroBRConnector) prepareSACIData(request *models.TakedownRequest, evidence *models.EvidencePack) string {
	domain := r.extractDomain(evidence.Defanged)

	data := fmt.Sprintf(`SACI-Adm Submission Data:
Domain: %s
Case ID: %s
Category: Brand Dispute / Typosquatting
Evidence ID: %s
Risk Score: %d/100
Analysis: %s
First Seen: %s

Required Documentation:
- Trademark registration certificate
- Evidence of bad faith registration
- Legal standing documentation
- Contact information of rights holder

Process: Manual submission required through SACI-Adm portal
Timeline: 30-60 days for resolution
`, domain, request.CaseID, request.EvidenceID, evidence.Risk.Score,
		evidence.Risk.Rationale, evidence.CollectedAt.Format("2006-01-02"))

	return data
}

// notifyCERTBR notifica o CERT.br para coordenação
func (r *RegistroBRConnector) notifyCERTBR(ctx context.Context, request *models.TakedownRequest, evidence *models.EvidencePack) error {
	// Preparar notificação para CERT.br
	notification := r.prepareCERTNotification(request, evidence)

	// Em produção, isso seria enviado por email para cert@cert.br
	// Por enquanto, apenas registramos o evento
	request.AddEvent("cert_br_notified", "email", "cert@cert.br", notification)

	return nil
}

// prepareCERTNotification prepara notificação para CERT.br
func (r *RegistroBRConnector) prepareCERTNotification(request *models.TakedownRequest, evidence *models.EvidencePack) string {
	domain := r.extractDomain(evidence.Defanged)

	return fmt.Sprintf(`Assunto: [Coordenação de Incidente] %s — %s

Prezados,

Estamos coordenando takedown do domínio %s com indícios de %s.
Solicitamos apoio na coordenação com a rede de origem caso necessário.

DADOS DO CASO:
- Case ID: %s
- Evidence ID: %s
- Domínio: %s
- Risk Score: %d/100
- Análise: %s
- Primeiro visto: %s

AÇÃO TOMADA:
- Notificação ao provedor de hosting
- Processo iniciado conforme procedimentos

Contato técnico: security@example.com

Atenciosamente,
CTI Security Team`,
		evidence.Risk.Category, domain, domain, evidence.Risk.Category,
		request.CaseID, request.EvidenceID, domain, evidence.Risk.Score,
		evidence.Risk.Rationale, evidence.CollectedAt.Format("2006-01-02 15:04:05 UTC"))
}

// extractDomain extrai o domínio de uma URL defanged
func (r *RegistroBRConnector) extractDomain(defanged string) string {
	// Remover defang
	clean := strings.ReplaceAll(defanged, "[.]", ".")
	clean = strings.ReplaceAll(clean, "hxxp://", "http://")
	clean = strings.ReplaceAll(clean, "hxxps://", "https://")

	// Parse URL
	if parsedURL, err := url.Parse(clean); err == nil {
		return parsedURL.Hostname()
	}

	// Fallback: assumir que é um domínio direto
	return clean
}
