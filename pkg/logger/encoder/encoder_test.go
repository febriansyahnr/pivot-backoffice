package encoder_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/logger/encoder"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

const maskedFields = "password,token,accessToken,otp,otpCode,refreshToken,secrets,secret,Authorization"

type TestUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Others   string `json:"others"`
}

type TestRespMap struct {
	Errors []string               `json:"errors"`
	Data   map[string]interface{} `json:"data"`
}

type TestRespPtrMap struct {
	Errors []string                `json:"errors"`
	Data   *map[string]interface{} `json:"data"`
}

type TestRespSliceInterface struct {
	List []interface{} `json:"list"`
}

type TestSliceString struct {
	Profile map[string]interface{} `json:"profile"`
	Secrets [3]string              `json:"secrets"` // Masked Array
	Ip      [1]string              `json:"ip"`
	Numbers []int                  `json:"numbers"`
	Errors  []string               `json:"errors"`
}

type TestNestedParentStruct struct {
	Header string                `json:"header"`
	Child  TestNestedChildStruct `json:"child"`
}

type TestNestedChildStruct struct {
	Data interface{} `json:"data"`
}

type TestNilValue struct {
	F1 *string           `json:"f1"`
	F2 interface{}       `json:"f2"`
	F3 []*string         `json:"f3"`
	F4 map[string]string `json:"f4"`
	F5 *[]TestUser       `json:"f5"`
}

type TestParentNilValue struct {
	F1    *map[string]interface{} `json:"f1"`
	F2    *[]*string              `json:"f2"`
	Child TestChildNilValue       `json:"child"`
}

type TestChildNilValue struct {
	F1 *interface{}        `json:"f1"`
	F2 *map[string]*string `json:"f2"`
	F3 *TestUser           `json:"f3"`
}

