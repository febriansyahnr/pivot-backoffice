package validation

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

type Fields map[string]interface{}

func (f *Fields) StatusCode() int {
	return http.StatusBadRequest
}

func (f *Fields) Code() string {
	return "40"
}

func (f *Fields) Error() string {
	buf, _ := json.Marshal(f)
	return string(buf)
}

type Validator interface {
	ScanStruct(input interface{}) error
}

type validate struct {
	*validator.Validate
	ut ut.Translator
}

func New() Validator {
	vld := validator.New()
	validatorExt.RegTypeShopspringDecimal(vld)
	_ = vld.RegisterValidation("slicestrcontains", validatorExt.SliceStrContains)

	english := en.New()
	uni := ut.New(english, english)

	trans, found := uni.GetTranslator("en")
	if !found {
		log.Panic("translator not found")
	}
	_ = enTranslations.RegisterDefaultTranslations(vld, trans)

	vld.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("label"), ",", 2)[0]
		if len(name) == 0 {
			name = strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		}

		if name == "-" {
			return ""
		}
		return name
	})

	return &validate{
		ut:       trans,
		Validate: vld,
	}
}

func (m *validate) ScanStruct(input interface{}) error {

	if input == nil {
		msg := "developer_vault. input cant be nil"
		return &Fields{"message": msg}
	}

	if err := m.Validate.Struct(input); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return &Fields{"message": err.Error()}
		}

		fields := Fields{}
		for _, e := range err.(validator.ValidationErrors) {

			structName := strings.Split(e.Namespace(), ".")[0]
			fieldName := strings.Replace(e.Namespace(), structName+".", "", 1)

			fields[fieldName] = e.Translate(m.ut)
		}
		return &fields
	}
	return nil
}
