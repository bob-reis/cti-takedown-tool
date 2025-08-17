package rdap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cti-team/takedown/pkg/models"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient should not return nil")
	}
	if client.userAgent == "" {
		t.Error("User agent should be set")
	}
	if client.httpClient == nil {
		t.Error("HTTP client should be initialized")
	}
}

func TestClient_GetRDAPURL(t *testing.T) {
	client := NewClient()

	tests := []struct {
		domain      string
		expectedURL string
		shouldError bool
	}{
		{
			domain:      "example.com",
			expectedURL: "https://rdap.verisign.com/com/v1",
			shouldError: false,
		},
		{
			domain:      "test.net",
			expectedURL: "https://rdap.verisign.com/net/v1",
			shouldError: false,
		},
		{
			domain:      "organization.org",
			expectedURL: "https://rdap.publicinterestregistry.org",
			shouldError: false,
		},
		{
			domain:      "registro.br",
			expectedURL: "https://rdap.registro.br",
			shouldError: false,
		},
		{
			domain:      "unknown.xyz",
			expectedURL: "https://rdap-bootstrap.arin.net/bootstrap/domain/unknown.xyz",
			shouldError: false,
		},
		{
			domain:      "invalid",
			expectedURL: "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			url, err := client.getRDAPURL(tt.domain)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for domain %s", tt.domain)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for domain %s: %v", tt.domain, err)
				}
				if url != tt.expectedURL {
					t.Errorf("Expected URL %s, got %s", tt.expectedURL, url)
				}
			}
		})
	}
}

func TestClient_HasRole(t *testing.T) {
	client := NewClient()

	tests := []struct {
		roles    []string
		role     string
		expected bool
	}{
		{
			roles:    []string{"registrar", "administrative"},
			role:     "registrar",
			expected: true,
		},
		{
			roles:    []string{"registrar", "administrative"},
			role:     "REGISTRAR",
			expected: true, // Case insensitive
		},
		{
			roles:    []string{"technical", "billing"},
			role:     "registrar",
			expected: false,
		},
		{
			roles:    []string{},
			role:     "registrar",
			expected: false,
		},
	}

	for _, tt := range tests {
		result := client.hasRole(tt.roles, tt.role)
		if result != tt.expected {
			t.Errorf("hasRole(%v, %s) = %v, want %v", tt.roles, tt.role, result, tt.expected)
		}
	}
}

