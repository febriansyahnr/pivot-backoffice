package constant

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type EmailPriority int

const (
	EmailPrioritL0 EmailPriority = iota
	EmailPrioritL1
	EmailPrioritL2
)

func (e EmailPriority) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.Itoa(int(e)))
}

const (
	ResetPasswordEvent             = "backend-portal:otp-reset-password"
	ResetPINEvent                  = "backend-portal:otp-reset-pin"
	UserInvitationEvent            = "backend-portal:team-member-invitation"
	UserInvitationWithCredEvent    = "backend-portal:team-member-invitation-with-cred"
	UserLoginEvent                 = "backend-portal:otp-login"
	FirstTimeLoginEvent            = "backend-portal:otp-first-time-login"
	ChangePasswordEvent            = "backend-portal:otp-change-password"
	UserLoginActivityEvent         = "backend-portal:login-activity"
	ForceAutoWithdrawalNotifyEvent = "backend-portal:force-auto-withdrawal-notify"
)

const (
	// Format:
	//
	// %s = User unique id
	NotificationRoutingKeyFmt       = "backend-portal.notifications.%s"
	CreateBulkDisbursementNotifType = "create-bulk-disbursement"
	CreateVAPaymentNotifType        = "notification-va-payment"
	CreateQrisPaymentNotifType      = "notification-qris-payment"
	CreateCardPaymentNotifType      = "notification-card-payment"
)

func GetNotificationMessage(paymentUUID, status string) (string, string) {
	var subject, message string

	switch status {
	case "SUCCESS":
		subject = "Payment Success"
		message = fmt.Sprintf("Your payment <b>%s</b> has been successfully paid.", paymentUUID)
	case "FAILED":
		subject = "Payment Failed"
		message = fmt.Sprintf("Your payment <b>%s</b> has failed.", paymentUUID)
	default:
		subject = "Payment " + status
		message = fmt.Sprintf("Your payment <b>%s</b> has %s.", paymentUUID, status)
	}

	return subject, message
}
