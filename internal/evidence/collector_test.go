package evidence

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cti-team/takedown/pkg/models"
)

func TestNewCollector(t *testing.T) {
	outputDir := "/tmp/test-evidence"
	collector := NewCollector(outputDir)

	if collector == nil {
		t.Fatal("NewCollector should not return nil")
	}
	if collector.outputDir != outputDir {
		t.Errorf("Output directory not set correctly")
	}
	if collector.timeout != 30*time.Second {
		t.Errorf("Timeout not set correctly")
	}
	if collector.userAgent == "" {
		t.Errorf("User agent should be set")
	}
}

func TestCollector_DefangIOC(t *testing.T) {
	collector := NewCollector("/tmp")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTTP URL",
			input:    "http://malicious.com/path",
			expected: "hxxp://malicious[.]com/path",
		},
		{
			name:     "HTTPS URL",
			input:    "https://phishing.org/login",
			expected: "hxxps://phishing[.]org/login",
		},
		{
			name:     "Domain only",
			input:    "suspicious.net",
			expected: "suspicious[.]net",
		},
		{
			name:     "Email address",
			input:    "admin@evil.com",
			expected: "admin[@]evil[.]com",
		},
		{
			name:     "Multiple dots",
			input:    "sub.domain.evil.com",
			expected: "sub[.]domain[.]evil[.]com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.defangIOC(tt.input)
			if result != tt.expected {
				t.Errorf("defangIOC(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCollector_ExtractTitle(t *testing.T) {
	collector := NewCollector("/tmp")

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Simple title",
			html:     "<html><head><title>Test Page</title></head><body></body></html>",
			expected: "Test Page",
		},
		{
			name:     "Title with whitespace",
			html:     "<html><head><title>  Login Page  </title></head></html>",
			expected: "Login Page",
		},
		{
			name:     "No title tag",
			html:     "<html><head></head><body>No title</body></html>",
			expected: "",
		},
		{
			name:     "Long title - truncated",
			html:     "<html><head><title>" + strings.Repeat("A", 150) + "</title></head></html>",
			expected: strings.Repeat("A", 100) + "...",
		},
		{
			name:     "Case insensitive",
			html:     "<HTML><HEAD><TITLE>Upper Case</TITLE></HEAD></HTML>",
			expected: "Upper Case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collector.extractTitle(tt.html)
			if result != tt.expected {
				t.Errorf("extractTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCollector_AssessRisk(t *testing.T) {
	collector := NewCollector("/tmp")

	tests := []struct {
		name     string
		ioc      *models.IOC
		evidence *models.EvidencePack
		minScore int
		maxScore int
		category string
	}{
		{
			name: "Phishing site with login forms",
			ioc: &models.IOC{
				Tags: []string{"phishing", "brand:TestBank"},
			},
			evidence: &models.EvidencePack{
				HTTP: models.HTTPInfo{
					Status: 200,
					Body:   "Please enter your login and password to continue",
					Headers: map[string]string{
						"TLS-Issuer": "Let's Encrypt Authority X3",
					},
				},
			},
			minScore: 60,
			maxScore: 100,
			category: "phishing",
		},
		{
			name: "Malware distribution",
			ioc: &models.IOC{
				Tags: []string{"malware", "high"},
			},
			evidence: &models.EvidencePack{
				HTTP: models.HTTPInfo{
					Status: 200,
				},
			},
			minScore: 40,
			maxScore: 100,
			category: "malware",
		},
		{
			name: "Site down - lower risk",
			ioc: &models.IOC{
				Tags: []string{"phishing"},
			},
			evidence: &models.EvidencePack{
				HTTP: models.HTTPInfo{
					Status: 0,
				},
				DNS: models.DNSRecord{
					A: []string{}, // No A records
				},
			},
			minScore: 0,
			maxScore: 40,
			category: "phishing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risk := collector.assessRisk(tt.ioc, tt.evidence)

			if risk.Score < tt.minScore || risk.Score > tt.maxScore {
				t.Errorf("Risk score %d not in expected range [%d, %d]", 
					risk.Score, tt.minScore, tt.maxScore)
			}
			if risk.Category != tt.category {
				t.Errorf("Expected category %s, got %s", tt.category, risk.Category)
			}
			if risk.Rationale == "" {
				t.Errorf("Risk rationale should not be empty")
			}
		})
	}
}

func TestCollector_CollectHTTP_MockServer(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "nginx/1.18.0")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(200)
		
		html := `<html>
			<head><title>Fake Banking Login</title></head>
			<body>
				<form>
					<input type="text" name="username" placeholder="Login">
					<input type="password" name="password" placeholder="Password">
				</form>
			</body>
		</html>`
		
		w.Write([]byte(html))
	}))
	defer server.Close()

	collector := NewCollector("/tmp")
	ctx := context.Background()

	httpInfo, err := collector.collectHTTP(ctx, server.URL, "test-evidence")
	if err != nil {
		t.Fatalf("collectHTTP failed: %v", err)
	}

	// Check HTTP status
	if httpInfo.Status != 200 {
		t.Errorf("Expected status 200, got %d", httpInfo.Status)
	}

	// Check headers
	if httpInfo.Headers["Server"] != "nginx/1.18.0" {
		t.Errorf("Server header not captured correctly")
	}
	if httpInfo.Headers["Content-Type"] != "text/html" {
		t.Errorf("Content-Type header not captured correctly")
	}

	// Check title extraction
	if httpInfo.Title != "Fake Banking Login" {
		t.Errorf("Title not extracted correctly, got: %s", httpInfo.Title)
	}

	// Check body content
	if !strings.Contains(httpInfo.Body, "username") {
		t.Errorf("Body should contain 'username'")
	}

	// Check TLS info (from test server)
	if httpInfo.Headers["TLS-Issuer"] == "" {
		t.Errorf("TLS information should be captured")
	}
}

