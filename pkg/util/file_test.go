package util

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createFormRequest(fieldName, fileName, content string) *http.Request {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	part, _ := writer.CreateFormFile(fieldName, fileName)
	part.Write([]byte(content))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/", buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestValidateFile(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		fileName  string
		content   string
		params    ValidateFileFormParams
		wantError bool
		expectErr error
	}{
		{
			"SUCCESS: Valid file",
			"upload",
			"test.csv",
			"data",
			ValidateFileFormParams{FieldName: "upload", Extension: ".csv", FileSize: 1000},
			false,
			nil,
		},
		{
			"ERROR: Wrong extension",
			"upload",
			"test.txt",
			"data",
			ValidateFileFormParams{FieldName: "upload", Extension: ".csv", FileSize: 1000},
			true,
			constant.ErrInvalidFileExtension,
		},
		{
			"ERROR: File too large",
			"upload",
			"test.csv",
			"very long content data",
			ValidateFileFormParams{FieldName: "upload", Extension: ".csv", FileSize: 5},
			true,
			constant.ErrFileTooLarge,
		},
		{
			"ERROR: Field not found",
			"upload",
			"test.csv",
			"data",
			ValidateFileFormParams{FieldName: "missing", Extension: ".csv", FileSize: 1000},
			true,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createFormRequest(tt.fieldName, tt.fileName, tt.content)

			file, header, err := ValidateFile(req, tt.params)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, file)
				assert.Nil(t, header)
				if tt.expectErr != nil {
					assert.Equal(t, tt.expectErr, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, file)
				assert.NotNil(t, header)
				assert.Equal(t, tt.fileName, header.Filename)
			}
		})
	}
}

func TestReadCsvFile(t *testing.T) {
	tests := []struct {
		name        string
		csvContent  string
		params      ReadCsvFileParams
		expected    [][]string
		expectError bool
	}{
		{
			"SUCCESS: Basic CSV",
			"name,age\nJohn,25\nJane,30",
			ReadCsvFileParams{Delimiter: ','},
			[][]string{{"name", "age"}, {"John", "25"}, {"Jane", "30"}},
			false,
		},
		{
			"SUCCESS: Ignore first row",
			"name,age\nJohn,25\nJane,30",
			ReadCsvFileParams{IgnoreFirstRow: true, Delimiter: ','},
			[][]string{{"John", "25"}, {"Jane", "30"}},
			false,
		},
		{
			"SUCCESS: Ignore first row with empty file",
			"name,age",
			ReadCsvFileParams{IgnoreFirstRow: true, Delimiter: ','},
			[][]string{},
			false,
		},
		{
			"SUCCESS: Semicolon delimiter",
			"name;age\nJohn;25",
			ReadCsvFileParams{Delimiter: ';'},
			[][]string{{"name", "age"}, {"John", "25"}},
			false,
		},
		{
			"SUCCESS: Trim leading space",
			"name, age\n John, 25",
			ReadCsvFileParams{Delimiter: ',', TrimLeadingSpace: true},
			[][]string{{"name", "age"}, {"John", "25"}},
			false,
		},
		{
			"SUCCESS: Empty file",
			"",
			ReadCsvFileParams{Delimiter: ','},
			nil,
			false,
		},
		{
			"ERROR: Invalid CSV format",
			"name,age\nJohn,25,extra\nJane",
			ReadCsvFileParams{Delimiter: ','},
			nil,
			true,
		},
		{
			"ERROR: Malformed quotes",
			`name,age
"John,25
Jane,30`,
			ReadCsvFileParams{Delimiter: ',', LazyQuotes: false},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			writer := multipart.NewWriter(buf)

			part, _ := writer.CreateFormFile("file", "test.csv")
			part.Write([]byte(tt.csvContent))
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/", buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			file, _, err := req.FormFile("file")
			require.NoError(t, err)

			records, err := ReadCsvFile(file, tt.params)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, records)
			}
		})
	}
}
