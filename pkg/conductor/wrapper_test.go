package conductor_test

import (
	"errors"
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/pkg/conductor"

	"github.com/conductor-sdk/conductor-go/sdk/model"
	"github.com/stretchr/testify/assert"
)

func TestNewTaskResultWithOutputAndError(t *testing.T) {
	tests := []struct {
		output                    any
		err                       error
		wantOutputData            map[string]any
		wantReasonForIncompletion string
	}{
		{
			output: struct {
				Status string `json:"status"`
			}{
				Status: "SUCCESS",
			},
			err: assert.AnError,
			wantOutputData: map[string]any{
				"status": "SUCCESS",
			},
			wantReasonForIncompletion: assert.AnError.Error(),
		},
		{
			output: map[string]any{
				"id": "8bdc6db8-8f46-4cb8-8dd4-3626e5d3b5ce", // NOSONAR
			},
			err: errors.New("Test error"), // NOSONAR
			wantOutputData: map[string]any{
				"id": "8bdc6db8-8f46-4cb8-8dd4-3626e5d3b5ce", // NOSONAR
			},
			wantReasonForIncompletion: "Test error", // NOSONAR
		},
	}
	for _, test := range tests {
		taskResult := NewTaskResultWithOutputAndError(&model.Task{}, test.output, test.err)

		assert.Equal(t, test.wantOutputData, taskResult.OutputData)
		assert.Equal(t, test.wantReasonForIncompletion, taskResult.ReasonForIncompletion)
	}
}
