package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskedCreditCardNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Short", "123456789", "123456789"},
		{"Long", "1234567890123456", "123456******3456"},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := MaskCreditCardNumber(tt.input)
			if actual != tt.expected {
				t.Errorf("MaskCreditCardNumber(%s): expected %s, actual %s", tt.input, tt.expected, actual)
			}
		})
	}
}

func TestMaskFullName(t *testing.T) {
	type args struct {
		fullName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Single word", args: args{"Michael"}, want: "M*****l"},
		{name: "Two words", args: args{"Michael Smith"}, want: "M*****l S***h"},
		{name: "Multiple words", args: args{"Christopher John Anderson"}, want: "C*********r J**n A******n"},
		{name: "Single character", args: args{"M"}, want: "*"},
		{name: "Empty string", args: args{""}, want: ""},
		{name: "Name with multiple spaces", args: args{"  Emily   Davis  "}, want: "E***y D***s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, MaskFullName(tt.args.fullName), "MaskFullName(%v)", tt.args.fullName)
		})
	}
}
