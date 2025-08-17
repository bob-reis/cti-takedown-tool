package state

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/cti-team/takedown/internal/enrichment"
	"github.com/cti-team/takedown/internal/evidence"
	"github.com/cti-team/takedown/internal/routing"
	"github.com/cti-team/takedown/pkg/models"
	"github.com/google/uuid"
)

// Machine representa a state machine para orquestração de takedowns
type Machine struct {
	collector  *evidence.Collector
	enricher   *enrichment.Service
	router     *routing.Engine
	connectors map[string]Connector
	requests   map[string]*models.TakedownRequest
	mutex      sync.RWMutex
	workers    int
	workChan   chan *models.TakedownRequest
	stopChan   chan struct{}
	ticker     *time.Ticker
}

// Connector representa um conector para diferentes tipos de targets
type Connector interface {
	Submit(ctx context.Context, request *models.TakedownRequest, evidence *models.EvidencePack) error
	CheckStatus(ctx context.Context, request *models.TakedownRequest) (*StatusUpdate, error)
	GetType() string
}

// StatusUpdate representa uma atualização de status de um conector
type StatusUpdate struct {
	Status       models.TakedownStatus
	ExternalID   string
	Notes        string
	NextFollowUp *time.Time
}

// NewMachine cria uma nova state machine
func NewMachine(collector *evidence.Collector, enricher *enrichment.Service, router *routing.Engine) *Machine {
	return &Machine{
		collector:  collector,
		enricher:   enricher,
		router:     router,
		connectors: make(map[string]Connector),
		requests:   make(map[string]*models.TakedownRequest),
		workers:    5,
		workChan:   make(chan *models.TakedownRequest, 100),
		stopChan:   make(chan struct{}),
		ticker:     time.NewTicker(1 * time.Minute), // Check a cada minuto
	}
}

// RegisterConnector registra um conector para um tipo de target
func (m *Machine) RegisterConnector(connector Connector) {
	m.connectors[connector.GetType()] = connector
}

// Start inicia a state machine
func (m *Machine) Start() {
	log.Println("Starting takedown state machine...")

	// Iniciar workers
	for i := 0; i < m.workers; i++ {
		go m.worker(i)
	}

	// Iniciar scheduler para verificar SLAs
	go m.scheduler()

	log.Printf("Started %d workers and scheduler", m.workers)
}

// Stop para a state machine
func (m *Machine) Stop() {
	log.Println("Stopping takedown state machine...")
	close(m.stopChan)
	m.ticker.Stop()
}

// ProcessIOC processa um novo IOC através da pipeline completa
func (m *Machine) ProcessIOC(ioc *models.IOC) error {
	// Criar caso base
	caseID := fmt.Sprintf("tdk-%s", uuid.New().String())

	request := &models.TakedownRequest{
		CaseID:    caseID,
		Status:    models.StatusDiscovered,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Priority:  ioc.GetSeverity(),
		Tags:      ioc.Tags,
	}

	request.AddEvent("case_created", "system", "", fmt.Sprintf("Processing IOC: %s", ioc.Value))

	// Registrar caso
	m.mutex.Lock()
	m.requests[caseID] = request
	m.mutex.Unlock()

	// Iniciar processamento
	return m.transitionTo(request, models.StatusTriage)
}

// transitionTo move o request para um novo estado
func (m *Machine) transitionTo(request *models.TakedownRequest, newStatus models.TakedownStatus) error {
	oldStatus := request.Status
	request.UpdateStatus(newStatus, fmt.Sprintf("Transitioned from %s to %s", oldStatus, newStatus))

	log.Printf("Case %s: %s -> %s", request.CaseID, oldStatus, newStatus)

	// Adicionar à fila de processamento
	select {
	case m.workChan <- request:
		return nil
	default:
		return fmt.Errorf("work queue is full")
	}
}

// worker processa requests da fila
func (m *Machine) worker(id int) {
	log.Printf("Worker %d started", id)

	for {
		select {
		case request := <-m.workChan:
			if err := m.processRequest(request); err != nil {
				log.Printf("Worker %d: Error processing %s: %v", id, request.CaseID, err)
				request.AddEvent("error", "system", "", err.Error())
			}

		case <-m.stopChan:
			log.Printf("Worker %d stopped", id)
			return
		}
	}
}

