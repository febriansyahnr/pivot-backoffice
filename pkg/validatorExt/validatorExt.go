package validatorExt

import (
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslator "github.com/go-playground/validator/v10/translations/en"
	"github.com/shopspring/decimal"
)

type Validate = validator.Validate

var (
	once sync.Once
	vld  *Validate
)

func RegTypeShopspringDecimal(vld *Validate) {
	vld.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if valuer, ok := field.Interface().(decimal.Decimal); ok {
			val, _ := valuer.Float64()
			return val
		}
		return nil
	}, decimal.Decimal{})
}

func New() *Validate {
	once.Do(func() {
		vld = validator.New(validator.WithRequiredStructEnabled())
		RegTypeShopspringDecimal(vld)
		_ = vld.RegisterValidation("iso_8601_datetime", Iso8601Datetime)
		_ = vld.RegisterValidation("alphanumspace", AlphanumericSpace)
		_ = vld.RegisterValidation("maxChar", StringMaxCharacter)
		_ = vld.RegisterValidation("slicestrcontains", SliceStrContains)

		// Register validation for special characters
		vld.RegisterValidation("nospecialchars", NoSpecialChars)

		// Register validation for only numbers
		vld.RegisterValidation("numberstring", NumberString)

		_ = enTranslator.RegisterDefaultTranslations(vld, GetTranslator())

		// Name Extractor
		vld.RegisterTagNameFunc(func(fl reflect.StructField) string {
			if name := strings.TrimSpace(fl.Tag.Get("name")); name != "" {
				return name
			}
			return fl.Name
		})

		// XB Payout Method
		vld.RegisterValidation("xb_payout_method", XBPayoutMethod)

		// Luhn algorithm for card number validation
		vld.RegisterValidation("luhn", LuhnCheck)
	})
	return vld
}

func GetTranslator() ut.Translator {
	en := en.New()
	uni := ut.New(en, en)

	trans, _ := uni.GetTranslator("en")

	return trans
}

func Iso8601Datetime(fl validator.FieldLevel) bool {
	if _, err := time.Parse(constant.ISO8601Datetime, fl.Field().String()); err != nil {
		return false
	}
	return true
}

func AlphanumericSpace(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Only allow letters, numbers, and spaces
	regex := regexp.MustCompile(`^[a-zA-Z0-9 ]*$`)
	return regex.MatchString(value)
}

func NumberString(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Only allow numbers
	regex := regexp.MustCompile(`^[0-9]+$`)
	return regex.MatchString(value)
}

func NoSpecialChars(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Only allow letters, numbers, and spaces
	regex := regexp.MustCompile(`^[a-zA-Z0-9 ]*$`)
	return regex.MatchString(value)
}

func StringMaxCharacter(fl validator.FieldLevel) bool {
	max, err := strconv.Atoi(fl.Param())
	if err != nil {
		return false
	}
	return utf8.RuneCountInString(fl.Field().String()) <= max
}

func SliceStrContains(fl validator.FieldLevel) bool {
	params := strings.Split(fl.Param(), " ")
	if len(params) == 0 || !fl.Field().CanInterface() {
		return false
	}

	values, ok := fl.Field().Interface().([]string)
	if !ok {
		return false
	}
	for _, p := range params {
		if !slices.Contains(values, p) {
			return false
		}
	}
	return true
}

func LuhnCheck(fl validator.FieldLevel) bool {
	number := fl.Field().String()
	sum := 0
	nDigits := len(number)
	parity := nDigits % 2

	for i := range nDigits {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}

func XBPayoutMethod(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		// only validate when value is not empty
		return true
	}

	supportedPayoutMethod := []string{
		constant.XbPayoutMethodBank,
		constant.XbPayoutMethodWallet,
		constant.XbPayoutMethodCash,
	}

	// validate value can only "bank", "wallet", "cash"
	return slices.Contains(supportedPayoutMethod, strings.ToUpper(value))
}
