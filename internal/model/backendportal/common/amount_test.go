package commonModel_test

import (
	"testing"

	"github.com/shopspring/decimal"

	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/common"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/proto/common"

	"github.com/stretchr/testify/assert"
)

func TestProtoAmount(t *testing.T) {
	tests := []struct {
		input *Amount
		want  *pb.Amount
	}{
		{},
		{
			input: &Amount{
				Currency: "IDR",
				Value:    "12000",
			},
			want: &pb.Amount{
				Currency: "IDR",
				Value:    "12000",
			},
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, test.input.ProtoAmount())
	}
}

func TestToDecimal(t *testing.T) {
	tests := []struct {
		input *Amount
		want  decimal.Decimal
	}{
		{
			input: &Amount{
				Currency: "IDR",
				Value:    "12000",
			},
			want: decimal.NewFromInt(12000),
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, test.input.ToDecimal())
	}
}
