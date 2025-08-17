package evidence

import (
	"github.com/cti-team/takedown/pkg/models"
	"testing"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	if c == nil {
		t.Fatalf("NewCollector returned nil")
	}
}

func TestCollectorCollectEvidence(t *testing.T) {
	c := NewCollector()
	ioc := &models.IOC{IndicatorID: "test"}
	pack, err := c.CollectEvidence(ioc)
	if err != nil {
		t.Fatalf("CollectEvidence returned error: %v", err)
	}
	if pack == nil {
		t.Fatalf("CollectEvidence returned nil")
	}
	if pack.EvidenceID != "stub" {
		t.Fatalf("unexpected EvidenceID: %s", pack.EvidenceID)
	}
}
