package evidence

import (
	"github.com/cti-team/takedown/pkg/models"
)

// Collector is a stub implementation used for building and testing.
// TODO: implement evidence collection logic.
type Collector struct{}

// NewCollector returns a new Collector instance.
func NewCollector() *Collector {
	return &Collector{}
}

// CollectEvidence returns an empty EvidencePack placeholder.
func (c *Collector) CollectEvidence(ioc *models.IOC) (*models.EvidencePack, error) {
	return &models.EvidencePack{EvidenceID: "stub"}, nil
}
