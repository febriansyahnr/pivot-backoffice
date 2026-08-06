package constant

import snapConstant "github.com/paper-indonesia/pdk/go/snap"

const (
	CallbackNamePayment                      = "PAYMENT"
	CallbackNameDisbursement                 = "PAYOUT" // for OPEN API only
	CallbackNameXB                           = "INTERNATIONAL_PAYOUT"
	CallbackNameVirtualCard                  = "VIRTUAL_CARD"
	CallbackNameWalletTopup                  = "WALLET_TOP_UP"
	CallbackNameWalletUserActivationName     = "WALLET_USER_ACTIVATION"
	CallbackNameWalletActivateAccountLinkage = "WALLET_ACCOUNT_LINKAGE_ACTIVATION"
	CallbackNameWalletTransaction            = "WALLET_TRANSACTION"
	CallbackNameWalletSNAPQrisMPM            = "WALLET_SNAP_QRIS_MPM"
	CallbackNameWalletSNAPDirectDebit        = "WALLET_SNAP_DIRECT_DEBIT"
	CallbackNameRefund                       = "REFUND"
	CallbackNameMerchantTopUp                = "MERCHANT_TOP_UP"
	CallbackNameSubAccountRegistration       = "SUB_ACCOUNT_REGISTRATION"
	CallbackNameWithdrawal                   = "WITHDRAWAL"
)

const (
	CallbackStatusPending   = "PENDING"
	CallbackStatusDelivered = "DELIVERED"
	CallbackStatusFailed    = "FAILED"
	CallbackStatusApproved  = "APPROVED"
	CallbackStatusRejected  = "REJECTED"
)

const (
	CallbackEventPayoutDone                   = "PAYOUT.DONE"
	CallbackEventPayoutPending                = "PAYOUT.PENDING"
	CallbackEventPayoutDelayed                = "PAYOUT.DELAYED"
	CallbackEventPayoutCancelled              = "PAYOUT.CANCELLED"
	CallbackEventPayoutSuccess                = "PAYOUT.SUCCESS" // For Single payout
	CallbackEventPayoutFailed                 = "PAYOUT.FAILED"  // For Single payout
	CallbackEventAccessTokenB2b               = "PAYMENT.ACCESS-TOKEN-B2B"
	CallbackEventPaymentVirtualAccountPaid    = "PAYMENT.VIRTUAL-ACCOUNT.PAID"
	CallbackEventPaymentCreditcardPaid        = "PAYMENT.CREDIT-CARD.PAID"
	CallbackEventPaymentQrisMpmPaid           = "PAYMENT.QRIS-MPM.PAID"
	CallbackEventVirtualCardNotification      = "VIRTUAL-CARD.NOTIFICATION"
	CallbackEventVirtualCardVisaNotification  = "VIRTUAL-CARD.VISA.NOTIFICATION"
	CallbackEventPhysicalCardVisaNotification = "PHYSICAL-CARD.VISA.NOTIFICATION"
	CallbackEventUnifiedPaymentPattern        = "PAYMENT.%s"
	CallbackEventUnifiedPaymentChargePattern  = "CHARGE.%s"
	CallbackEventWalletTopUp                  = "WALLET.TOP-UP"
	CallbackEventWalletUserActivation         = "WALLET.USER-ACTIVATION"
	CallbackEventWalletUserActivationKYC      = "WALLET.USER-ACTIVATION.KYC"
	CallbackEventWalletActivateAccountLinkage = "WALLET.ACCOUNT_LINKAGE.ACTIVATION"
	CallbackEventWalletTransaction            = "WALLET.TRANSACTION"
	CallbackEventWalletSNAPQrisMPM            = "WALLET.SNAP.QRIS-MPM"
	CallbackEventWalletSNAPDirectDebit        = "WALLET.SNAP.DIRECT-DEBIT"
	CallbackEventRefundPending                = "REFUND.PENDING"
	CallbackEventRefundWaitingBankTransfer    = "REFUND.WAITING_BANK_TRANSFER"
	CallbackEventRefundSuccess                = "REFUND.SUCCESS"
	CallbackEventRefundFailed                 = "REFUND.FAILED"
	CallbackEventRefundCancelled              = "REFUND.CANCELLED"
	CallbackEventRefundPattern                = "REFUND.%s"
	CallbackEventAccessTokenB2bTest           = "PAYMENT.ACCESS-TOKEN-B2B.TEST"
	CallbackEventPaymentVirtualAccountTest    = "PAYMENT.VIRTUAL-ACCOUNT.TEST"
	CallbackEventPaymentQrisMpmTest           = "PAYMENT.QRIS-MPM.TEST"
	CallbackEventWalletSNAPQrisMPMTest        = "WALLET.SNAP.QRIS-MPM.TEST"
	CallbackEventWalletSNAPDirectDebitTest    = "WALLET.SNAP.DIRECT-DEBIT.TEST"
	CallbackEventMerchantTopUpSuccess         = "MERCHANT-TOP-UP.SUCCESS"

	CallbackEventSubAccountRegistrationPattern = "SUB.ACTIVATION.%s"
	CallbackEventWithdrawPattern               = "WITHDRAW.%s"
)

