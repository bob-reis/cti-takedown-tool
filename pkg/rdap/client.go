package rdap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cti-team/takedown/pkg/models"
)

// Client representa um cliente RDAP
type Client struct {
	httpClient *http.Client
	userAgent  string
}

// NewClient cria um novo cliente RDAP
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		userAgent: "CTI-Takedown/1.0",
	}
}

// RDAPResponse representa uma resposta RDAP simplificada
type RDAPResponse struct {
	ObjectClassName string   `json:"objectClassName"`
	Handle          string   `json:"handle"`
	LDHName         string   `json:"ldhName"`
	Entities        []Entity `json:"entities"`
	Events          []Event  `json:"events"`
	Status          []string `json:"status"`
	Nameservers     []string `json:"nameservers"`
}

type Entity struct {
	ObjectClassName string        `json:"objectClassName"`
	Handle          string        `json:"handle"`
	Roles           []string      `json:"roles"`
	VCardArray      []interface{} `json:"vcardArray"`
	Entities        []Entity      `json:"entities"`
}

type Event struct {
	EventAction string    `json:"eventAction"`
	EventDate   time.Time `json:"eventDate"`
}

// LookupDomain realiza lookup RDAP para um domínio
func (c *Client) LookupDomain(domain string) (*models.AbuseContact, error) {
	domain = strings.ToLower(strings.TrimSpace(domain))

	// Determinar servidor RDAP baseado no TLD
	rdapURL, err := c.getRDAPURL(domain)
	if err != nil {
		return nil, fmt.Errorf("failed to determine RDAP URL: %w", err)
	}

	// Fazer requisição RDAP
	resp, err := c.makeRequest(rdapURL + "/domain/" + domain)
	if err != nil {
		return nil, fmt.Errorf("RDAP request failed: %w", err)
	}

	// Parsear resposta
	var rdapResp RDAPResponse
	if err := json.Unmarshal(resp, &rdapResp); err != nil {
		return nil, fmt.Errorf("failed to parse RDAP response: %w", err)
	}

	// Extrair informações de contato
	contact := &models.AbuseContact{
		Domain: domain,
	}

	// Processar entidades para encontrar registrar e contatos
	for _, entity := range rdapResp.Entities {
		c.processEntity(entity, contact)
	}

	return contact, nil
}

// makeRequest faz uma requisição HTTP
func (c *Client) makeRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/rdap+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RDAP server returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// processEntity processa uma entidade RDAP para extrair informações relevantes
func (c *Client) processEntity(entity Entity, contact *models.AbuseContact) {
	// Verificar se é registrar
	if c.hasRole(entity.Roles, "registrar") {
		contact.Registrar = &models.RegistrarInfo{
			Name: c.extractEntityName(entity),
		}

		// Extrair contato de abuse do registrar
		abuseEmail := c.extractAbuseEmail(entity)
		if abuseEmail != "" {
			contact.Abuse.Email = abuseEmail
		}
	}

	// Verificar se é serviço de privacy/proxy
	if c.hasRole(entity.Roles, "proxy") || c.hasRole(entity.Roles, "privacy") {
		contact.Privacy = true
	}

	// Processar entidades aninhadas
	for _, subEntity := range entity.Entities {
		c.processEntity(subEntity, contact)
	}
}

// hasRole verifica se uma entidade tem um papel específico
func (c *Client) hasRole(roles []string, role string) bool {
	for _, r := range roles {
		if strings.EqualFold(r, role) {
			return true
		}
	}
	return false
}

// extractEntityName extrai o nome da entidade do vCard
func (c *Client) extractEntityName(entity Entity) string {
	if len(entity.VCardArray) < 2 {
		return entity.Handle
	}

	vcard := entity.VCardArray[1]
	vcardArray, ok := vcard.([]interface{})
	if !ok {
		return entity.Handle
	}

	for _, item := range vcardArray {
		if itemArray, ok := item.([]interface{}); ok && len(itemArray) >= 4 {
			if prop, ok := itemArray[0].(string); ok && prop == "fn" {
				if name, ok := itemArray[3].(string); ok {
					return name
				}
			}
		}
	}

	return entity.Handle
}

// extractAbuseEmail extrai email de abuse do vCard
func (c *Client) extractAbuseEmail(entity Entity) string {
	vcardArray := c.getVCardArray(entity)
	if vcardArray == nil {
		return ""
	}

	for _, item := range vcardArray {
		if email := c.extractEmailFromItem(item); email != "" {
			return email
		}
	}

	return ""
}

// getVCardArray extrai e valida o array vCard da entidade
func (c *Client) getVCardArray(entity Entity) []interface{} {
	if len(entity.VCardArray) < 2 {
		return nil
	}

	vcard := entity.VCardArray[1]
	vcardArray, ok := vcard.([]interface{})
	if !ok {
		return nil
	}

	return vcardArray
}

// extractEmailFromItem extrai email de abuse de um item vCard
func (c *Client) extractEmailFromItem(item interface{}) string {
	itemArray := c.validateVCardItem(item)
	if itemArray == nil {
		return ""
	}

	if !c.isEmailProperty(itemArray) {
		return ""
	}

	email := c.getEmailValue(itemArray)
	if c.isAbuseEmail(email) {
		return email
	}

	return ""
}

// validateVCardItem valida se um item vCard tem o formato esperado
func (c *Client) validateVCardItem(item interface{}) []interface{} {
	itemArray, ok := item.([]interface{})
	if !ok || len(itemArray) < 4 {
		return nil
	}
	return itemArray
}

// isEmailProperty verifica se o item é uma propriedade de email
func (c *Client) isEmailProperty(itemArray []interface{}) bool {
	prop, ok := itemArray[0].(string)
	return ok && prop == "email"
}

// getEmailValue extrai o valor do email do item vCard
func (c *Client) getEmailValue(itemArray []interface{}) string {
	email, ok := itemArray[3].(string)
	if !ok {
		return ""
	}
	return email
}

// isAbuseEmail verifica se o email é um email de abuse
func (c *Client) isAbuseEmail(email string) bool {
	return email != "" && strings.Contains(strings.ToLower(email), "abuse")
}

// getRDAPURL determina a URL do servidor RDAP baseado no TLD
func (c *Client) getRDAPURL(domain string) (string, error) {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid domain format")
	}

	tld := parts[len(parts)-1]

	// Mapeamento de TLDs conhecidos para servidores RDAP
	rdapServers := map[string]string{
		"com":    "https://rdap.verisign.com/com/v1",
		"net":    "https://rdap.verisign.com/net/v1",
		"org":    "https://rdap.publicinterestregistry.org",
		"br":     "https://rdap.registro.br",
		"info":   "https://rdap.afilias.net/rdap/afilias",
		"biz":    "https://rdap.afilias.net/rdap/afilias",
		"name":   "https://rdap.verisign.com/name/v1",
		"mobi":   "https://rdap.afilias.net/rdap/afilias",
		"pro":    "https://rdap.afilias.net/rdap/afilias",
		"travel": "https://rdap.nic.travel",
		"xxx":    "https://rdap.centralnic.com/xxx",
		"jobs":   "https://rdap.afilias.net/rdap/afilias",
		"cat":    "https://rdap.centralnic.com/cat",
		"tel":    "https://rdap.centralnic.com/tel",
	}

	if url, exists := rdapServers[tld]; exists {
		return url, nil
	}

	// Fallback para bootstrap IANA
	return fmt.Sprintf("https://rdap-bootstrap.arin.net/bootstrap/domain/%s", domain), nil
}
