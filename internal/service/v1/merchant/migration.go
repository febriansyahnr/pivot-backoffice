package merchant

import (
	"context"
	"fmt"
	"sync"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/panjf2000/ants/v2"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MerchantService) MigrateMerchantSecretsToEncryption(ctx context.Context) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/MigrateMerchantSecretsToEncryption")
	defer segment.End()

	merchants, err := s.repo.ListUnencryptedMerchantSecretsForMigration(ctx)
	if err != nil {
		return err

	} else if len(merchants) == 0 {
		s.logger.Info(ctx, "No data will be migrated, the process is stopped")
		return nil
	}

	wg := new(sync.WaitGroup)

	s.logger.Info(ctx, "Starting the merchant secret encryption process", logger.Int("total", len(merchants)))

	pool, err := ants.NewPoolWithFuncGeneric(25, func(data merchant.UnencryptedMerchantSecretsForMigration) {
		defer wg.Done()

		err := error(nil)
		defer func() {
			s.logger.Info(ctx, "Secret encryption process for merchant ID: "+data.MerchantId, logger.Bool("completed", err == nil), logger.Error(err))
		}()

		wrappedSecrets, err := s.encryption.BatchEncrypt(ctx, vault.BatchEncryptRequest{
			BatchInput: []vault.BatchEncryptInput{
				{Plaintext: []byte(data.CallbackApiKey)}, // Mapped Index 0
				{Plaintext: []byte(data.JITApiKey)},      // Mapped Index 1
				{Plaintext: []byte(data.Secret)},         // Mapped Index 2
			},
		})
		if err != nil {
			return

		} else if len(wrappedSecrets) != 3 {
			err = fmt.Errorf("encryption results do not match, result %d should be 3", len(wrappedSecrets))
			return
		}

		keyVersion := wrappedSecrets[0].KeyVersion

		request := merchant.MigrateMerchantSecretsToEncryption{
			MerchantId:            data.MerchantId,
			CallbackApiKey:        wrappedSecrets[0].Ciphertext,
			Secret:                wrappedSecrets[2].Ciphertext,
			CallbackApiKeyVersion: keyVersion,
			SecretVersion:         keyVersion,
		}
		if data.JITApiKey != "" {
			request.JITApiKey = wrappedSecrets[1].Ciphertext
			request.JITApiKeyVersion = keyVersion
		}
		err = s.repo.MigrateMerchantSecretsToEncryption(ctx, request)
	})
	if err != nil {
		return fmt.Errorf("failed to set up the worker pool, %v", err)
	}
	defer pool.Release()

	for _, merchant := range merchants {
		wg.Add(1)

		_ = pool.Invoke(merchant)
	}

	wg.Wait()

	return nil
}