// processRequest processa um request baseado no seu estado atual
func (m *Machine) processRequest(request *models.TakedownRequest) error {
	ctx := context.Background()

	switch request.Status {
	case models.StatusTriage:
		return m.handleTriage(ctx, request)

	case models.StatusEvidencePack:
		return m.handleEvidenceCollection(ctx, request)

	case models.StatusRoute:
		return m.handleRouting(ctx, request)

	case models.StatusSubmit:
		return m.handleSubmission(ctx, request)

	case models.StatusFollowUp:
		return m.handleFollowUp(ctx, request)

	default:
		return fmt.Errorf("unknown status: %s", request.Status)
	}
}

// handleTriage realiza triagem do caso
func (m *Machine) handleTriage(ctx context.Context, request *models.TakedownRequest) error {
	// Análise básica de prioridade e validade
	request.AddEvent("triage_started", "system", "", "Starting triage analysis")

	// Por enquanto, aprovamos todos os casos
	// TODO: Implementar regras de triagem mais sofisticadas

	return m.transitionTo(request, models.StatusEvidencePack)
}

// handleEvidenceCollection coleta evidências
func (m *Machine) handleEvidenceCollection(ctx context.Context, request *models.TakedownRequest) error {
	request.AddEvent("evidence_collection_started", "system", "", "Starting evidence collection")

	// Buscar IOC original
	// TODO: Implementar storage/retrieve de IOCs
	// Por enquanto, vamos simular

	ioc := &models.IOC{
		IndicatorID: "temp-ioc",
		Type:        models.IOCTypeURL,
		Value:       "https://suspicious-domain.com/login",
		Tags:        request.Tags,
	}

	// Coletar evidências
	evidence, err := m.collector.CollectEvidence(ioc)
	if err != nil {
		return fmt.Errorf("evidence collection failed: %w", err)
	}

	request.EvidenceID = evidence.EvidenceID
	request.AddEvent("evidence_collected", "system", evidence.EvidenceID,
		fmt.Sprintf("Evidence collected, risk score: %d", evidence.Risk.Score))

	return m.transitionTo(request, models.StatusRoute)
}

// handleRouting determina targets para o takedown
func (m *Machine) handleRouting(ctx context.Context, request *models.TakedownRequest) error {
	request.AddEvent("routing_started", "system", "", "Determining takedown targets")

	// Usar enrichment service para descobrir contatos
	contacts, err := m.enricher.EnrichIOC(ctx, request.EvidenceID)
	if err != nil {
		return fmt.Errorf("enrichment failed: %w", err)
	}

	// Usar routing engine para determinar actions
	actions := m.router.DetermineActions(request.Tags, contacts)

	if len(actions) == 0 {
		request.AddEvent("no_actions", "system", "", "No valid targets found")
		return m.transitionTo(request, models.StatusClosed)
	}

	// Criar requests para cada target
	for _, action := range actions {
		// Por enquanto, usar o primeiro target
		request.Target = action.Target
		request.RequestedAction = action.Action
		break
	}

	request.AddEvent("routing_completed", "system", "",
		fmt.Sprintf("Target: %s (%s)", request.Target.Entity, request.Target.Type))

	return m.transitionTo(request, models.StatusSubmit)
}

// handleSubmission submete o takedown request
func (m *Machine) handleSubmission(ctx context.Context, request *models.TakedownRequest) error {
	connector, exists := m.connectors[request.Target.Type]
	if !exists {
		return fmt.Errorf("no connector found for type: %s", request.Target.Type)
	}

	request.AddEvent("submission_started", "system", "",
		fmt.Sprintf("Submitting to %s", request.Target.Entity))

	// TODO: Carregar evidence pack
	evidence := &models.EvidencePack{} // placeholder

	err := connector.Submit(ctx, request, evidence)
	if err != nil {
		return fmt.Errorf("submission failed: %w", err)
	}

	request.AddEvent("submitted", "connector", "", "Successfully submitted to target")

	return m.transitionTo(request, models.StatusSubmitted)
}

