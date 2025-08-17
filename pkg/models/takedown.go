package models

import "time"

// TakedownStatus representa os estados do takedown conforme spec seção 2
type TakedownStatus string

const (
	StatusDiscovered   TakedownStatus = "discovered"
	StatusTriage       TakedownStatus = "triage"
	StatusEvidencePack TakedownStatus = "evidence_pack"
	StatusRoute        TakedownStatus = "route"
	StatusSubmit       TakedownStatus = "submit"
	StatusSubmitted    TakedownStatus = "submitted"
	StatusAcked        TakedownStatus = "acked"
	StatusFollowUp     TakedownStatus = "follow_up"
	StatusOutcome      TakedownStatus = "outcome"
	StatusClosed       TakedownStatus = "closed"
)

// TakedownAction representa a ação solicitada
type TakedownAction string

const (
	ActionSuspendDomain TakedownAction = "suspend_domain"
	ActionRemoveContent TakedownAction = "remove_content"
	ActionBlockNS       TakedownAction = "block_ns"
	ActionWarningList   TakedownAction = "warning_list"
	ActionBlocklist     TakedownAction = "blocklist"
)

// TakedownTarget representa um alvo para o takedown
type TakedownTarget struct {
	Type    string `json:"type"`   // registrar, hosting, cdn, search, blocklist
	Entity  string `json:"entity"` // nome da entidade
	Email   string `json:"email,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Webform string `json:"webform,omitempty"`
}

// SLA representa configurações de SLA
type SLA struct {
	FirstResponseHours int `json:"first_response_hours"`
	EscalateAfterHours int `json:"escalate_after_hours"`
	RetryIntervalHours int `json:"retry_interval_hours"`
}

// TakedownEvent representa um evento no histórico
type TakedownEvent struct {
	Timestamp time.Time `json:"t"`
	Event     string    `json:"event"`
	Channel   string    `json:"channel,omitempty"` // email, webform, api
	Reference string    `json:"ref,omitempty"`     // case ID, ticket number
	Notes     string    `json:"notes,omitempty"`
}

// TakedownRequest representa uma solicitação de takedown conforme spec 8.4
type TakedownRequest struct {
	CaseID          string          `json:"case_id"`
	Target          TakedownTarget  `json:"target"`
	EvidenceID      string          `json:"evidence_id"`
	RequestedAction TakedownAction  `json:"requested_action"`
	Status          TakedownStatus  `json:"status"`
	SLA             SLA             `json:"sla"`
	History         []TakedownEvent `json:"history"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	NextActionAt    *time.Time      `json:"next_action_at,omitempty"`
	ExternalCaseID  string          `json:"external_case_id,omitempty"`
	Priority        string          `json:"priority"` // low, medium, high, critical
	Assignee        string          `json:"assignee,omitempty"`
	Tags            []string        `json:"tags,omitempty"`
}

// AddEvent adiciona um evento ao histórico
func (tr *TakedownRequest) AddEvent(event, channel, reference, notes string) {
	tr.History = append(tr.History, TakedownEvent{
		Timestamp: time.Now().UTC(),
		Event:     event,
		Channel:   channel,
		Reference: reference,
		Notes:     notes,
	})
	tr.UpdatedAt = time.Now().UTC()
}

// UpdateStatus atualiza o status e adiciona evento
func (tr *TakedownRequest) UpdateStatus(newStatus TakedownStatus, notes string) {
	oldStatus := tr.Status
	tr.Status = newStatus
	tr.AddEvent(string(newStatus), "", "", notes)

	// Calcular próxima ação baseada no SLA
	tr.calculateNextAction()

	// Log da transição
	if oldStatus != newStatus {
		tr.AddEvent("status_change", "", "",
			"Changed from "+string(oldStatus)+" to "+string(newStatus))
	}
}

// IsOverdue verifica se o takedown está atrasado
func (tr *TakedownRequest) IsOverdue() bool {
	if tr.NextActionAt == nil {
		return false
	}
	return time.Now().UTC().After(*tr.NextActionAt)
}

// GetAge retorna a idade do caso em horas
func (tr *TakedownRequest) GetAge() float64 {
	return time.Since(tr.CreatedAt).Hours()
}

// calculateNextAction calcula a próxima ação baseada no status e SLA
func (tr *TakedownRequest) calculateNextAction() {
	now := time.Now().UTC()

	switch tr.Status {
	case StatusSubmitted:
		// Aguardar primeira resposta
		nextAction := tr.CreatedAt.Add(time.Duration(tr.SLA.FirstResponseHours) * time.Hour)
		tr.NextActionAt = &nextAction

	case StatusAcked:
		// Aguardar follow-up
		lastEvent := tr.getLastEventTime()
		nextAction := lastEvent.Add(time.Duration(tr.SLA.RetryIntervalHours) * time.Hour)
		tr.NextActionAt = &nextAction

	case StatusFollowUp:
		// Verificar se deve escalar
		if tr.GetAge() > float64(tr.SLA.EscalateAfterHours) {
			nextAction := now.Add(24 * time.Hour) // escalar em 24h
			tr.NextActionAt = &nextAction
		} else {
			// Próximo follow-up
			lastEvent := tr.getLastEventTime()
			nextAction := lastEvent.Add(time.Duration(tr.SLA.RetryIntervalHours) * time.Hour)
			tr.NextActionAt = &nextAction
		}

	default:
		tr.NextActionAt = nil
	}
}

// getLastEventTime retorna o timestamp do último evento
func (tr *TakedownRequest) getLastEventTime() time.Time {
	if len(tr.History) == 0 {
		return tr.CreatedAt
	}
	return tr.History[len(tr.History)-1].Timestamp
}
