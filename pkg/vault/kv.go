package vault

import (
	"context"
	"errors"
	"fmt"
	"sync"

	vault "github.com/hashicorp/vault/api"
)

type keyValue struct {
	kv          *vault.KVv2
	mountPath   string
	secretPath  string
	sharedCache *sync.Map
}

type Secret struct {
	Version int
	Data    map[string]any
}

type SecretKey[T comparable] struct {
	Version int
	Value   T
}

var ErrKeyNotFound = errors.New("secret key not found")

func (k *keyValue) GetSecret(ctx context.Context) (*Secret, error) {
	secret, err := k.kv.Get(ctx, k.secretPath)
	if err != nil {
		return nil, err
	}
	return &Secret{
		Version: secret.VersionMetadata.Version,
		Data:    secret.Data,
	}, nil
}

func (k *keyValue) GetSecretVersion(ctx context.Context, version int) (*Secret, error) {

	dataKey := k.secretDataPath()
	if cache, ok := k.sharedCache.Load(dataKey); ok {
		return cache.(*Secret), nil
	}

	result, err := k.kv.GetVersion(ctx, k.secretPath, version)
	if err != nil {
		return nil, err
	}

	secret := &Secret{
		Version: result.VersionMetadata.Version,
		Data:    result.Data,
	}
	k.sharedCache.Store(dataKey, secret)

	return secret, nil
}

func (k *keyValue) GetSecretKeyString(ctx context.Context, key string) (*SecretKey[string], error) {
	secret, err := k.GetSecret(ctx)
	if err != nil {
		return nil, err
	}
	return k.secretKeyString(secret, key)
}

func (k *keyValue) GetSecretKeyVersionString(ctx context.Context, version int, key string) (*SecretKey[string], error) {
	secret, err := k.GetSecretVersion(ctx, version)
	if err != nil {
		return nil, err
	}
	return k.secretKeyString(secret, key)
}

func (k *keyValue) secretDataPath() string {
	return fmt.Sprintf("%s/data/%s", k.mountPath, k.secretPath)
}

func (keyValue) secretKeyString(secret *Secret, key string) (*SecretKey[string], error) {
	value, ok := secret.Data[key].(string)
	if !ok {
		return nil, fmt.Errorf("key %s: %w", key, ErrKeyNotFound)
	}
	return &SecretKey[string]{
		Version: secret.Version, Value: value,
	}, nil
}
