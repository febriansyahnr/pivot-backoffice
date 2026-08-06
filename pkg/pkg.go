package pkg

import (
	"github.com/paper-indonesia/pdk/v2/encrypt"
	"github.com/paper-indonesia/pdk/v2/gcp"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// The interface used in the PDK package to create aliases and generate mocks for Unit Test purposes.
type (
	ILogger           interface{ logger.ILogger }
	IGCPSecretManager interface{ gcp.ISecretManager }
	Encrypter         interface{ encrypt.Encrypter }
	// Etc ...
)
