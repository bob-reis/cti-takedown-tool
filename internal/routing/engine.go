package routing

import (
	"strings"

	"github.com/cti-team/takedown/pkg/models"
)

// Engine determina ações de takedown baseado em regras
type Engine struct {
	rules []Rule
}

// Rule representa uma regra de roteamento
type Rule struct {
	Match   []string           // tags que devem estar presentes
	Actions []ActionDefinition // ações a serem executadas
}

// ActionDefinition define uma ação específica
type ActionDefinition struct {
	Target models.TakedownTarget
	Action models.TakedownAction
	SLA    models.SLA
}

// NewEngine cria uma nova engine de roteamento
func NewEngine() *Engine {
	engine := &Engine{}
	engine.loadDefaultRules()
	return engine
}

// loadDefaultRules carrega regras padrão baseadas na spec
func (e *Engine) loadDefaultRules() {
	e.rules = []Rule{
		// Regra para phishing
		{
			Match: []string{"phishing"},
			Actions: []ActionDefinition{
				{
					Target: models.TakedownTarget{Type: "registrar"},
					Action: models.ActionSuspendDomain,
					SLA: models.SLA{
						FirstResponseHours: 48,
						EscalateAfterHours: 120,
						RetryIntervalHours: 48,
					},
				},
				{
					Target: models.TakedownTarget{Type: "hosting"},
					Action: models.ActionRemoveContent,
					SLA: models.SLA{
						FirstResponseHours: 48,
						EscalateAfterHours: 96,
						RetryIntervalHours: 24,
					},
				},
				{
					Target: models.TakedownTarget{Type: "search"},
					Action: models.ActionWarningList,
					SLA: models.SLA{
						FirstResponseHours: 24,
						EscalateAfterHours: 72,
						RetryIntervalHours: 24,
					},
				},
				{
					Target: models.TakedownTarget{Type: "blocklist"},
					Action: models.ActionBlocklist,
					SLA: models.SLA{
						FirstResponseHours: 24,
						EscalateAfterHours: 72,
						RetryIntervalHours: 24,
					},
				},
			},
		},

		// Regra para malware
		{
			Match: []string{"malware"},
			Actions: []ActionDefinition{
				{
					Target: models.TakedownTarget{Type: "hosting"},
					Action: models.ActionRemoveContent,
					SLA: models.SLA{
						FirstResponseHours: 24,
						EscalateAfterHours: 72,
						RetryIntervalHours: 24,
					},
				},
				{
					Target: models.TakedownTarget{Type: "blocklist"},
					Action: models.ActionBlocklist,
					SLA: models.SLA{
						FirstResponseHours: 24,
						EscalateAfterHours: 72,
						RetryIntervalHours: 24,
					},
				},
			},
		},

		// Regra para C2
		{
			Match: []string{"c2"},
			Actions: []ActionDefinition{
				{
					Target: models.TakedownTarget{Type: "hosting"},
					Action: models.ActionRemoveContent,
					SLA: models.SLA{
						FirstResponseHours: 12,
						EscalateAfterHours: 48,
						RetryIntervalHours: 12,
					},
				},
				{
					Target: models.TakedownTarget{Type: "registrar"},
					Action: models.ActionSuspendDomain,
					SLA: models.SLA{
						FirstResponseHours: 24,
						EscalateAfterHours: 72,
						RetryIntervalHours: 24,
					},
				},
			},
		},

		// Regra para typosquatting/brand (marca)
		{
			Match: []string{"brand:*"},
			Actions: []ActionDefinition{
				{
					Target: models.TakedownTarget{Type: "registrar"},
					Action: models.ActionSuspendDomain,
					SLA: models.SLA{
						FirstResponseHours: 72,
						EscalateAfterHours: 168, // 7 dias
						RetryIntervalHours: 72,
					},
				},
			},
		},
	}
}

