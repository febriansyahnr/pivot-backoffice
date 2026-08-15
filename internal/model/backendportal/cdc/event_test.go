package cdcModel_test

import (
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/cdc"

	"github.com/stretchr/testify/assert"
)

// testData is a simple struct for testing generic Event parsing
type testData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    *Event[testData]
		wantErr bool
	}{
		{
			name: "valid create event",
			input: []byte(`{
				"before": null,
				"after": {"id": "1", "name": "test"},
				"source": {"table": "test_table"},
				"op": "c",
				"ts_ms": 1234567890
			}`),
			want: &Event[testData]{
				Before: nil,
				After:  &testData{ID: "1", Name: "test"},
				Source: Source{Table: "test_table"},
				Op:     OpCreate,
				TsMs:   1234567890,
			},
			wantErr: false,
		},
		{
			name: "valid update event",
			input: []byte(`{
				"before": {"id": "1", "name": "old"},
				"after": {"id": "1", "name": "new"},
				"source": {"table": "test_table"},
				"op": "u",
				"ts_ms": 1234567890
			}`),
			want: &Event[testData]{
				Before: &testData{ID: "1", Name: "old"},
				After:  &testData{ID: "1", Name: "new"},
				Source: Source{Table: "test_table"},
				Op:     OpUpdate,
				TsMs:   1234567890,
			},
			wantErr: false,
		},
		{
			name: "valid delete event",
			input: []byte(`{
				"before": {"id": "1", "name": "deleted"},
				"after": null,
				"source": {"table": "test_table"},
				"op": "d",
				"ts_ms": 1234567890
			}`),
			want: &Event[testData]{
				Before: &testData{ID: "1", Name: "deleted"},
				After:  nil,
				Source: Source{Table: "test_table"},
				Op:     OpDelete,
				TsMs:   1234567890,
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   []byte(`{invalid json}`),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEvent[testData](tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEventIsEmpty(t *testing.T) {
	tests := []struct {
		name  string
		event *Event[testData]
		want  bool
	}{
		{
			name:  "both nil",
			event: &Event[testData]{Before: nil, After: nil},
			want:  true,
		},
		{
			name:  "before not nil",
			event: &Event[testData]{Before: &testData{ID: "1"}, After: nil},
			want:  false,
		},
		{
			name:  "after not nil",
			event: &Event[testData]{Before: nil, After: &testData{ID: "1"}},
			want:  false,
		},
		{
			name:  "both not nil",
			event: &Event[testData]{Before: &testData{ID: "1"}, After: &testData{ID: "1"}},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.event.IsEmpty())
		})
	}
}

func TestEventIsCreate(t *testing.T) {
	tests := []struct {
		name  string
		event *Event[testData]
		want  bool
	}{
		{
			name:  "create operation",
			event: &Event[testData]{Op: OpCreate},
			want:  true,
		},
		{
			name:  "update operation",
			event: &Event[testData]{Op: OpUpdate},
			want:  false,
		},
		{
			name:  "delete operation",
			event: &Event[testData]{Op: OpDelete},
			want:  false,
		},
		{
			name:  "unknown operation",
			event: &Event[testData]{Op: "x"},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.event.IsCreate())
		})
	}
}

func TestEventIsUpdate(t *testing.T) {
	tests := []struct {
		name  string
		event *Event[testData]
		want  bool
	}{
		{
			name:  "update operation",
			event: &Event[testData]{Op: OpUpdate},
			want:  true,
		},
		{
			name:  "create operation",
			event: &Event[testData]{Op: OpCreate},
			want:  false,
		},
		{
			name:  "delete operation",
			event: &Event[testData]{Op: OpDelete},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.event.IsUpdate())
		})
	}
}

func TestEventIsDelete(t *testing.T) {
	tests := []struct {
		name  string
		event *Event[testData]
		want  bool
	}{
		{
			name:  "delete operation",
			event: &Event[testData]{Op: OpDelete},
			want:  true,
		},
		{
			name:  "create operation",
			event: &Event[testData]{Op: OpCreate},
			want:  false,
		},
		{
			name:  "update operation",
			event: &Event[testData]{Op: OpUpdate},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.event.IsDelete())
		})
	}
}

func TestEventGetCurrent(t *testing.T) {
	before := &testData{ID: "1", Name: "before"}
	after := &testData{ID: "1", Name: "after"}

	tests := []struct {
		name  string
		event *Event[testData]
		want  *testData
	}{
		{
			name:  "create returns after",
			event: &Event[testData]{Before: nil, After: after, Op: OpCreate},
			want:  after,
		},
		{
			name:  "update returns after",
			event: &Event[testData]{Before: before, After: after, Op: OpUpdate},
			want:  after,
		},
		{
			name:  "delete returns before",
			event: &Event[testData]{Before: before, After: nil, Op: OpDelete},
			want:  before,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.event.GetCurrent())
		})
	}
}

func TestEventGetPrevious(t *testing.T) {
	before := &testData{ID: "1", Name: "before"}
	after := &testData{ID: "1", Name: "after"}

	tests := []struct {
		name  string
		event *Event[testData]
		want  *testData
	}{
		{
			name:  "create returns nil",
			event: &Event[testData]{Before: nil, After: after, Op: OpCreate},
			want:  nil,
		},
		{
			name:  "update returns before",
			event: &Event[testData]{Before: before, After: after, Op: OpUpdate},
			want:  before,
		},
		{
			name:  "delete returns before",
			event: &Event[testData]{Before: before, After: nil, Op: OpDelete},
			want:  before,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.event.GetPrevious())
		})
	}
}
