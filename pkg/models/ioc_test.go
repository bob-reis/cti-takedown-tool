package models

import (
	"testing"
	"time"
)

func TestIOC_GetBrand(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{
			name:     "Extract brand from tag",
			tags:     []string{"phishing", "brand:AcmeBank", "high"},
			expected: "AcmeBank",
		},
		{
			name:     "No brand tag",
			tags:     []string{"phishing", "high"},
			expected: "",
		},
		{
			name:     "Multiple brand tags - first one",
			tags:     []string{"brand:Bank1", "brand:Bank2", "phishing"},
			expected: "Bank1",
		},
		{
			name:     "Empty tags",
			tags:     []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ioc := &IOC{Tags: tt.tags}
			result := ioc.GetBrand()
			if result != tt.expected {
				t.Errorf("GetBrand() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIOC_HasTag(t *testing.T) {
	ioc := &IOC{
		Tags: []string{"phishing", "brand:MyBank", "high"},
	}

	tests := []struct {
		tag      string
		expected bool
	}{
		{"phishing", true},
		{"brand:MyBank", true},
		{"high", true},
		{"malware", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			result := ioc.HasTag(tt.tag)
			if result != tt.expected {
				t.Errorf("HasTag(%s) = %v, want %v", tt.tag, result, tt.expected)
			}
		})
	}
}

func TestIOC_GetSeverity(t *testing.T) {
	tests := []struct {
		name     string
		tags     []string
		expected string
	}{
		{
			name:     "Critical severity",
			tags:     []string{"phishing", "critical"},
			expected: "critical",
		},
		{
			name:     "High severity",
			tags:     []string{"phishing", "high"},
			expected: "high",
		},
		{
			name:     "Medium severity",
			tags:     []string{"phishing", "medium"},
			expected: "medium",
		},
		{
			name:     "Low severity",
			tags:     []string{"phishing", "low"},
			expected: "low",
		},
		{
			name:     "No severity tag - default medium",
			tags:     []string{"phishing", "brand:Bank"},
			expected: "medium",
		},
		{
			name:     "Multiple severity tags - first priority (critical)",
			tags:     []string{"phishing", "low", "critical", "high"},
			expected: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ioc := &IOC{Tags: tt.tags}
			result := ioc.GetSeverity()
			if result != tt.expected {
				t.Errorf("GetSeverity() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIOCType_Constants(t *testing.T) {
	// Test that IOC type constants are correct
	if IOCTypeURL != "url" {
		t.Errorf("IOCTypeURL should be 'url', got %v", IOCTypeURL)
	}
	if IOCTypeDomain != "domain" {
		t.Errorf("IOCTypeDomain should be 'domain', got %v", IOCTypeDomain)
	}
	if IOCTypeIP != "ip" {
		t.Errorf("IOCTypeIP should be 'ip', got %v", IOCTypeIP)
	}
	if IOCTypeHash != "hash" {
		t.Errorf("IOCTypeHash should be 'hash', got %v", IOCTypeHash)
	}
}

func TestIOC_Creation(t *testing.T) {
	now := time.Now()
	ioc := &IOC{
		IndicatorID: "test-ioc-1",
		Type:        IOCTypeURL,
		Value:       "https://malicious-site.com",
		FirstSeen:   now,
		Source:      "test",
		Tags:        []string{"phishing", "brand:TestBank"},
	}

	if ioc.IndicatorID != "test-ioc-1" {
		t.Errorf("IndicatorID not set correctly")
	}
	if ioc.Type != IOCTypeURL {
		t.Errorf("Type not set correctly")
	}
	if ioc.Value != "https://malicious-site.com" {
		t.Errorf("Value not set correctly")
	}
	if !ioc.HasTag("phishing") {
		t.Errorf("Should have phishing tag")
	}
	if ioc.GetBrand() != "TestBank" {
		t.Errorf("Should extract brand correctly")
	}
}