func TestEncoderJSONWithMaskSensitiveData(t *testing.T) {
	encoderConfig := zapcore.EncoderConfig{}

	enc := NewJSONEncoder(encoderConfig, WithMaskSensitiveData(strings.Split(maskedFields, ",")))

	createDataFromBuffer := func(data interface{}) interface{} {
		buf := new(bytes.Buffer)
		require.NoError(t, json.NewEncoder(buf).Encode(data))

		var tmp interface{}
		require.NoError(t, json.Unmarshal(buf.Bytes(), &tmp))

		return tmp
	}

	tests := []struct {
		name   string
		fields []zap.Field
		want   string
	}{
		{
			name:   "Preparation",
			fields: []zap.Field{},
			want:   `{}`,
		},
		{
			name: "Empty Key",
			fields: []zap.Field{
				{
					Type:   zapcore.StringType,
					String: "Test empty key",
				},
			},
			want: `{"": "Test empty key"}`,
		},
		{
			name: "Nil Value",
			fields: []zap.Field{
				zap.Any("data", TestNilValue{}),
			},
			want: `{"data":{"f1":null,"f2":null,"f3":null,"f4":null,"f5":null}}`,
		},
		{
			name: "Nested Nil Value",
			fields: []zap.Field{
				zap.Any("data", TestParentNilValue{}),
			},
			want: `{"data":{"f1":null,"f2":null,"child":{"f1":null,"f2":null,"f3":null}}}`,
		},
		{
			name: "Bytes From Struct TestNilValue",
			fields: []zap.Field{
				zap.Any("data", createDataFromBuffer(&TestNilValue{})),
			},
			want: `{"data":{"f1":null,"f2":null,"f3":null,"f4":null,"f5":null}}`,
		},
		{
			name: "Bytes From Struct TestParentNilValue",
			fields: []zap.Field{
				zap.Any("data", createDataFromBuffer(&TestParentNilValue{})),
			},
			want: `{"data":{"f1":null,"f2":null,"child":{"f1":null,"f2":null,"f3":null}}}`,
		},
		{
			name: "Other Fields",
			fields: []zap.Field{
				zap.Bool("bool", true),
				zap.Duration("duration", time.Millisecond),
				zap.Int64("int64", 64),
				zap.Float64("float64", 32.0),
				zap.Namespace("nameSpace"),
			},
			want: `{"bool":true,"duration":1000000,"int64":64,"float64":32,"nameSpace":{}}`,
		},
		{
			name: "Error Field",
			fields: []zap.Field{
				zap.String("message", "Failed"),
				zap.Error(errors.New("Stuck")),
				zap.NamedError("error2", errors.New("encoded")),
			},
			want: `{"message":"Failed","error":"Stuck","error2":"encoded"}`,
		},
		{
			name: "Simple Zap Fields",
			fields: []zap.Field{
				zap.String("username", "john"),
				zap.String("password", "123456"),
			},
			want: `{"username":"john","password":"******"}`,
		},
		{
			name: "All Fields Are Masked",
			fields: []zap.Field{
				zap.String("password", "123"),
				zap.String("token", "abcde"),
				zap.String("accesstoken", "acs"),
				zap.String("otp", "123456"),
				zap.String("OTPCode", "999999"),
				zap.String("refreshToken", "a"),
			},
			want: `{"password":"***","token":"*****","accesstoken":"***","otp":"******","OTPCode":"******","refreshToken":"*"}`,
		},
		{
			name: "Secret List",
			fields: []zap.Field{
				zap.Any("secrets", []string{"1", "12", "123"}),
			},
			want: `{"secrets":["*","**","***"]}`,
		},
		{
			name: "Secret List With Custom Field",
			fields: []zap.Field{
				zap.Any("secrets", []string{"1", "12", "123"}),
				zap.Any("data", []map[string]string{
					{"name": "John"},
					{"name": "Endru"},
				}),
				zap.Any("errors", []string{}),
			},
			want: `{"secrets":["*","**","***"], "data":[{"name":"John"},{"name":"Endru"}], "errors":[]}`,
		},
		{
			name: "Struct",
			fields: []zap.Field{
				zap.Any("data", TestUser{
					Username: "john_wick",
					Password: "qwerty",
					Others:   "-",
				}),
			},
			want: `{"data":{"username":"john_wick","password":"******","others":"-"}}`,
		},
		{
			name: "Pointer Struct",
			fields: []zap.Field{
				zap.Any("data", &TestUser{
					Username: "hello",
					Password: "123456",
					Others:   "none",
				}),
			},
			want: `{"data":{"username":"hello","password":"******","others":"none"}}`,
		},
		{
			name: "Struct With Map Field",
			fields: []zap.Field{
				zap.Any("response", TestRespMap{
					Errors: []string{},
					Data: map[string]interface{}{
						"username":    "joko",
						"accessToken": "xxcfbghg",
					},
				}),
			},
			want: `{"response": {"errors": [], "data": {"username": "joko", "accessToken": "********"}}}`,
		},
		{
			name: "Struct With Ptr Map Field",
			fields: []zap.Field{
				zap.Any("response", TestRespMap{
					Errors: []string{},
					Data: map[string]interface{}{
						"username":    "hendru",
						"accessToken": "acd",
					},
				}),
			},
			want: `{"response": {"errors": [], "data": {"username": "hendru", "accessToken": "***"}}}`,
		},
		{
			name: "Struct With Slice Interface",
			fields: []zap.Field{
				zap.Any("request", TestRespSliceInterface{
					List: []interface{}{
						"ABC", 123, true,
					},
				}),
			},
			want: `{"request": {"list": ["ABC", 123, true]}}`,
		},
		{
			name: "Struct With Slice String",
			fields: []zap.Field{
				zap.Any("response", TestSliceString{
					Profile: map[string]interface{}{
						"id":    1,
						"email": "john.wick@harsya.com",
					},
					Secrets: [3]string{"abc", "123456", "jn"},
					Ip:      [1]string{"1.2.3.4"},
					Numbers: []int{1, 2, 3, 4},
					Errors:  []string{},
				}),
			},
			want: `{"response": {"profile": {"id": 1, "email":"john.wick@harsya.com"}, "secrets": ["***", "******", "**"], "ip": ["1.2.3.4"], "numbers": [1,2,3,4], "errors": []}}`,
		},
		{
			name: "Struct With Nested Field",
			fields: []zap.Field{
				zap.Any("body", TestNestedParentStruct{
					Header: "Info",
					Child: TestNestedChildStruct{
						Data: map[string]string{
							"OTPCode": "123456",
						},
					},
				}),
			},
			want: `{"body": {"header": "Info", "child": {"data": {"OTPCode": "******"}}}}`,
		},
		{
			name: "Slice Pointer Struct",
			fields: []zap.Field{
				zap.Any(
					"data", []*TestUser{
						{Username: "A1", Password: "ABC", Others: "A"},
						{Username: "B1", Password: "1234", Others: "B"},
					}),
			},
			want: `{"data": [{"username":"A1","password":"***","others":"A"}, {"username":"B1","password":"****","others":"B"}]}`,
		},
		{
			name: "Slice Pointer Map Interface",
			fields: []zap.Field{
				zap.Any("data", []*map[string]interface{}{
					{"id": 1, "name": "john"}, {"id": 2, "name": "hendru"},
				}),
			},
			want: `{"data": [{"id":1, "name":"john"}, {"id": 2, "name":"hendru"}]}`,
		},
		{
			name: "Map Interface",
			fields: []zap.Field{
				zap.Any("data", map[string]interface{}{
					"retries":     4,
					"time":        time.Time{},
					"accessToken": "qwerty.abc",
				}),
			},
			want: `{"data": {"retries":4, "time":"0001-01-01T00:00:00Z", "accessToken":"**********"}}`,
		},
		{
			name: "Map Interface With Nested Map",
			fields: []zap.Field{
				zap.Any("data", map[string]interface{}{
					"id": 1,
					"profile": map[string]interface{}{
						"name":     "widyaadebagus",
						"password": "qwerty!@#",
						"access": map[string]string{
							"token": "token",
						},
					},
					"secret":    "1234",
					"registers": map[string]int{"first": 1},
				}),
			},
			want: `{"data": {"id": 1, "profile": {"name": "widyaadebagus", "password": "*********","access": {"token": "*****"}}, "secret":"****", "registers": {"first": 1}}}`,
		},
		{
			name: "Map Interface With Slice Field",
			fields: []zap.Field{
				zap.Any("response", map[string]interface{}{
					"profile": map[string]interface{}{
						"id":    1,
						"email": "john.wick@harsya.com",
					},
					"secrets": []string{"abc", "123456", "jn"}, // Masked Slice
					"ip":      [1]string{"1.2.3.4"},
					"numbers": []int{1, 2, 3, 4},
					"errors":  []string{},
				}),
			},
			want: `{"response": {"profile": {"id": 1, "email":"john.wick@harsya.com"}, "secrets": ["***", "******", "**"], "ip": ["1.2.3.4"], "numbers": [1,2,3,4], "errors": []}}`,
		},
		{
			name: "Map String With Value Slice String",
			fields: []zap.Field{
				zap.Any("request", map[string]interface{}{
					"headers": http.Header{
						"Content-Type":  {"application/json"},
						"Authorization": {"Bearer token"},
					}}),
			},
			want: `{"request": {"headers": {"Content-Type": ["application/json"], "Authorization": ["************"]}}}`,
		},
		{
			name: "Bytes String",
			fields: []zap.Field{
				zap.ByteString("request", []byte(`{"username":"johnwick", "password":"abc123456"}`)),
			},
			want: `{"request": "{\"password\":\"*********\",\"username\":\"johnwick\"}"}`,
		},
		{
			name: "Bytes String With Invalid Bind To Map",
			fields: []zap.Field{
				zap.ByteString("request", []byte(`{x}`)),
			},
			want: `{"request": "null"}`,
		},
		{
			name: "Bytes",
			fields: []zap.Field{
				zap.Binary("request", []byte(`{"username":"johnwick", "password":"abc123456"}`)),
			},
			want: `{"request": "eyJwYXNzd29yZCI6IioqKioqKioqKiIsInVzZXJuYW1lIjoiam9obndpY2sifQ=="}`, // Encode Base64: {"password":"*********","username":"johnwick"}
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			beforeEncodeFields := make([]zap.Field, len(test.fields))
			copy(beforeEncodeFields, test.fields)

			buf, err := enc.EncodeEntry(zapcore.Entry{}, test.fields)
			require.NoError(t, err)

			if !assert.JSONEq(t, test.want, buf.String()) {
				t.Log("Output:", buf.String())
			}
			assert.Equal(t, beforeEncodeFields, test.fields)

			buf.Free()
		})
	}
}

