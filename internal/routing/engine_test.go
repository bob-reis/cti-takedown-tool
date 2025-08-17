package routing

import (
	"testing"

	"github.com/cti-team/takedown/pkg/models"
)

func TestEngine_DetermineActions_Phishing(t *testing.T) {
	engine := NewEngine()

	// Create mock contacts
	contacts := &models.AbuseContact{
		Domain: "phishing-site.com",
		Registrar: &models.RegistrarInfo{
			Name:   "GoDaddy.com, LLC",
			IANAID: 146,
		},
		Abuse: models.ContactInfo{
			Email: "abuse@godaddy.com",
		},
		Hosting: &models.HostingInfo{
			ASN:  12345,
			Name: "Example Hosting",
			Abuse: models.ContactInfo{
				Email: "abuse@examplehosting.com",
			},
		},
	}

	tags := []string{"phishing", "brand:TestBank"}
	actions := engine.DetermineActions(tags, contacts)

	// Phishing should target registrar, hosting, search, and blocklists
	if len(actions) == 0 {
		t.Fatalf("Expected actions for phishing, got none")
	}

	// Check that we have registrar and hosting actions
	hasRegistrar := false
	hasHosting := false

	for _, action := range actions {
		switch action.Target.Type {
		case "registrar":
			hasRegistrar = true
			if action.Target.Entity != "GoDaddy.com, LLC" {
				t.Errorf("Expected registrar entity 'GoDaddy.com, LLC', got %s", action.Target.Entity)
			}
			if action.Action != models.ActionSuspendDomain {
				t.Errorf("Expected suspend_domain action for registrar, got %s", action.Action)
			}
		case "hosting":
			hasHosting = true
			if action.Target.Entity != "Example Hosting" {
				t.Errorf("Expected hosting entity 'Example Hosting', got %s", action.Target.Entity)
			}
			if action.Action != models.ActionRemoveContent {
				t.Errorf("Expected remove_content action for hosting, got %s", action.Action)
			}
		}
	}

	if !hasRegistrar {
		t.Errorf("Missing registrar action for phishing")
	}
	if !hasHosting {
		t.Errorf("Missing hosting action for phishing")
	}
}

func TestEngine_DetermineActions_Malware(t *testing.T) {
	engine := NewEngine()

	contacts := &models.AbuseContact{
		Hosting: &models.HostingInfo{
			ASN:  54321,
			Name: "Malware Host Inc",
			Abuse: models.ContactInfo{
				Email: "abuse@malwarehost.com",
			},
		},
	}

	tags := []string{"malware"}
	actions := engine.DetermineActions(tags, contacts)

	if len(actions) == 0 {
		t.Fatalf("Expected actions for malware, got none")
	}

	// Check for hosting action
	hasHosting := false
	for _, action := range actions {
		if action.Target.Type == "hosting" {
			hasHosting = true
			if action.Action != models.ActionRemoveContent {
				t.Errorf("Expected remove_content action for malware hosting, got %s", action.Action)
			}
		}
	}

	if !hasHosting {
		t.Errorf("Missing hosting action for malware")
	}
}

func TestEngine_DetermineActions_C2(t *testing.T) {
	engine := NewEngine()

	contacts := &models.AbuseContact{
		Registrar: &models.RegistrarInfo{
			Name: "Example Registrar",
		},
		Hosting: &models.HostingInfo{
			ASN:  99999,
			Name: "C2 Infrastructure Host",
			Abuse: models.ContactInfo{
				Email: "abuse@c2host.com",
			},
		},
	}

	tags := []string{"c2", "critical"}
	actions := engine.DetermineActions(tags, contacts)

	if len(actions) == 0 {
		t.Fatalf("Expected actions for C2, got none")
	}

	// C2 should have critical SLA override
	hasHosting := false
	hasRegistrar := false

	for _, action := range actions {
		switch action.Target.Type {
		case "hosting":
			hasHosting = true
			// Check SLA is tightened for critical
			if action.SLA.FirstResponseHours != 12 {
				t.Errorf("Expected 12h first response for C2 hosting, got %d", action.SLA.FirstResponseHours)
			}
		case "registrar":
			hasRegistrar = true
		}
	}

	if !hasHosting {
		t.Errorf("Missing hosting action for C2")
	}
	if !hasRegistrar {
		t.Errorf("Missing registrar action for C2")
	}
}

