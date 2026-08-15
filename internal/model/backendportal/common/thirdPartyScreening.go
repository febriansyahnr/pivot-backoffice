package commonModel

import (
	amlcommon "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/amlProcessor"
	dukcapilmodel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/dukcapil"
)

// ThirdPartyScreeningData represents unified storage structure for all third-party screening providers
// Uses name:dob as key pattern for both providers to enable per-person screening data storage
type ThirdPartyScreeningData struct {
	AML      map[string]*amlcommon.ScreeningResponse         `json:"aml,omitempty"`      // AML screening data keyed by name:dob
	Dukcapil map[string]*dukcapilmodel.DukcapilScreeningData `json:"dukcapil,omitempty"` // Dukcapil screening data keyed by name:dob
}
