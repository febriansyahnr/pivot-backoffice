package merchant

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/account"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/industry"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func (s *MerchantService) Create(ctx context.Context, merchant *merchantModel.Merchant, userId string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/Create")
	defer segment.End()

	var (
		user    *userModel.User
		errFind error
	)

	if userId != "" {
		// check if user is exist
		user, errFind = s.User.FindUserByID(ctx, userId)
		if errFind != nil {
			s.logger.Error(ctx, "error when find user by id", logger.Error(errFind))
			return errors.New(responseHttp.HttpErrInternal, errFind)

		} else if user == nil {
			return errors.New(responseHttp.HttpErrNotFound, fmt.Errorf("user not found"))

		} else if user.MerchantId != "" {
			return errors.New(responseHttp.HttpErrRequest, fmt.Errorf("user already have merchant"))
		}
	}

	if loc, err := s.locationRepo.GetDistrictById(ctx, merchant.DistrictId); err != nil {
		s.logger.Error(ctx, "error when get district by id", logger.Error(err))
		return errors.New(responseHttp.HttpErrInternal, err)

	} else if loc == nil {
		return errors.New(responseHttp.HttpErrUnprocessableContent, constant.ErrDistrictNotFound)
	}

	// Validate parent and child industry combination
	parentProvided := merchant.ParentIndustry.Valid
	childProvided := merchant.ChildIndustry.Valid

	// Return early: if only one field is provided (XOR logic)
	if parentProvided != childProvided {
		return errors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("parent industry and child industry must be provided together"))
	}

	// Return early: if both provided but invalid combination
	if parentProvided && childProvided {
		if valid, err := s.industrySvc.ValidateIndustry(ctx, merchant.ParentIndustry.String, merchant.ChildIndustry.String); err != nil {
			return errors.New(responseHttp.HttpErrInternal, err)
		} else if !valid {
			return errors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("invalid parent and child industry combination"))
		}
	}

	// Return early: if MCC is invalid
	if merchant.MCC.Valid {
		if valid, err := s.industrySvc.IsValidMCC(ctx, merchant.MCC.String); err != nil {
			return errors.New(responseHttp.HttpErrInternal, err)
		} else if !valid {
			return errors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("invalid MCC code"))
		}
	}

	// Return early: if MCC doesn't match industry combination
	if merchant.ParentIndustry.Valid && merchant.ChildIndustry.Valid && merchant.MCC.Valid {
		if err := s.industrySvc.ValidateIndustryMCCCombination(ctx, merchant.ParentIndustry.String, merchant.ChildIndustry.String, merchant.MCC.String); err != nil {
			return errors.New(responseHttp.HttpErrUnprocessableContent, err)
		}
	}

	// Return early: if Digital Status is invalid
	if merchant.DigitalStatus.Valid && !industry.IsValidDigitalStatus(merchant.DigitalStatus.String) {
		return errors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("digital status must be 'Digital' or 'Non-digital'"))
	}

	// Return early: if Country of Entity is invalid
	if merchant.CountryOfEntity.Valid && merchant.CountryOfEntity.String != "" {
		country, err := s.countrySvc.FindByCode(ctx, merchant.CountryOfEntity.String)
		if err != nil {
			return err
		}
		if country == nil {
			return errors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("invalid country code for country of entity"))
		}
		merchant.CountryOfEntity = sql.NullString{String: country.Code, Valid: true}
	}

	if merchant.RiskLevel.Valid && !constant.IsValidRiskLevel(merchant.RiskLevel.String) {
		return errors.New(responseHttp.HttpErrUnprocessableContent, fmt.Errorf("invalid risk level: must be one of %v", constant.ValidMerchantRiskLevels))
	}

	// generate MID
	mid, err := s.repo.GenerateNewMID(ctx)
	if err != nil {
		s.logger.Error(ctx, "error when generate MID", logger.Error(err))
		return errors.New(responseHttp.HttpErrInternal, err)
	}
	// assign MID on success generate
	merchant.MID = sql.NullString{String: mid, Valid: true}

	// Generate callback api key
	callbackApiKey, _ := util.GenerateRandomString(32)

	// Generate JIT api key
	jitApiKey, _ := util.GenerateRandomString(32)

	// Merchant secret (B2B credential)
	merchantSecret, _ := util.GenerateRandomString(40)

	// Wrapped API Key
	wrappedSecrets, err := s.encryption.BatchEncrypt(ctx, vault.BatchEncryptRequest{
		BatchInput: []vault.BatchEncryptInput{
			{Plaintext: []byte(callbackApiKey)}, // Mapped Index 0
			{Plaintext: []byte(jitApiKey)},      // Mapped Index 1
			{Plaintext: []byte(merchantSecret)}, // Mapped Index 2
		},
	})
	if err != nil {
		s.logger.Error(ctx, "failed while encrypting merchant credentials", logger.Error(err))
		return errors.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser)

	} else if total := len(wrappedSecrets); total != 3 {
		s.logger.Error(ctx, fmt.Sprintf("invalid encryption result, total %d should be 3", total))
		return errors.New(responseHttp.HttpErrInternal, constant.ErrInternalServerForUser)
	}
	// ATTENTION: The mapping results must match the batch request order.
	merchant.CallbackApiKey = &wrappedSecrets[0].Ciphertext
	merchant.CallbackApiKeyVersion = wrappedSecrets[0].KeyVersion
	merchant.JITApiKey = wrappedSecrets[1].Ciphertext
	merchant.JITApiKeyVersion = wrappedSecrets[1].KeyVersion

	if err := s.repo.Create(ctx, merchant); err != nil {
		s.logger.Error(ctx, "error when create merchant", logger.Error(err))
		return errors.New(responseHttp.HttpErrInternal, err)
	}

	if user != nil {
		// update user merchant_id
		user.MerchantId = merchant.UUID
		if err := s.User.Update(ctx, user); err != nil {
			s.logger.Error(ctx, "error when update user", logger.Error(err))
			return errors.New(responseHttp.HttpErrInternal, err)
		}
	}

	// generate public and private key
	privKey, err := s.cryptoExt.GenerateRandomPKCS8Key()
	if err != nil {
		s.logger.Error(ctx, "error when generate public and private key", logger.Error(err))
		return errors.New(responseHttp.HttpErrInternal, err)
	}

	merchantAuthUUID := uuid.New()

	secretKey := s.cryptoExt.SecretKeyFromUUID(merchantAuthUUID)
	encPriv, err := s.cryptoExt.Encrypt(string(privKey), secretKey)
	if err != nil {
		s.logger.Error(ctx, "error when encrypt key", logger.Error(err))
		return errors.New(responseHttp.HttpErrInternal, err)
	}

	// create default merchant auth
	merchantAuth := &merchantModel.MerchantAuth{
		UUID:          merchantAuthUUID.String(),
		MerchantID:    merchant.UUID,
		Secret:        wrappedSecrets[2].Ciphertext,
		SecretVersion: wrappedSecrets[2].KeyVersion,
		SnapPrivateKey: sql.NullString{
			String: encPriv,
			Valid:  true,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	// skip on error
	_ = s.repo.CreateMerchantAuth(ctx, merchantAuth)

	// create account for merchant
	err = s.createAccountForMerchant(ctx, merchant.UUID)
	if err != nil {
		s.logger.Error(ctx, "error when create account for merchant", logger.Error(err))
		return errors.New(responseHttp.HttpErrInternal, err)
	}

	// publish event
	eventRequest := &pb.EventMerchantActionRequest{
		Event: constant.EventMerchantCreated,
		Data:  merchant.ToProtoDataEvent(),
	}
	payload, err := proto.Marshal(eventRequest)
	errors.LogProtoMarshalError(ctx, s.logger, err, eventRequest)

	s.logger.Info(ctx, "Send event to sync create new merchant", logger.String("data", base64.StdEncoding.EncodeToString(payload)))

	_ = s.rabbitMqExt.Publish(ctx, rabbitMqExt.MerchantActionRoutingKey, nil, payload)

	return nil
}

func (s *MerchantService) createAccountForMerchant(ctx context.Context, merchantID string) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/merchant/createAccountForMerchant")
	defer segment.End()

	merchantUUID, _ := uuid.Parse(merchantID)
	defaultCurrency := "IDR"

	// we don't need to check err because userType and usecase is already valid
	account, _ := account_model.NewAccount(&account_model.NewAccountRequest{
		ReferenceID: merchantUUID,
		UserType:    constant.UserTypeMerchant,
		Usecase:     constant.TypeDisbursement,
		Currency:    defaultCurrency,
	})

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return errors.New(responseHttp.HttpErrDatabase, err)
	}

	return nil
}