const (
	CallbackSnapAccessTokenB2bEndpoint               = "/snap/api/v1.0/access-token/b2b"
	CallbackSnapAccessTokenB2bEndpointWithoutVersion = "/access-token/b2b"
	CallbackSnapTransferVaPaymentEndpoint            = "/transfer-va/payment"
	CallbackSnapQrMpmNotifyEndpoint                  = "/qr/qr-mpm-notify"
	CallbackSnapWalletQrisMpmEndpoint                = "/wallet/qris-mpm-notify"
	CallbackSnapWalletDirectDebitEndpoint            = "/wallet/direct-debit-notify"
)

const (
	CallbackMasterPaymentSNAPAccessTokenB2b = "PAYMENT_SNAP_ACCESS_TOKEN"
	CallbackMasterPaymentSNAPVA             = "PAYMENT_SNAP_VA"
	CallbackMasterPaymentSNAPQRIS           = "PAYMENT_SNAP_QRIS"
	CallbackMasterWalletSNAPQrisMPM         = "WALLET_SNAP_QRIS_MPM"
	CallbackMasterWalletSNAPDirectDebit     = "WALLET_SNAP_DIRECT_DEBIT"
	CallbackRefundID                        = "REFUND"
)

func GetCallbackEventTitle(event string) string {
	switch event {
	case CallbackEventPayoutDone:
		return "Payout Done"
	case CallbackEventPayoutPending:
		return "Payout Pending"
	case CallbackEventPaymentVirtualAccountPaid:
		return "Payment Virtual Account Paid"
	case CallbackEventPaymentCreditcardPaid:
		return "Payment Creditcard Paid"
	case CallbackEventPaymentQrisMpmPaid:
		return "Payment QRIS MPM Paid"
	case CallbackEventVirtualCardNotification:
		return "Virtual Card Master Card Notification"
	case CallbackEventRefundPending:
		return "Refund Pending"
	case CallbackEventRefundWaitingBankTransfer:
		return "Refund Waiting Bank Transfer"
	case CallbackEventRefundSuccess:
		return "Refund Success"
	case CallbackEventRefundFailed:
		return "Refund Failed"
	case CallbackEventRefundCancelled:
		return "Refund Cancelled"
	}

	return ""
}

func IsPaymentVA(event string) bool {
	return event == CallbackEventPaymentVirtualAccountPaid
}

func BuildSnapChannelIdByEvent(event string) string {
	switch event {
	case CallbackEventPaymentVirtualAccountPaid, CallbackEventPaymentVirtualAccountTest:
		return snapConstant.SNAP_SERVICE_VIRTUAL_ACCOUNT
	case CallbackEventPaymentQrisMpmPaid, CallbackEventPaymentQrisMpmTest:
		return "95221" // based on ASPI dev
	}
	return ""
}

func BuildCallbackSnapEndpointByEvent(event string) string {
	switch event {
	case CallbackEventPaymentVirtualAccountPaid:
		return CallbackSnapTransferVaPaymentEndpoint
	case CallbackEventPaymentQrisMpmPaid:
		return CallbackSnapQrMpmNotifyEndpoint // based on ASPI dev
	}

	return ""
}

func BuildCallbackSnapEndpointByName(event string) string {
	switch event {
	case CallbackMasterPaymentSNAPAccessTokenB2b:
		return CallbackSnapAccessTokenB2bEndpointWithoutVersion
	case CallbackMasterPaymentSNAPVA:
		return CallbackSnapTransferVaPaymentEndpoint
	case CallbackMasterPaymentSNAPQRIS:
		return CallbackSnapQrMpmNotifyEndpoint
	case CallbackMasterWalletSNAPQrisMPM:
		return CallbackSnapWalletQrisMpmEndpoint
	case CallbackMasterWalletSNAPDirectDebit:
		return CallbackSnapWalletDirectDebitEndpoint
	}

	return ""
}

func BuildCallbackSnapStatusTestByName(event string) string {
	switch event {
	case CallbackMasterPaymentSNAPAccessTokenB2b:
		return CallbackEventAccessTokenB2bTest
	case CallbackMasterPaymentSNAPVA:
		return CallbackEventPaymentVirtualAccountTest
	case CallbackMasterPaymentSNAPQRIS:
		return CallbackEventPaymentQrisMpmTest
	case CallbackMasterWalletSNAPQrisMPM:
		return CallbackEventWalletSNAPQrisMPMTest
	case CallbackMasterWalletSNAPDirectDebit:
		return CallbackEventWalletSNAPDirectDebitTest
	}

	return ""
}

func BuildCallbackSnapStatusSuccessByEvent(event string) string {
	switch event {
	case CallbackEventAccessTokenB2bTest:
		return CallbackEventAccessTokenB2b
	case CallbackEventPaymentVirtualAccountTest:
		return CallbackEventPaymentVirtualAccountPaid
	case CallbackEventPaymentQrisMpmTest:
		return CallbackEventPaymentQrisMpmPaid
	case CallbackEventWalletSNAPQrisMPMTest:
		return CallbackEventWalletSNAPQrisMPM
	case CallbackEventWalletSNAPDirectDebitTest:
		return CallbackEventWalletSNAPDirectDebit
	}

	return ""
}
