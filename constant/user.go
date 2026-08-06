package constant

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
)

const (
	UserStatusActive   = "ACTIVE"
	UserStatusInactive = "INACTIVE"
	UserStatusInvited  = "INVITED"
	UserStatusBlocked  = "BLOCKED"

	UserLoggedInDeviceStatusActive = "ACTIVE"

	UserSystemType = "SYSTEM"

	UserLoginSessionTerminated  = "LOGIN_SESSION_TERMINATED"
	UserLoginRoleSessionChanged = "USER_ROLE_SESSION_CHANGED"
)

const (
	UserSortColName = "name"
)

type UserFeatureIdentifier string

const (
	UserIdentifierUserInvitation UserFeatureIdentifier = "krt6ujvmm8p6zl5orjci82nsod90ewzg"

	UserKeyFormatting                  = "backend-portal:users:%s:%s"
	UserKeyWithoutIdentifierFormatting = "backend-portal:users:%s"
	UserInvitationName                 = "user-invitation"
	UserTokenNamespace                 = "token"
	UserInvitationTotalResendField     = "total_resend"
	UserInvitationLastTokenField       = "last_token"
)

func (t *UserFeatureIdentifier) EmailSender() string {
	return config.DefaultEmailSender()
}

func (t *UserFeatureIdentifier) ExpireDuration() time.Duration {
	switch *t {
	default:
		return 0

	case UserIdentifierUserInvitation:
		return 24 * time.Hour
	}
}

func (t *UserFeatureIdentifier) FeatureName() string {
	switch *t {
	default:
		return ""

	case UserIdentifierUserInvitation:
		return UserInvitationName
	}
}

func (t *UserFeatureIdentifier) Event() string {
	switch *t {
	default:
		return ""

	case UserIdentifierUserInvitation:
		return UserInvitationEvent
	}
}

func (t *UserFeatureIdentifier) MexSendUserInvitation() int {
	return 5 // TODO: use config instead
}
