package file_test

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/pkg/file"
	"github.com/tealeg/xlsx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMultipartRequest(fieldName, content string) *http.Request {
	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)

	ff, _ := mw.CreateFormFile(fieldName, "test.txt")
	io.WriteString(ff, content)
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/", buf)
	req.Header.Set(constant.HeaderContentType, mw.FormDataContentType())
	return req
}

func TestCopyTempMultipartFile(t *testing.T) {
	tests := []struct {
		name      string
		dir       string
		pattern   string
		wantError bool
	}{
		{"SUCCESS: Valid params", os.TempDir(), "test-*.txt", false},
		{"ERROR: Invalid pattern", os.TempDir(), "/invalid", true},
		{"ERROR: Invalid dir", "/nonexistent", "test-*.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMultipartRequest("file", "test content")
			ff, _, _ := req.FormFile("file")

			f, err := CopyTempMultipartFile(tt.dir, tt.pattern, ff)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, f)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, f)
				f.Close()
			}
		})
	}
}

func TestFile_Methods(t *testing.T) {
	req := createMultipartRequest("file", "test content")
	ff, _, _ := req.FormFile("file")

	f, err := CopyTempMultipartFile(os.TempDir(), "test-*.txt", ff)
	require.NoError(t, err)
	require.NotNil(t, f)

	// Test Name method
	assert.Contains(t, f.Name(), "test-")

	// Test ReadExcel without reader
	sheet, err := f.ReadExcel()
	assert.Error(t, err)
	assert.Nil(t, sheet)

	// Test Close method
	err = f.Close()
	assert.NoError(t, err)
}

func TestWithReadExcel(t *testing.T) {
	tests := []struct {
		name        string
		readerFunc  func(*os.File) (*xlsx.Sheet, error)
		expectError bool
	}{
		{
			"SUCCESS: Valid reader",
			func(*os.File) (*xlsx.Sheet, error) { return &xlsx.Sheet{}, nil },
			false,
		},
		{
			"ERROR: Reader error",
			func(*os.File) (*xlsx.Sheet, error) { return nil, errors.New("read error") },
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createMultipartRequest("file", "test")
			ff, _, _ := req.FormFile("file")

			f, err := CopyTempMultipartFile(os.TempDir(), "test-*.xlsx", ff, WithReadExcel(tt.readerFunc))
			require.NoError(t, err)
			defer f.Close()

			sheet, err := f.ReadExcel()
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, sheet)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sheet)
			}
		})
	}
}