// DetermineActions determina as ações necessárias baseado nas tags e contatos
func (e *Engine) DetermineActions(tags []string, contacts *models.AbuseContact) []ActionDefinition {
	var actions []ActionDefinition

	// Encontrar regras que fazem match com as tags
	for _, rule := range e.rules {
		if e.matchRule(rule.Match, tags) {
			// Aplicar ações da regra, populando com contatos reais
			for _, actionDef := range rule.Actions {
				enrichedAction := e.enrichAction(actionDef, contacts)
				if enrichedAction != nil {
					actions = append(actions, *enrichedAction)
				}
			}
		}
	}

	// Remover duplicatas e priorizar
	return e.prioritizeActions(actions)
}

// matchRule verifica se as tags fazem match com os critérios da regra
func (e *Engine) matchRule(ruleMatch, tags []string) bool {
	for _, requiredTag := range ruleMatch {
		found := false
		for _, tag := range tags {
			// Suporte para wildcards simples
			if strings.HasSuffix(requiredTag, "*") {
				prefix := requiredTag[:len(requiredTag)-1]
				if strings.HasPrefix(tag, prefix) {
					found = true
					break
				}
			} else if tag == requiredTag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// enrichAction enriquece uma ação com informações de contato reais
func (e *Engine) enrichAction(actionDef ActionDefinition, contacts *models.AbuseContact) *ActionDefinition {
	enriched := actionDef

	switch actionDef.Target.Type {
	case "registrar":
		if contacts.Registrar != nil {
			enriched.Target.Entity = contacts.Registrar.Name
			enriched.Target.Email = contacts.GetPrimaryAbuseEmail()
		} else {
			return nil // Sem registrar disponível
		}

	case "hosting":
		if contacts.Hosting != nil {
			enriched.Target.Entity = contacts.Hosting.Name
			enriched.Target.Email = contacts.Hosting.Abuse.Email
		} else {
			return nil // Sem hosting disponível
		}

	case "cdn":
		if contacts.CDN != nil {
			enriched.Target.Entity = contacts.CDN.Name
			enriched.Target.Email = contacts.CDN.Abuse.Email
			enriched.Target.Webform = contacts.CDN.Webform
		} else {
			return nil // Sem CDN disponível
		}

	case "search":
		// Search engines/warnings - usar contatos padrão
		enriched.Target.Entity = "Google Safe Browsing"
		enriched.Target.Webform = "https://safebrowsing.google.com/safebrowsing/report_phish/"

	case "blocklist":
		// Blocklists - usar contatos padrão
		enriched.Target.Entity = "URLhaus"
		enriched.Target.Webform = "https://urlhaus.abuse.ch/browse/"

	default:
		return nil // Tipo não suportado
	}

	return &enriched
}

// prioritizeActions remove duplicatas e prioriza ações
func (e *Engine) prioritizeActions(actions []ActionDefinition) []ActionDefinition {
	// Mapa para remover duplicatas baseado no tipo de target
	seen := make(map[string]ActionDefinition)

	// Ordem de prioridade para tipos de target
	priority := map[string]int{
		"hosting":   1, // Mais rápido para remover conteúdo
		"cdn":       2, // CDN pode ser rápido também
		"registrar": 3, // Registrar é mais demorado mas mais efetivo
		"search":    4, // Warnings são complementares
		"blocklist": 5, // Blocklists são complementares
	}

	for _, action := range actions {
		key := action.Target.Type

		// Manter apenas a ação de maior prioridade para cada tipo
		if existing, exists := seen[key]; !exists || priority[action.Target.Type] < priority[existing.Target.Type] {
			seen[key] = action
		}
	}

	// Converter de volta para slice
	var result []ActionDefinition
	for _, action := range seen {
		result = append(result, action)
	}

	return result
}

// AddRule adiciona uma nova regra de roteamento
func (e *Engine) AddRule(rule Rule) {
	e.rules = append(e.rules, rule)
}

// GetRules retorna todas as regras configuradas
func (e *Engine) GetRules() []Rule {
	return e.rules
}
