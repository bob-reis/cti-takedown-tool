package models

import "time"

// IOCType representa os tipos de indicadores suportados
type IOCType string

const (
	IOCTypeURL    IOCType = "url"
	IOCTypeDomain IOCType = "domain"
	IOCTypeIP     IOCType = "ip"
	IOCTypeHash   IOCType = "hash"
)

// IOC representa um indicador de comprometimento conforme spec 8.1
type IOC struct {
	IndicatorID string    `json:"indicator_id"`
	Type        IOCType   `json:"type"`
	Value       string    `json:"value"`
	FirstSeen   time.Time `json:"first_seen"`
	Source      string    `json:"source"`
	Tags        []string  `json:"tags"`
}

// GetBrand extrai a tag de marca do IOC (ex: "brand:AcmeBank" -> "AcmeBank")
func (ioc *IOC) GetBrand() string {
	for _, tag := range ioc.Tags {
		if len(tag) > 6 && tag[:6] == "brand:" {
			return tag[6:]
		}
	}
	return ""
}

// HasTag verifica se o IOC possui uma tag específica
func (ioc *IOC) HasTag(tag string) bool {
	for _, t := range ioc.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetSeverity retorna a severidade baseada nas tags
func (ioc *IOC) GetSeverity() string {
	severityTags := []string{"critical", "high", "medium", "low"}
	for _, severity := range severityTags {
		if ioc.HasTag(severity) {
			return severity
		}
	}
	return "medium" // default
}
