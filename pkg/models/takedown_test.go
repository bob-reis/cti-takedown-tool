package models

import (
	"testing"
	"time"
)

func TestTakedownStatus_Constants(t *testing.T) {
	statuses := []TakedownStatus{
		StatusDiscovered,
		StatusTriage,
		StatusEvidencePack,
		StatusRoute,
		StatusSubmit,
		StatusSubmitted,
		StatusAcked,
		StatusFollowUp,
		StatusOutcome,
		StatusClosed,
	}

	expectedStatuses := []string{
		"discovered",
		"triage",
		"evidence_pack",
		"route",
		"submit",
		"submitted",
		"acked",
		"follow_up",
		"outcome",
		"closed",
	}

	for i, status := range statuses {
		if string(status) != expectedStatuses[i] {
			t.Errorf("Status %d should be %s, got %s", i, expectedStatuses[i], string(status))
		}
	}
}

func TestTakedownAction_Constants(t *testing.T) {
	actions := []TakedownAction{
		ActionSuspendDomain,
		ActionRemoveContent,
		ActionBlockNS,
		ActionWarningList,
		ActionBlocklist,
	}

	expectedActions := []string{
		"suspend_domain",
		"remove_content",
		"block_ns",
		"warning_list",
		"blocklist",
	}

	for i, action := range actions {
		if string(action) != expectedActions[i] {
			t.Errorf("Action %d should be %s, got %s", i, expectedActions[i], string(action))
		}
	}
}

func TestTakedownRequest_AddEvent(t *testing.T) {
	request := &TakedownRequest{
		CaseID:    "test-case-123",
		Status:    StatusDiscovered,
		CreatedAt: time.Now().UTC(),
		History:   []TakedownEvent{},
	}

	// Add first event
	request.AddEvent("test_event", "email", "ref123", "Test event notes")

	if len(request.History) != 1 {
		t.Errorf("History should have 1 event, got %d", len(request.History))
	}

	event := request.History[0]
	if event.Event != "test_event" {
		t.Errorf("Event should be 'test_event', got %s", event.Event)
	}
	if event.Channel != "email" {
		t.Errorf("Channel should be 'email', got %s", event.Channel)
	}
	if event.Reference != "ref123" {
		t.Errorf("Reference should be 'ref123', got %s", event.Reference)
	}
	if event.Notes != "Test event notes" {
		t.Errorf("Notes should be 'Test event notes', got %s", event.Notes)
	}

	// Add second event
	request.AddEvent("another_event", "webform", "", "Another note")

	if len(request.History) != 2 {
		t.Errorf("History should have 2 events, got %d", len(request.History))
	}

	// Check that UpdatedAt was updated
	if request.UpdatedAt.IsZero() {
		t.Errorf("UpdatedAt should be set")
	}
}

func TestTakedownRequest_UpdateStatus(t *testing.T) {
	request := &TakedownRequest{
		CaseID:    "test-case-123",
		Status:    StatusDiscovered,
		CreatedAt: time.Now().UTC(),
		History:   []TakedownEvent{},
		SLA: SLA{
			FirstResponseHours: 48,
			EscalateAfterHours: 120,
			RetryIntervalHours: 24,
		},
	}

	oldStatus := request.Status
	request.UpdateStatus(StatusSubmitted, "Moving to submitted")

	// Check status was updated
	if request.Status != StatusSubmitted {
		t.Errorf("Status should be %s, got %s", StatusSubmitted, request.Status)
	}

	// Check that event was added
	if len(request.History) < 1 {
		t.Errorf("History should have at least 1 event")
	}

	// Check for status change event
	statusChangeFound := false
	for _, event := range request.History {
		if event.Event == "status_change" {
			statusChangeFound = true
			expectedNotes := "Changed from " + string(oldStatus) + " to " + string(StatusSubmitted)
			if event.Notes != expectedNotes {
				t.Errorf("Status change notes incorrect, got: %s", event.Notes)
			}
		}
	}
	if !statusChangeFound {
		t.Errorf("Status change event not found in history")
	}

	// Check that NextActionAt was calculated
	if request.NextActionAt == nil {
		t.Errorf("NextActionAt should be set after status update")
	}
}

