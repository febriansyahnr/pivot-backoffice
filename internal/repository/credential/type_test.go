package credential_test

import "github.com/stretchr/testify/mock"

var (
	timeType                    = mock.AnythingOfType("time.Time")
	clientSecretSumPtrSliceType = mock.AnythingOfType("*[]credential.ClientSecretSummary")
	clientSecretPtrType         = mock.AnythingOfType("*credential.ClientSecret")
)
