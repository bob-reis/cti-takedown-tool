package models

import "time"

// DNSRecord representa registros DNS coletados
type DNSRecord struct {
	A     []string `json:"A,omitempty"`
	AAAA  []string `json:"AAAA,omitempty"`
	CNAME []string `json:"CNAME,omitempty"`
	MX    []string `json:"MX,omitempty"`
	TXT   []string `json:"TXT,omitempty"`
	NS    []string `json:"NS,omitempty"`
	SOA   string   `json:"SOA,omitempty"`
	TTL   int      `json:"TTL,omitempty"`
}

// HTTPInfo representa informações HTTP coletadas
type HTTPInfo struct {
	Headers    map[string]string `json:"headers"`
	Status     int               `json:"status"`
	Chain      []string          `json:"chain,omitempty"`      // redirects
	Title      string            `json:"title,omitempty"`      // page title
	Body       string            `json:"body,omitempty"`       // first 1KB of body
	Screenshot string            `json:"screenshot,omitempty"` // path to screenshot
}

// TLSInfo representa informações do certificado TLS
type TLSInfo struct {
	Issuer    string    `json:"issuer"`
	CN        string    `json:"cn"`            // Common Name
	SAN       []string  `json:"san,omitempty"` // Subject Alternative Names
	NotBefore time.Time `json:"not_before"`
	NotAfter  time.Time `json:"not_after"`
	Serial    string    `json:"serial,omitempty"`
	Algorithm string    `json:"algorithm,omitempty"`
}

// RiskAssessment representa avaliação de risco
type RiskAssessment struct {
	Score     int    `json:"score"`     // 0-100
	Rationale string `json:"rationale"` // explicação do score
	Category  string `json:"category"`  // phishing, malware, c2, etc
}

// EvidencePack representa o pacote de evidências conforme spec 8.2
type EvidencePack struct {
	EvidenceID  string         `json:"evidence_id"`
	IOC         string         `json:"ioc"` // IOC ID relacionado
	CollectedAt time.Time      `json:"collected_at"`
	Screenshots []string       `json:"screenshots"`   // paths to files
	HAR         string         `json:"har,omitempty"` // path to HAR file
	DNS         DNSRecord      `json:"dns"`
	HTTP        HTTPInfo       `json:"http"`
	TLS         *TLSInfo       `json:"tls,omitempty"`
	IntelRefs   []string       `json:"intel_refs,omitempty"` // external references
	Risk        RiskAssessment `json:"risk"`
	Defanged    string         `json:"defanged"` // defanged version of IOC
}

// GetDefangedURL retorna uma versão defanged da URL para comunicação
func (e *EvidencePack) GetDefangedURL(original string) string {
	if e.Defanged != "" {
		return e.Defanged
	}

	// Implementação básica de defang
	defanged := original
	defanged = replaceInString(defanged, "http://", "hxxp://")
	defanged = replaceInString(defanged, "https://", "hxxps://")
	defanged = replaceInString(defanged, ".", "[.]")

	return defanged
}

func replaceInString(s, old, new string) string {
	// Simple string replacement
	result := ""
	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}