// handleFollowUp realiza follow-up de casos
func (m *Machine) handleFollowUp(ctx context.Context, request *models.TakedownRequest) error {
	connector, exists := m.connectors[request.Target.Type]
	if !exists {
		return fmt.Errorf("no connector found for type: %s", request.Target.Type)
	}

	status, err := connector.CheckStatus(ctx, request)
	if err != nil {
		log.Printf("Status check failed for %s: %v", request.CaseID, err)
		// Agendar próximo follow-up
		nextTime := time.Now().Add(time.Duration(request.SLA.RetryIntervalHours) * time.Hour)
		request.NextActionAt = &nextTime
		return nil
	}

	if status.ExternalID != "" {
		request.ExternalCaseID = status.ExternalID
	}

	request.AddEvent("status_update", "connector", status.ExternalID, status.Notes)

	// Verificar se o caso foi resolvido
	if status.Status == models.StatusOutcome {
		return m.transitionTo(request, models.StatusOutcome)
	}

	// Agendar próximo follow-up
	if status.NextFollowUp != nil {
		request.NextActionAt = status.NextFollowUp
	}

	return nil
}

// scheduler verifica periodicamente casos que precisam de ação
func (m *Machine) scheduler() {
	for {
		select {
		case <-m.ticker.C:
			m.checkPendingActions()

		case <-m.stopChan:
			return
		}
	}
}

// checkPendingActions verifica casos que precisam de follow-up
func (m *Machine) checkPendingActions() {
	now := time.Now().UTC()

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, request := range m.requests {
		if m.shouldProcessRequest(request, now) {
			m.processScheduledRequest(request)
		}
	}
}

// shouldProcessRequest verifica se um request deve ser processado agora
func (m *Machine) shouldProcessRequest(request *models.TakedownRequest, now time.Time) bool {
	return request.NextActionAt != nil && now.After(*request.NextActionAt)
}

// processScheduledRequest processa um request agendado baseado no seu status
func (m *Machine) processScheduledRequest(request *models.TakedownRequest) {
	switch request.Status {
	case models.StatusSubmitted:
		m.promoteToFollowUp(request)
	case models.StatusFollowUp:
		m.handleScheduledFollowUp(request)
	}
}

// promoteToFollowUp promove um request submetido para follow-up
func (m *Machine) promoteToFollowUp(request *models.TakedownRequest) {
	if err := m.transitionTo(request, models.StatusFollowUp); err != nil {
		request.AddEvent("transition_failed", "system", "",
			fmt.Sprintf("failed to transition: %v", err))
	}
}

// handleScheduledFollowUp processa um follow-up agendado
func (m *Machine) handleScheduledFollowUp(request *models.TakedownRequest) {
	if m.shouldEscalate(request) {
		m.escalateRequest(request)
	} else {
		m.continueFollowUp(request)
	}
}

// shouldEscalate verifica se um request deve ser escalado
func (m *Machine) shouldEscalate(request *models.TakedownRequest) bool {
	return request.GetAge() > float64(request.SLA.EscalateAfterHours)
}

// escalateRequest escala um request que está atrasado
func (m *Machine) escalateRequest(request *models.TakedownRequest) {
	overdueHours := request.GetAge() - float64(request.SLA.EscalateAfterHours)
	request.AddEvent("escalation_needed", "system", "",
		fmt.Sprintf("Case overdue by %.1f hours", overdueHours))
}

// continueFollowUp continua o follow-up de um request
func (m *Machine) continueFollowUp(request *models.TakedownRequest) {
	if err := m.transitionTo(request, models.StatusFollowUp); err != nil {
		request.AddEvent("transition_failed", "system", "",
			fmt.Sprintf("failed to transition: %v", err))
	}
}

// GetRequest retorna informações de um request
func (m *Machine) GetRequest(caseID string) (*models.TakedownRequest, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	request, exists := m.requests[caseID]
	return request, exists
}

// ListRequests retorna todos os requests
func (m *Machine) ListRequests() []*models.TakedownRequest {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var requests []*models.TakedownRequest
	for _, request := range m.requests {
		requests = append(requests, request)
	}

	return requests
}
