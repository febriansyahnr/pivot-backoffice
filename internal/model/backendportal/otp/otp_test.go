package otp_test

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshalBinary(t *testing.T) {
	// Create a test OTPCache with one OTP
	src := OTPCache{
		OTPList: []OTPList{
			{
				OTP:       "0987654",
				ExpiredAt: time.Now().Add(10 * time.Minute),
				Verify:    true,
			},
		},
	}
	dst := OTPCache{}

	// Marshal the source OTPCache
	data, err := src.MarshalBinary()
	assert.Nil(t, err)
	
	// Unmarshal into the destination OTPCache
	err = dst.UnmarshalBinary(data)
	assert.Nil(t, err)
	
	// Verify they match
	assert.Equal(t, len(src.OTPList), len(dst.OTPList))
	assert.Equal(t, src.OTPList[0].OTP, dst.OTPList[0].OTP)
	assert.Equal(t, src.OTPList[0].Verify, dst.OTPList[0].Verify)
	assert.WithinDuration(t, src.OTPList[0].ExpiredAt, dst.OTPList[0].ExpiredAt, time.Second)
}

func TestSendOTPReqUnmarshalJSON(t *testing.T) {
	tests := []struct {
		payload    string
		wantErr    string
		wantResult SendOTPReq
	}{
		{
			payload: `B`,
			wantErr: "invalid character 'B' looking for beginning of value",
		},
		{
			payload: `{"email": "email01@example.com", "event": "invalid-event"}`,
			wantErr: "event not registered",
		},
		{
			payload: `{"email": "email01@example.com", "event": "forgot-password"}`,
			wantResult: SendOTPReq{
				Email: "email01@example.com",
				Event: constant.OTPIdentifierForgotPassword,
			},
		},
		{
			payload: `{"email": "email02@example.com", "event": "reset-pin"}`,
			wantResult: SendOTPReq{
				Email: "email02@example.com",
				Event: constant.OTPIdentifierResetPIN,
			},
		},
		{
			payload: `{"email": "email03@example.com", "event": "user-login"}`,
			wantResult: SendOTPReq{
				Email: "email03@example.com",
				Event: constant.OTPIdentifierUserLogin,
			},
		},
		{
			payload: `{"email": "email04@example.com", "event": "first-time-login"}`,
			wantResult: SendOTPReq{
				Email: "email04@example.com",
				Event: constant.OTPIdentifierFirstTimeLogin,
			},
		},
	}
	for _, test := range tests {
		data := SendOTPReq{}

		if err := data.UnmarshalJSON([]byte(test.payload)); test.wantErr == "" {
			require.Nil(t, err)
			assert.Equal(t, test.wantResult, data)

		} else {
			require.NotNil(t, err)
			assert.ErrorContains(t, err, test.wantErr)
		}
	}
}

func TestGetLatestOTP(t *testing.T) {
	tests := []struct {
		name     string
		otpCache OTPCache
		want     *OTPList
	}{
		{
			name:     "Empty OTP List",
			otpCache: OTPCache{OTPList: []OTPList{}},
			want:     nil,
		},
		{
			name: "Single OTP in List",
			otpCache: OTPCache{OTPList: []OTPList{
				{
					OTP:       "123456",
					ExpiredAt: time.Now().Add(time.Hour),
					Verify:    false,
				},
			}},
			want: &OTPList{
				OTP:       "123456",
				ExpiredAt: time.Now().Add(time.Hour),
				Verify:    false,
			},
		},
		{
			name: "Multiple OTPs in List",
			otpCache: OTPCache{OTPList: []OTPList{
				{
					OTP:       "111111",
					ExpiredAt: time.Now().Add(time.Hour),
					Verify:    true,
				},
				{
					OTP:       "222222",
					ExpiredAt: time.Now().Add(2 * time.Hour),
					Verify:    false,
				},
			}},
			want: &OTPList{
				OTP:       "222222",
				ExpiredAt: time.Now().Add(2 * time.Hour),
				Verify:    false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.otpCache.GetLatestOTP()
			
			if tt.want == nil {
				assert.Nil(t, got)
				return
			}
			
			// Only compare the OTP and Verify since ExpiredAt will have different timestamps
			assert.Equal(t, tt.want.OTP, got.OTP)
			assert.Equal(t, tt.want.Verify, got.Verify)
			
			// Just check if ExpiredAt is within a reasonable time range
			if tt.name == "Single OTP in List" {
				assert.WithinDuration(t, time.Now().Add(time.Hour), got.ExpiredAt, 2*time.Second)
			} else if tt.name == "Multiple OTPs in List" {
				assert.WithinDuration(t, time.Now().Add(2*time.Hour), got.ExpiredAt, 2*time.Second)
			}
		})
	}
}

func TestSuspendUser(t *testing.T) {
	src := SuspendUser{
		Status:     true,
		RetryAfter: time.Now().UTC(),
	}
	dst := SuspendUser{}

	// Marshal the source SuspendUser
	data, err := src.MarshalBinary()
	assert.Nil(t, err)
	
	// Unmarshal into the destination SuspendUser
	err = dst.UnmarshalBinary(data)
	assert.Nil(t, err)
	
	// Verify fields match
	assert.Equal(t, src.Status, dst.Status)
	assert.WithinDuration(t, src.RetryAfter, dst.RetryAfter, time.Second)
}
