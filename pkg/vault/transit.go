package vault

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	vault "github.com/hashicorp/vault/api"
)

type transit struct {
	logical    *vault.Logical
	secretPath string
	secretKey  string
}

func (t *transit) Encrypt(ctx context.Context, request EncryptRequest) (*EncryptResponse, error) {

	path := fmt.Sprintf("%s/encrypt/%s", t.secretPath, t.secretKey)

	payload := map[string]any{
		"plaintext": base64.StdEncoding.EncodeToString(request.Plaintext),
	}
	if request.KeyVersion > 0 {
		payload["key_version"] = request.KeyVersion
	}
	if len(request.Context) > 0 {
		payload["context"] = base64.StdEncoding.EncodeToString(request.Context)
	}

	response, err := t.performRequest(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	result, err := t.parseEncryptResult(response.Data)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (t *transit) BatchEncrypt(ctx context.Context, request BatchEncryptRequest) ([]EncryptResponse, error) {

	if len(request.BatchInput) == 0 {
		return nil, ErrInvalidAttribute("items to be encrypted cannot be empty")
	}

	path := fmt.Sprintf("%s/encrypt/%s", t.secretPath, t.secretKey)

	payload := map[string]any{}
	if request.KeyVersion > 0 {
		payload["key_version"] = request.KeyVersion
	}

	batchInput := make([]map[string]string, len(request.BatchInput))
	for i, input := range request.BatchInput {
		batchInput[i] = map[string]string{
			"plaintext": base64.StdEncoding.EncodeToString(input.Plaintext),
		}
		if len(input.Context) > 0 {
			batchInput[i]["context"] = base64.StdEncoding.EncodeToString(input.Context)
		}
	}
	payload["batch_input"] = batchInput

	response, err := t.performRequest(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	batchResultsAny, ok := response.Data["batch_results"]
	if !ok {
		return nil, ErrMissingExpectedElement("batch_results")
	}

	batchResults, ok := batchResultsAny.([]any)
	if !ok {
		return nil, ErrInvalidDataType("batch_results")
	}

	results := make([]EncryptResponse, len(batchResults))

	for i, val := range batchResults {

		result, ok := val.(map[string]any)
		if !ok {
			return nil, ErrInvalidDataType(fmt.Sprintf("batch item %d: invalid data type", i+1))
		}
		if results[i], err = t.parseEncryptResult(result); err != nil {
			return nil, err
		}
	}
	return results, nil
}

func (t *transit) Decrypt(ctx context.Context, request DecryptRequest) (*DecryptResponse, error) {

	path := fmt.Sprintf("%s/decrypt/%s", t.secretPath, t.secretKey)

	payload := map[string]any{
		"ciphertext": request.Ciphertext,
	}
	if len(request.Context) > 0 {
		payload["context"] = base64.StdEncoding.EncodeToString(request.Context)
	}

	response, err := t.performRequest(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	plaintextBase64, ok := response.Data["plaintext"].(string)
	if !ok {
		return nil, ErrMissingExpectedElement("plaintext")
	}

	plaintext, err := base64.StdEncoding.DecodeString(plaintextBase64)
	if err != nil {
		return nil, fmt.Errorf("decode plaintext: %w", err)
	}
	return &DecryptResponse{Plaintext: plaintext}, nil
}

// Helper method that marshals payload to JSON and sends it to Vault logical backend via WriteBytesWithContext. Centralizes request execution logic.
func (t *transit) performRequest(ctx context.Context, path string, payload any) (*vault.Secret, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to perform json marshal: %w", err)
	}

	response, err := t.logical.WriteBytesWithContext(ctx, path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to %s: %w", path, err)

	} else if response == nil {
		return nil, fmt.Errorf("secret engine or secret key not found: %s", path)
	}
	return response, nil
}

// Parses Vault encryption response data to extract ciphertext and key version. Validates presence of required fields.
func (t *transit) parseEncryptResult(result map[string]any) (encrypted EncryptResponse, err error) {
	ciphertext, ok := result["ciphertext"].(string)
	if !ok {
		return encrypted, ErrMissingExpectedElement("ciphertext")
	}

	keyVersionAny, ok := result["key_version"]
	if !ok {
		return encrypted, ErrMissingExpectedElement("key_version")
	}

	keyVersion, _ := strconv.Atoi(fmt.Sprintf("%v", keyVersionAny))
	return EncryptResponse{
		Ciphertext: ciphertext, KeyVersion: uint(keyVersion),
	}, nil
}
