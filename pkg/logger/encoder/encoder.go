package encoder

import (
	"encoding/json"
	"reflect"
	"strings"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type (
	Entry         = zapcore.Entry
	Field         = zapcore.Field
	EncoderConfig = zapcore.EncoderConfig

	EncoderFunc func(*encoder)
)

type Encoder interface {
	zapcore.Encoder
	Inspector
}

type encoder struct {
	zapcore.Encoder

	disguisedKeys map[string]bool
}

var allowedInspectEncodeEntry = map[zapcore.FieldType]bool{
	zapcore.ArrayMarshalerType: true,
	zapcore.BinaryType:         true,
	zapcore.ByteStringType:     true,
	zapcore.StringType:         true,
	zapcore.ReflectType:        true,
}

func WithMaskSensitiveData(fields []string) EncoderFunc {
	return func(e *encoder) {

		e.disguisedKeys = map[string]bool{}

		for _, field := range fields {
			e.disguisedKeys[strings.ToLower(field)] = true
		}
	}
}

func NewJSONEncoder(ec EncoderConfig, opts ...EncoderFunc) Encoder {
	e := &encoder{Encoder: zapcore.NewJSONEncoder(ec)}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func (e *encoder) Clone() zapcore.Encoder {
	return e
}

func (e *encoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	if len(e.disguisedKeys) == 0 {
		return e.Encoder.EncodeEntry(entry, fields)
	}

	filtered := make([]zapcore.Field, len(fields))

	for i, field := range fields {

		filtered[i] = field

		if field.Key == "" || !allowedInspectEncodeEntry[field.Type] {
			continue
		}

		if field.String != "" {
			filtered[i].String = e.inspectValue(field.Key, field.String)

		} else if field.Interface != nil {

			vo := reflect.Indirect(
				reflect.ValueOf(field.Interface),
			)

			switch vo.Type().Kind() {
			default:
				filtered[i].Interface = e.Inspects(field.Interface)

			case reflect.Slice, reflect.Array:
				if vo.Len() == 0 {
					break
				}
				if vo.Index(0).Type().Kind() == reflect.String {
					filtered[i].Interface = e.inspectSliceStr(field.Key, field.Interface)

				} else if vo.Index(0).Type().Kind() == reflect.Uint8 {
					var tmp interface{}
					_ = json.Unmarshal(field.Interface.([]byte), &tmp)

					filtered[i].Interface, _ = json.Marshal(e.inspectMap(tmp))

				} else {
					filtered[i].Interface = e.Inspects(field.Interface)
				}
			}
		}
	}
	return e.Encoder.EncodeEntry(entry, filtered)
}