func TestCollector_CollectHTTP_ErrorHandling(t *testing.T) {
	collector := NewCollector("/tmp")
	ctx := context.Background()

	// Test with invalid URL
	_, err := collector.collectHTTP(ctx, "invalid-url", "test-evidence")
	if err == nil {
		t.Errorf("Expected error for invalid URL")
	}

	// Test with non-existent domain
	_, err = collector.collectHTTP(ctx, "https://this-domain-should-not-exist-12345.com", "test-evidence")
	if err == nil {
		t.Errorf("Expected error for non-existent domain")
	}
}

func TestCollector_CollectEvidence_URL(t *testing.T) {
	// Skip if we can't create temp directory
	tempDir := "/tmp/test-evidence-" + fmt.Sprintf("%d", time.Now().Unix())
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Skipf("Cannot create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	collector := NewCollector(tempDir)

	ioc := &models.IOC{
		IndicatorID: "test-ioc-123",
		Type:        models.IOCTypeURL,
		Value:       "https://example.com/test", // Use a real domain that should resolve
		Tags:        []string{"test", "phishing"},
	}

	evidence, err := collector.CollectEvidence(ioc)
	if err != nil {
		// Network errors are acceptable in tests
		if strings.Contains(err.Error(), "DNS collection failed") || 
		   strings.Contains(err.Error(), "no such host") {
			t.Skipf("Network error during test: %v", err)
		}
		t.Fatalf("CollectEvidence failed: %v", err)
	}

	// Check evidence structure
	if evidence.EvidenceID == "" {
		t.Errorf("Evidence ID should be set")
	}
	if evidence.IOC != ioc.IndicatorID {
		t.Errorf("IOC reference not correct")
	}
	if evidence.CollectedAt.IsZero() {
		t.Errorf("Collection timestamp should be set")
	}
	if evidence.Defanged == "" {
		t.Errorf("Defanged IOC should be set")
	}

	// Check risk assessment
	if evidence.Risk.Score < 0 || evidence.Risk.Score > 100 {
		t.Errorf("Risk score should be between 0-100, got %d", evidence.Risk.Score)
	}
	if evidence.Risk.Category == "" {
		t.Errorf("Risk category should be set")
	}
}

func TestCollector_CollectEvidence_Domain(t *testing.T) {
	tempDir := "/tmp/test-evidence-domain-" + fmt.Sprintf("%d", time.Now().Unix())
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		t.Skipf("Cannot create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	collector := NewCollector(tempDir)

	ioc := &models.IOC{
		IndicatorID: "test-domain-123",
		Type:        models.IOCTypeDomain,
		Value:       "example.com", // Use a real domain
		Tags:        []string{"test", "malware"},
	}

	evidence, err := collector.CollectEvidence(ioc)
	if err != nil {
		// Network errors are acceptable in tests
		if strings.Contains(err.Error(), "DNS collection failed") || 
		   strings.Contains(err.Error(), "no such host") {
			t.Skipf("Network error during test: %v", err)
		}
		t.Fatalf("CollectEvidence failed: %v", err)
	}

	// Verify domain-specific behavior
	if evidence.Defanged != "example[.]com" {
		t.Errorf("Domain defanging not correct, got: %s", evidence.Defanged)
	}
}

func TestCollector_CollectEvidence_UnsupportedType(t *testing.T) {
	collector := NewCollector("/tmp")

	ioc := &models.IOC{
		IndicatorID: "test-hash-123",
		Type:        models.IOCTypeHash,
		Value:       "d41d8cd98f00b204e9800998ecf8427e",
		Tags:        []string{"test"},
	}

	_, err := collector.CollectEvidence(ioc)
	if err == nil {
		t.Errorf("Expected error for unsupported IOC type")
	}
	if !strings.Contains(err.Error(), "unsupported IOC type") {
		t.Errorf("Error message should mention unsupported type")
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{5, 3, 3},
		{2, 7, 2},
		{4, 4, 4},
		{0, 1, 0},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// Benchmark tests
func BenchmarkDefangIOC(b *testing.B) {
	collector := NewCollector("/tmp")
	url := "https://very-long-malicious-domain-name-for-testing.evil.com/very/long/path/to/malicious/content"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.defangIOC(url)
	}
}

func BenchmarkExtractTitle(b *testing.B) {
	collector := NewCollector("/tmp")
	html := `<html><head><title>Test Title for Benchmarking</title></head><body>` +
		strings.Repeat("<p>Content</p>", 100) + `</body></html>`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.extractTitle(html)
	}
}