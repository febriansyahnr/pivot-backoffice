package vault

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	vault "github.com/hashicorp/vault/api"
)

type Config struct {
	// Address is the address of the Vault server. This should be a complete
	// URL such as "http://vault.example.com". If you need a custom SSL
	// cert or want to enable insecure mode, you need to specify a custom
	// HttpClient.
	Address string

	// HttpClient is the HTTP client to use. Vault sets sane defaults for the
	// http.Client and its associated http.Transport created in DefaultConfig.
	// If you must modify Vault's defaults, it is suggested that you start with
	// that client and modify as needed rather than start with an empty client
	// (or http.DefaultClient).
	HttpClient *http.Client

	// MinRetryWait controls the minimum time to wait before retrying when a 5xx
	// error occurs. Defaults to 1000 milliseconds.
	MinRetryWait time.Duration

	// MaxRetryWait controls the maximum time to wait before retrying when a 5xx
	// error occurs. Defaults to 1500 milliseconds.
	MaxRetryWait time.Duration

	// MaxRetries controls the maximum number of times to retry when a 5xx
	// error occurs. Set to -1 to disable retrying. Defaults to 2 (for a total
	// of three tries).
	MaxRetries int

	// Timeout, given a non-negative value, will apply the request timeout
	// to each request function unless an earlier deadline is passed to the
	// request function through context.Context. Note that this timeout is
	// not applicable to Logical().ReadRaw* (raw response) functions.
	// Defaults to 60 seconds.
	Timeout time.Duration

	// Token is the Vault authentication token used to authorize API requests.
	Token string
}

type Client struct {
	client *vault.Client
	cache  *sync.Map
}

func New(config Config) (*Client, error) {
	if config.Address == "" {
		return nil, errors.New("vault address is required")
	}
	if config.Token == "" {
		return nil, errors.New("vault token is required")
	}

	vaultConfig := vault.DefaultConfig()
	vaultConfig.Address = config.Address
	vaultConfig.MinRetryWait = config.MinRetryWait
	vaultConfig.MaxRetryWait = config.MaxRetryWait
	vaultConfig.HttpClient = config.HttpClient

	if config.MaxRetries < 0 {
		vaultConfig.MaxRetries = 0 // Disable retry

	} else if config.MaxRetries > 0 {
		vaultConfig.MaxRetries = config.MaxRetries
	}

	client, err := vault.NewClient(vaultConfig)
	if err != nil {
		return nil, fmt.Errorf("init vault client: %w", err)
	}
	client.SetToken(config.Token)

	return &Client{client: client, cache: new(sync.Map)}, nil
}

func (c *Client) NewKeyValue(mountPath, secretPath string) IVaultKeyValue {
	return &keyValue{
		kv:          c.client.KVv2(mountPath),
		mountPath:   mountPath,
		secretPath:  secretPath,
		sharedCache: c.cache,
	}
}

func (c *Client) NewTransit(secretPath, secretKey string) IVaultTransit {
	return &transit{
		logical:    c.client.Logical(),
		secretPath: secretPath,
		secretKey:  secretKey,
	}
}