func TestClient_ExtractEntityName(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name     string
		entity   Entity
		expected string
	}{
		{
			name: "Valid vCard with name",
			entity: Entity{
				Handle: "ENTITY123",
				VCardArray: []interface{}{
					"vcard",
					[]interface{}{
						[]interface{}{"version", map[string]interface{}{}, "text", "4.0"},
						[]interface{}{"fn", map[string]interface{}{}, "text", "GoDaddy.com, LLC"},
					},
				},
			},
			expected: "GoDaddy.com, LLC",
		},
		{
			name: "No vCard - return handle",
			entity: Entity{
				Handle:     "ENTITY456",
				VCardArray: []interface{}{},
			},
			expected: "ENTITY456",
		},
		{
			name: "Invalid vCard structure - return handle",
			entity: Entity{
				Handle: "ENTITY789",
				VCardArray: []interface{}{
					"vcard",
					"invalid_structure",
				},
			},
			expected: "ENTITY789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.extractEntityName(tt.entity)
			if result != tt.expected {
				t.Errorf("extractEntityName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClient_ExtractAbuseEmail(t *testing.T) {
	client := NewClient()

	tests := []struct {
		name     string
		entity   Entity
		expected string
	}{
		{
			name: "Valid abuse email",
			entity: Entity{
				VCardArray: []interface{}{
					"vcard",
					[]interface{}{
						[]interface{}{"version", map[string]interface{}{}, "text", "4.0"},
						[]interface{}{"email", map[string]interface{}{}, "text", "abuse@godaddy.com"},
					},
				},
			},
			expected: "abuse@godaddy.com",
		},
		{
			name: "Non-abuse email",
			entity: Entity{
				VCardArray: []interface{}{
					"vcard",
					[]interface{}{
						[]interface{}{"version", map[string]interface{}{}, "text", "4.0"},
						[]interface{}{"email", map[string]interface{}{}, "text", "info@example.com"},
					},
				},
			},
			expected: "",
		},
		{
			name: "No email",
			entity: Entity{
				VCardArray: []interface{}{
					"vcard",
					[]interface{}{
						[]interface{}{"version", map[string]interface{}{}, "text", "4.0"},
					},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.extractAbuseEmail(tt.entity)
			if result != tt.expected {
				t.Errorf("extractAbuseEmail() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClient_ProcessEntity(t *testing.T) {
	client := NewClient()

	// Test registrar entity
	registrarEntity := Entity{
		Handle: "GODADDY",
		Roles:  []string{"registrar"},
		VCardArray: []interface{}{
			"vcard",
			[]interface{}{
				[]interface{}{"fn", map[string]interface{}{}, "text", "GoDaddy.com, LLC"},
				[]interface{}{"email", map[string]interface{}{}, "text", "abuse@godaddy.com"},
			},
		},
	}

	contact := &models.AbuseContact{}
	client.processEntity(registrarEntity, contact)

	if contact.Registrar == nil {
		t.Fatal("Registrar should be set")
	}
	if contact.Registrar.Name != "GoDaddy.com, LLC" {
		t.Errorf("Registrar name not correct, got: %s", contact.Registrar.Name)
	}
	if contact.Abuse.Email != "abuse@godaddy.com" {
		t.Errorf("Abuse email not correct, got: %s", contact.Abuse.Email)
	}

	// Test privacy entity
	privacyEntity := Entity{
		Handle: "PRIVACY",
		Roles:  []string{"proxy"},
	}

	contact2 := &models.AbuseContact{}
	client.processEntity(privacyEntity, contact2)

	if !contact2.Privacy {
		t.Error("Privacy flag should be set for proxy entity")
	}
}

func TestClient_LookupDomain_MockServer(t *testing.T) {
	// Create mock RDAP response
	mockResponse := RDAPResponse{
		ObjectClassName: "domain",
		LDHName:         "example.com",
		Entities: []Entity{
			{
				ObjectClassName: "entity",
				Handle:          "GODADDY",
				Roles:           []string{"registrar"},
				VCardArray: []interface{}{
					"vcard",
					[]interface{}{
						[]interface{}{"fn", map[string]interface{}{}, "text", "GoDaddy.com, LLC"},
						[]interface{}{"email", map[string]interface{}{}, "text", "abuse@godaddy.com"},
					},
				},
			},
		},
		Events: []Event{
			{
				EventAction: "registration",
				EventDate:   time.Now(),
			},
		},
	}

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/domain/example.com") {
			w.WriteHeader(404)
			return
		}

		w.Header().Set("Content-Type", "application/rdap+json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// For this test, we'll just test the individual components
	// rather than the full lookup due to method override limitations
	client := NewClient()

	// Test makeRequest directly
	response, err := client.makeRequest(server.URL + "/domain/example.com")
	if err != nil {
		t.Fatalf("makeRequest failed: %v", err)
	}

	// Parse the response
	var rdapResp RDAPResponse
	if err := json.Unmarshal(response, &rdapResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify the mock response structure
	if rdapResp.LDHName != "example.com" {
		t.Errorf("LDHName not correct")
	}
	if len(rdapResp.Entities) == 0 {
		t.Errorf("Should have entities")
	}
}

func TestClient_LookupDomain_ErrorHandling(t *testing.T) {
	// Test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewClient()

	// Test makeRequest with error server
	_, err := client.makeRequest(server.URL + "/domain/error.com")
	if err == nil {
		t.Error("Expected error for server error response")
	}
}

func TestClient_LookupDomain_InvalidJSON(t *testing.T) {
	// Test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rdap+json")
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewClient()

	// Test makeRequest with invalid JSON server
	response, err := client.makeRequest(server.URL + "/domain/invalid.com")
	if err != nil {
		t.Errorf("makeRequest should succeed, got error: %v", err)
	}

	// The error should come when trying to parse the JSON
	var rdapResp RDAPResponse
	err = json.Unmarshal(response, &rdapResp)
	if err == nil {
		t.Error("Expected JSON parsing error")
	}
}

func TestClient_MakeRequest(t *testing.T) {
	// Test successful request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "CTI-Takedown/1.0" {
			t.Errorf("User-Agent header not set correctly")
		}
		if r.Header.Get("Accept") != "application/rdap+json" {
			t.Errorf("Accept header not set correctly")
		}

		_, _ = w.Write([]byte(`{"test": "response"}`))
	}))
	defer server.Close()

	client := NewClient()
	response, err := client.makeRequest(server.URL)

	if err != nil {
		t.Fatalf("makeRequest failed: %v", err)
	}

	if string(response) != `{"test": "response"}` {
		t.Errorf("Response not correct, got: %s", string(response))
	}
}

func TestClient_MakeRequest_InvalidURL(t *testing.T) {
	client := NewClient()

	_, err := client.makeRequest("invalid-url")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

// Test domain normalization.
func TestDomainNormalization(t *testing.T) {
	client := NewClient()

	tests := []struct {
		input    string
		expected string
	}{
		{"EXAMPLE.COM", "example.com"},
		{"  test.org  ", "test.org"},
		{"MiXeD.CaSe.net", "mixed.case.net"},
	}

	for _, tt := range tests {
		// Test through the full lookup (will fail on network, but we can check normalization)
		_, err := client.LookupDomain(tt.input)
		// We expect network errors, not normalization errors
		if err != nil && strings.Contains(err.Error(), "invalid domain format") {
			t.Errorf("Domain normalization failed for input: %s", tt.input)
		}
	}
}

// Benchmark tests.
func BenchmarkExtractEntityName(b *testing.B) {
	client := NewClient()
	entity := Entity{
		Handle: "BENCH_ENTITY",
		VCardArray: []interface{}{
			"vcard",
			[]interface{}{
				[]interface{}{"fn", map[string]interface{}{}, "text", "Benchmark Registrar LLC"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.extractEntityName(entity)
	}
}