func TestTakedownRequest_IsOverdue(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	tests := []struct {
		name         string
		nextActionAt *time.Time
		expected     bool
	}{
		{
			name:         "No next action set",
			nextActionAt: nil,
			expected:     false,
		},
		{
			name:         "Next action in future",
			nextActionAt: &future,
			expected:     false,
		},
		{
			name:         "Next action in past - overdue",
			nextActionAt: &past,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &TakedownRequest{
				NextActionAt: tt.nextActionAt,
			}
			result := request.IsOverdue()
			if result != tt.expected {
				t.Errorf("IsOverdue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTakedownRequest_GetAge(t *testing.T) {
	oneHourAgo := time.Now().UTC().Add(-1 * time.Hour)
	request := &TakedownRequest{
		CreatedAt: oneHourAgo,
	}

	age := request.GetAge()

	// Age should be approximately 1 hour (allow some variance for test execution time)
	if age < 0.9 || age > 1.1 {
		t.Errorf("Age should be approximately 1 hour, got %f", age)
	}
}

func TestSLA_Structure(t *testing.T) {
	sla := SLA{
		FirstResponseHours: 48,
		EscalateAfterHours: 120,
		RetryIntervalHours: 24,
	}

	if sla.FirstResponseHours != 48 {
		t.Errorf("FirstResponseHours should be 48, got %d", sla.FirstResponseHours)
	}
	if sla.EscalateAfterHours != 120 {
		t.Errorf("EscalateAfterHours should be 120, got %d", sla.EscalateAfterHours)
	}
	if sla.RetryIntervalHours != 24 {
		t.Errorf("RetryIntervalHours should be 24, got %d", sla.RetryIntervalHours)
	}
}

func TestTakedownTarget_Structure(t *testing.T) {
	target := TakedownTarget{
		Type:    "registrar",
		Entity:  "GoDaddy.com, LLC",
		Email:   "abuse@godaddy.com",
		Phone:   "+1-555-0123",
		Webform: "https://www.godaddy.com/abuse",
	}

	if target.Type != "registrar" {
		t.Errorf("Type should be 'registrar', got %s", target.Type)
	}
	if target.Entity != "GoDaddy.com, LLC" {
		t.Errorf("Entity not correct")
	}
	if target.Email != "abuse@godaddy.com" {
		t.Errorf("Email not correct")
	}
}

func TestTakedownEvent_Structure(t *testing.T) {
	now := time.Now().UTC()
	event := TakedownEvent{
		Timestamp: now,
		Event:     "submitted",
		Channel:   "email",
		Reference: "GD-12345",
		Notes:     "Submitted to GoDaddy via email",
	}

	if event.Event != "submitted" {
		t.Errorf("Event should be 'submitted', got %s", event.Event)
	}
	if event.Channel != "email" {
		t.Errorf("Channel should be 'email', got %s", event.Channel)
	}
	if event.Reference != "GD-12345" {
		t.Errorf("Reference should be 'GD-12345', got %s", event.Reference)
	}
}

func TestTakedownRequest_calculateNextAction(t *testing.T) {
	now := time.Now().UTC()
	tr := &TakedownRequest{
		Status:    StatusSubmitted,
		CreatedAt: now,
		SLA:       SLA{FirstResponseHours: 1, RetryIntervalHours: 2, EscalateAfterHours: 24},
	}
	tr.calculateNextAction()
	if tr.NextActionAt == nil || !tr.NextActionAt.Equal(tr.CreatedAt.Add(1*time.Hour)) {
		t.Fatalf("expected next action 1h after creation, got %v", tr.NextActionAt)
	}

	// Acked uses last event timestamp
	last := now.Add(30 * time.Minute)
	tr.Status = StatusAcked
	tr.History = []TakedownEvent{{Timestamp: last}}
	tr.calculateNextAction()
	expected := last.Add(2 * time.Hour)
	if tr.NextActionAt == nil || !tr.NextActionAt.Equal(expected) {
		t.Fatalf("expected next action %v, got %v", expected, tr.NextActionAt)
	}

	// FollowUp without escalation
	tr.Status = StatusFollowUp
	tr.CreatedAt = now
	tr.History = []TakedownEvent{{Timestamp: last}}
	tr.calculateNextAction()
	expected = last.Add(2 * time.Hour)
	if tr.NextActionAt == nil || !tr.NextActionAt.Equal(expected) {
		t.Fatalf("expected next action %v, got %v", expected, tr.NextActionAt)
	}

	// FollowUp with escalation
	tr.CreatedAt = now.Add(-25 * time.Hour)
	tr.calculateNextAction()
	if tr.NextActionAt == nil {
		t.Fatalf("expected next action for escalation")
	}
	diff := tr.NextActionAt.Sub(time.Now().UTC())
	if diff < 23*time.Hour || diff > 25*time.Hour {
		t.Fatalf("expected escalation around 24h, got %v", diff)
	}

	// Default case
	tr.Status = StatusRoute
	tr.calculateNextAction()
	if tr.NextActionAt != nil {
		t.Fatalf("expected nil next action for status %s", tr.Status)
	}
}

func TestTakedownRequest_getLastEventTime(t *testing.T) {
	now := time.Now().UTC()
	tr := &TakedownRequest{CreatedAt: now}
	if got := tr.getLastEventTime(); !got.Equal(now) {
		t.Fatalf("expected %v, got %v", now, got)
	}
	later := now.Add(2 * time.Hour)
	tr.History = []TakedownEvent{{Timestamp: later}}
	if got := tr.getLastEventTime(); !got.Equal(later) {
		t.Fatalf("expected %v, got %v", later, got)
	}
}
