package callbackService_test

import "github.com/stretchr/testify/mock"

var (
	validApiKey         = "api-key"
	callbackReqMockType = mock.AnythingOfType("callbackModel.CallbackRequest")
	ptrCallbackMockType = mock.AnythingOfType("*callback_model.Callback")
)