func TestEngine_DetermineActions_Brand(t *testing.T) {
	engine := NewEngine()

	contacts := &models.AbuseContact{
		Registrar: &models.RegistrarInfo{
			Name: "Brand Dispute Registrar",
		},
		Abuse: models.ContactInfo{
			Email: "abuse@registrar.com",
		},
	}

	tags := []string{"brand:TestBank", "typosquatting"}
	actions := engine.DetermineActions(tags, contacts)

	if len(actions) == 0 {
		t.Fatalf("Expected actions for brand dispute, got none")
	}

	// Brand disputes should primarily target registrar
	hasRegistrar := false
	for _, action := range actions {
		if action.Target.Type == "registrar" {
			hasRegistrar = true
			if action.Action != models.ActionSuspendDomain {
				t.Errorf("Expected suspend_domain for brand dispute, got %s", action.Action)
			}
			// Brand disputes have longer SLAs
			if action.SLA.FirstResponseHours != 72 {
				t.Errorf("Expected 72h first response for brand disputes, got %d", action.SLA.FirstResponseHours)
			}
		}
	}

	if !hasRegistrar {
		t.Errorf("Missing registrar action for brand dispute")
	}
}

func TestEngine_MatchRule(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name      string
		ruleMatch []string
		tags      []string
		expected  bool
	}{
		{
			name:      "Exact match",
			ruleMatch: []string{"phishing"},
			tags:      []string{"phishing", "high"},
			expected:  true,
		},
		{
			name:      "Multiple requirements - all present",
			ruleMatch: []string{"phishing", "brand"},
			tags:      []string{"phishing", "brand:TestBank", "high"},
			expected:  false, // brand != brand:TestBank exact match
		},
		{
			name:      "Wildcard match",
			ruleMatch: []string{"brand:*"},
			tags:      []string{"phishing", "brand:TestBank"},
			expected:  true,
		},
		{
			name:      "Missing required tag",
			ruleMatch: []string{"phishing", "malware"},
			tags:      []string{"phishing", "high"},
			expected:  false,
		},
		{
			name:      "Empty rule match",
			ruleMatch: []string{},
			tags:      []string{"phishing"},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.matchRule(tt.ruleMatch, tt.tags)
			if result != tt.expected {
				t.Errorf("matchRule() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEngine_EnrichAction(t *testing.T) {
	engine := NewEngine()

	contacts := &models.AbuseContact{
		Registrar: &models.RegistrarInfo{
			Name: "Test Registrar",
		},
		Abuse: models.ContactInfo{
			Email: "abuse@testregistrar.com",
		},
		Hosting: &models.HostingInfo{
			Name: "Test Hosting",
			Abuse: models.ContactInfo{
				Email: "abuse@testhosting.com",
			},
		},
		CDN: &models.CDNInfo{
			Name: "Test CDN",
			Abuse: models.ContactInfo{
				Email: "abuse@testcdn.com",
			},
			Webform: "https://testcdn.com/abuse",
		},
	}

	tests := []struct {
		name       string
		actionDef  ActionDefinition
		shouldWork bool
	}{
		{
			name: "Registrar action",
			actionDef: ActionDefinition{
				Target: models.TakedownTarget{Type: "registrar"},
				Action: models.ActionSuspendDomain,
			},
			shouldWork: true,
		},
		{
			name: "Hosting action",
			actionDef: ActionDefinition{
				Target: models.TakedownTarget{Type: "hosting"},
				Action: models.ActionRemoveContent,
			},
			shouldWork: true,
		},
		{
			name: "CDN action",
			actionDef: ActionDefinition{
				Target: models.TakedownTarget{Type: "cdn"},
				Action: models.ActionRemoveContent,
			},
			shouldWork: true,
		},
		{
			name: "Search action",
			actionDef: ActionDefinition{
				Target: models.TakedownTarget{Type: "search"},
				Action: models.ActionWarningList,
			},
			shouldWork: true,
		},
		{
			name: "Unsupported action type",
			actionDef: ActionDefinition{
				Target: models.TakedownTarget{Type: "unknown"},
				Action: models.ActionBlocklist,
			},
			shouldWork: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.enrichAction(tt.actionDef, contacts)

			if tt.shouldWork {
				if result == nil {
					t.Errorf("Expected enriched action, got nil")
				} else {
					// Check that entity was populated
					if result.Target.Entity == "" {
						t.Errorf("Entity should be populated")
					}
				}
			} else {
				if result != nil {
					t.Errorf("Expected nil for unsupported action, got %v", result)
				}
			}
		})
	}
}

func TestEngine_PrioritizeActions(t *testing.T) {
	engine := NewEngine()

	actions := []ActionDefinition{
		{
			Target: models.TakedownTarget{Type: "blocklist"},
			Action: models.ActionBlocklist,
		},
		{
			Target: models.TakedownTarget{Type: "hosting"},
			Action: models.ActionRemoveContent,
		},
		{
			Target: models.TakedownTarget{Type: "registrar"},
			Action: models.ActionSuspendDomain,
		},
		{
			Target: models.TakedownTarget{Type: "search"},
			Action: models.ActionWarningList,
		},
		{
			Target: models.TakedownTarget{Type: "cdn"},
			Action: models.ActionRemoveContent,
		},
		// Duplicate hosting - should be deduplicated
		{
			Target: models.TakedownTarget{Type: "hosting"},
			Action: models.ActionRemoveContent,
		},
	}

	result := engine.prioritizeActions(actions)

	// Should deduplicate - no duplicate hosting
	if len(result) != 5 {
		t.Errorf("Expected 5 unique actions after deduplication, got %d", len(result))
	}

	// Check that all types are represented
	types := make(map[string]bool)
	for _, action := range result {
		types[action.Target.Type] = true
	}

	expectedTypes := []string{"hosting", "cdn", "registrar", "search", "blocklist"}
	for _, expectedType := range expectedTypes {
		if !types[expectedType] {
			t.Errorf("Missing action type: %s", expectedType)
		}
	}
}

func TestEngine_AddRule(t *testing.T) {
	engine := NewEngine()
	initialRuleCount := len(engine.GetRules())

	newRule := Rule{
		Match: []string{"test_category"},
		Actions: []ActionDefinition{
			{
				Target: models.TakedownTarget{Type: "test"},
				Action: models.ActionSuspendDomain,
			},
		},
	}

	engine.AddRule(newRule)

	newRuleCount := len(engine.GetRules())
	if newRuleCount != initialRuleCount+1 {
		t.Errorf("Expected %d rules after adding one, got %d", initialRuleCount+1, newRuleCount)
	}
}

func TestEngine_NoMatchingContacts(t *testing.T) {
	engine := NewEngine()

	// Contacts with no registrar or hosting
	contacts := &models.AbuseContact{
		Domain: "orphan-domain.com",
	}

	tags := []string{"phishing"}
	actions := engine.DetermineActions(tags, contacts)

	// Should still get search and blocklist actions even without registrar/hosting
	hasSearch := false
	hasBlocklist := false

	for _, action := range actions {
		switch action.Target.Type {
		case "search":
			hasSearch = true
		case "blocklist":
			hasBlocklist = true
		}
	}

	if !hasSearch {
		t.Errorf("Should have search action even without contacts")
	}
	if !hasBlocklist {
		t.Errorf("Should have blocklist action even without contacts")
	}
}

func TestEngine_LoadDefaultRules(t *testing.T) {
	engine := NewEngine()
	rules := engine.GetRules()

	if len(rules) == 0 {
		t.Errorf("Should have default rules loaded")
	}

	// Check that we have rules for major categories
	categories := []string{"phishing", "malware", "c2", "brand"}

	for _, category := range categories {
		found := false
		for _, rule := range rules {
			for _, match := range rule.Match {
				if match == category || (match == "brand:*" && category == "brand") {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			t.Errorf("Missing rule for category: %s", category)
		}
	}
}
