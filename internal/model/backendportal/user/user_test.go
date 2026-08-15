package user

import (
	"database/sql"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	now = time.Now()

	user = &User{
		UUID:       "uuid-uuid-uuid",
		Email:      "test@gmail.com",
		Name:       "test",
		Password:   "pass",
		Blocked:    sql.NullTime{Time: now.Add(-time.Hour), Valid: true},
		MerchantId: "merchant-id",
		CreatedAt:  now,
		PinHash:    sql.NullString{String: "123", Valid: true},
	}

	response = &UserResponse{
		UUID:       "uuid-uuid-uuid",
		Email:      "test@gmail.com",
		Name:       "test",
		Blocked:    now.Add(-time.Hour),
		MerchantId: "merchant-id",
		CreatedAt:  now,
		IsEmptyPin: 0,
	}
)

func TestUser_ToResponse(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *User
		Expected *UserResponse
	}{
		{
			Name:     "it should return user response",
			Input:    user,
			Expected: response,
		},
		{
			Name: "it should return user response with empty role",
			Input: &User{
				UUID:       "uuid-uuid-uuid",
				Email:      "test@gmail.com",
				Name:       "test",
				Password:   "pass",
				Blocked:    sql.NullTime{Time: now.Add(-time.Hour), Valid: true},
				MerchantId: "merchant-id",
				Role:       sql.NullString{String: "admin", Valid: true},
				CreatedAt:  now,
			},
			Expected: &UserResponse{
				UUID:       "uuid-uuid-uuid",
				Email:      "test@gmail.com",
				Name:       "test",
				Blocked:    now.Add(-time.Hour),
				MerchantId: "merchant-id",
				Role:       "admin",
				CreatedAt:  now,
				IsEmptyPin: 1,
			},
		},
		{
			Name: "it should return user response with deactivated user",
			Input: &User{
				UUID:       "uuid-uuid-uuid",
				Email:      "test@gmail.com",
				Name:       "test",
				Password:   "pass",
				Blocked:    sql.NullTime{Time: now.Add(-time.Hour), Valid: true},
				MerchantId: "merchant-id",
				Role:       sql.NullString{String: "admin", Valid: true},
				CreatedAt:  now,
				DeactivatedAt: sql.NullTime{
					Time:  now,
					Valid: true,
				},
			},
			Expected: &UserResponse{
				UUID:          "uuid-uuid-uuid",
				Email:         "test@gmail.com",
				Name:          "test",
				Blocked:       now.Add(-time.Hour),
				MerchantId:    "merchant-id",
				Role:          "admin",
				CreatedAt:     now,
				IsEmptyPin:    1,
				DeactivatedAt: now.String(),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			newUserResponse := tc.Input.ToResponse()
			require.Equal(t, tc.Expected, newUserResponse)
		})
	}
}

func TestUser_EncryptPassword(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    string
		Expected string
	}{
		{
			Name:     "it should return encrypted password",
			Input:    "pass",
			Expected: "d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			newUser := &User{}
			newUser.Password = newUser.EncryptPassword(tc.Input)

			require.Equal(t, tc.Expected, newUser.Password)
		})
	}
}

func TestUser_ComparePassword(t *testing.T) {
	expectedUser := &User{
		Password: "d74ff0ee8da3b9806b18c877dbf29bbde50b5bd8e4dad7a3a725000feb82e8f1", // pass
	}

	testCases := []struct {
		Name     string
		Input    string
		Output   *User
		Expected bool
	}{
		{
			Name:     "it should return true",
			Input:    "pass",
			Output:   expectedUser,
			Expected: true,
		},
		{
			Name:     "it should return false",
			Input:    "password",
			Output:   expectedUser,
			Expected: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			require.Equal(t, tc.Expected, tc.Output.ComparePassword(tc.Input))
		})
	}
}

func TestNewMerchantUser(t *testing.T) {
	testCases := []struct {
		Name  string
		Input *MerchantUserRequest
	}{
		{
			Name: "it should return new user",
			Input: &MerchantUserRequest{
				Email:      "test@gmail.com",
				Name:       "test",
				MerchantId: "merchant-id",
				Invitation: true,
			},
		},
		{
			Name: "it should return new user with empty password when invitation is false",
			Input: &MerchantUserRequest{
				Email:      "test2@gmail.com",
				Name:       "test2",
				MerchantId: "merchant-id-2",
				Invitation: false,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			newUser := NewMerchantUser(tc.Input)
			require.Equal(t, tc.Input.Email, newUser.Email)
			require.Equal(t, tc.Input.Name, newUser.Name)
			require.Equal(t, tc.Input.MerchantId, newUser.MerchantId)
			require.Equal(t, tc.Input.MerchantName, newUser.MerchantName)
			assert.Equal(t, newUser.Status, constant.UserStatusInvited)

			if tc.Input.Invitation {
				assert.NotEmpty(t, newUser.Password)
			} else {
				assert.Empty(t, newUser.Password)
			}
		})
	}
}