func TestEncoderJSONWithoutMaskSensitiveData(t *testing.T) {

	enc := NewJSONEncoder(zapcore.EncoderConfig{})

	tests := []struct {
		name   string
		fields []zap.Field
		want   string
	}{
		{
			name: "Unmask specific field",
			fields: []zap.Field{
				zap.String("username", "john"),
				zap.String("password", "123456"),
			},
			want: `{"username":"john","password":"123456"}`,
		},
		{
			name: "All fields are masked",
			fields: []zap.Field{
				zap.String("password", "234567"),
				zap.String("token", "56yhgbfg"),
				zap.String("accesstoken", "acs"),
				zap.String("otp", "123456"),
				zap.String("OTPCode", "999999"),
				zap.String("refreshToken", "afghkmnb"),
			},
			want: `{"password":"234567","token":"56yhgbfg","accesstoken":"acs","otp":"123456","OTPCode":"999999","refreshToken":"afghkmnb"}`,
		},
		{
			name: "Secret List",
			fields: []zap.Field{
				zap.Any("secrets", []string{"1", "12", "123"}),
			},
			want: `{"secrets":["1","12","123"]}`,
		},
		{
			name: "Secret List With Custom Field",
			fields: []zap.Field{
				zap.Any("secrets", []string{"1", "12", "123"}),
				zap.Any("data", []map[string]string{
					{"name": "John"},
					{"name": "Endru"},
				}),
				zap.Any("errors", []string{}),
			},
			want: `{"secrets":["1","12","123"], "data":[{"name":"John"},{"name":"Endru"}], "errors":[]}`,
		},
		{
			name: "Struct",
			fields: []zap.Field{
				zap.Any("data", TestUser{
					Username: "john_wick",
					Password: "qwerty",
					Others:   "-",
				}),
			},
			want: `{"data":{"username":"john_wick","password":"qwerty","others":"-"}}`,
		},
		{
			name: "Pointer Struct",
			fields: []zap.Field{
				zap.Any("data", &TestUser{
					Username: "hello",
					Password: "123456",
					Others:   "none",
				}),
			},
			want: `{"data":{"username":"hello","password":"123456","others":"none"}}`,
		},
		{
			name: "Struct With Map Field",
			fields: []zap.Field{
				zap.Any("response", TestRespMap{
					Errors: []string{},
					Data: map[string]interface{}{
						"username":    "joko",
						"accessToken": "xxcfbghg",
					},
				}),
			},
			want: `{"response": {"errors": [], "data": {"username": "joko", "accessToken": "xxcfbghg"}}}`,
		},
		{
			name: "Struct With Ptr Map Field",
			fields: []zap.Field{
				zap.Any("response", TestRespMap{
					Errors: []string{},
					Data: map[string]interface{}{
						"username":    "hendru",
						"accessToken": "acd",
					},
				}),
			},
			want: `{"response": {"errors": [], "data": {"username": "hendru", "accessToken": "acd"}}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			beforeEncodeFields := make([]zap.Field, len(test.fields))
			copy(beforeEncodeFields, test.fields)

			buf, err := enc.EncodeEntry(zapcore.Entry{}, test.fields)
			require.NoError(t, err)

			if !assert.JSONEq(t, test.want, buf.String()) {
				t.Log("Output:", buf.String())
			}
			assert.Equal(t, beforeEncodeFields, test.fields)

			buf.Free()
		})
	}
}

func TestWithInitialFields(t *testing.T) {
	enc := NewJSONEncoder(
		zapcore.EncoderConfig{}, WithMaskSensitiveData(strings.Split(maskedFields, ",")),
	)

	ws := &zaptest.Buffer{}
	ws.Reset()

	logger := zap.New(zapcore.NewCore(enc, ws, zap.NewAtomicLevelAt(zap.InfoLevel)))

	initialFields := []zapcore.Field{
		zap.String("app", "backend"), zap.String("env", "testing"),
	}

	for i := 0; i < 32; i++ {
		logger.Core().With(initialFields).Write(
			zapcore.Entry{
				Message: "testing",
			},
			[]zap.Field{
				zap.Any("data", map[string]string{"message": "OK", "secret": "abc1"}),
			},
		)
		if !assert.JSONEq(t, `{"app":"backend","env":"testing","data":{"message":"OK","secret":"****"}}`, ws.String()) {
			t.Log("Output:", ws.String())
		}
		ws.Reset()
	}
}
