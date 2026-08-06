package pdf_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	gcsMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/gcs"
	. "github.com/paper-indonesia/pivot-backoffice/pkg/pdf"

	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeFilesToPDF(t *testing.T) {
	gcs := gcsMock.NewGCSService(t)

	p := NewPDFGenerator(WithGCSService(gcs))

	cwd, _ := os.Getwd()

	tests := []struct {
		name      string
		files     []MergeFile
		setupMock func()
		wantErr   string
	}{
		{
			name:    "ERROR:File list is empty",
			wantErr: "file list is empty",
		},
		{
			name:    "ERROR:Ext file is not supported",
			files:   []MergeFile{{From: LocalFile, Location: "file.xlsx"}},
			wantErr: "ext .xlsx is not supported",
		},
		{
			name:    "ERROR:Source file is not supported",
			files:   []MergeFile{{From: "google drive", Location: "file.pdf"}},
			wantErr: "source google drive is not supported",
		},
		{
			name: "ERROR:GCS read file from bucket",
			setupMock: func() {
				gcs.On(
					"ReadAll", c.BackgroundCtxMockType(), c.StringMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			files:   []MergeFile{{From: "gcs", Location: "file.pdf"}},
			wantErr: "some error",
		},
		{
			name: "SUCCESS",
			files: []MergeFile{
				{From: LocalFile, Location: cwd + "/samples/image1.PnG"},
				{From: LocalFile, Location: cwd + "/samples/image2.JPG"},
				{From: LocalFile, Location: cwd + "/samples/image3.jpeg"},
				{From: LocalFile, Location: cwd + "/samples/file1.pdf"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			if raw, err := p.MergeFilesToPDF(context.Background(), test.files); test.wantErr == "" {
				require.NoError(t, err)
				assert.NotEmpty(t, raw)
				assert.Greater(t, len(raw), 350000)
				assert.NoError(t, pdfapi.Validate(bytes.NewReader(raw), nil))

			} else {
				require.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}
