package pdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	pdfmodel "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

var supportMergeExts = []string{
	".jpeg", ".png", ".jpg", ".pdf",
}

var buffpool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

const (
	GCSFile   = "gcs"
	LocalFile = "file"
)

type GCSReader interface {
	ReadAll(ctx context.Context, bucket, object string) ([]byte, error)
}

type PDFGenerator interface {
	MergeFilesToPDF(ctx context.Context, files []MergeFile) ([]byte, error)
}

type MergeFile struct {
	From     string
	Bucket   string
	Location string
}

type PDFOption func(*pdf)

type pdf struct {
	gcs GCSReader

	config       *pdfmodel.Configuration
	importConfig *pdfcpu.Import
}

func WithGCSService(gcs GCSReader) PDFOption {
	return func(p *pdf) {
		p.gcs = gcs
	}
}

func NewPDFGenerator(opts ...PDFOption) *pdf {
	p := &pdf{
		config:       pdfmodel.NewDefaultConfiguration(),
		importConfig: pdfcpu.DefaultImportConfig(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *pdf) MergeFilesToPDF(ctx context.Context, files []MergeFile) ([]byte, error) {
	if len(files) == 0 {
		return nil, errors.New("file list is empty")
	}

	fw := buffpool.Get().(*bytes.Buffer)

	defer buffpool.Put(fw)
	defer fw.Reset()

	var rs []io.ReadSeeker

	for _, file := range files {
		if !slices.Contains(supportMergeExts, strings.ToLower(filepath.Ext(file.Location))) {
			return nil, fmt.Errorf("ext %s is not supported", filepath.Ext(file.Location))
		}

		var raw []byte
		var err error

		switch file.From {
		default:
			return nil, fmt.Errorf("source %s is not supported", file.From)

		case LocalFile:
			raw, err = os.ReadFile(file.Location)

		case GCSFile:
			raw, err = p.gcs.ReadAll(ctx, file.Bucket, file.Location)
		}
		if err != nil {
			return nil, err
		}

		if filepath.Ext(file.Location) == ".pdf" {
			rs = append(rs, bytes.NewReader(raw))
			continue
		}

		w := new(bytes.Buffer)
		if err = pdfapi.ImportImages(nil, w, []io.Reader{bytes.NewBuffer(raw)}, p.importConfig, p.config); err != nil {
			return nil, err
		}
		rs = append(rs, bytes.NewReader(w.Bytes()))

		w.Reset()
	}

	if err := pdfapi.MergeRaw(rs, fw, false, p.config); err != nil {
		return nil, err
	}
	return fw.Bytes(), nil
}
